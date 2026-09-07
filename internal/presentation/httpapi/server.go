package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"time"

	"ddd-example/internal/option"
	"ddd-example/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

// Server http服务
type Server struct {
	server *http.Server

	auth *authController
}

// ServerProvider 提供Server实例
func ServerProvider(injector do.Injector) (*Server, error) {
	s := &Server{
		auth: do.MustInvoke[*authController](injector),
	}

	opt := do.MustInvoke[*option.Options](injector)
	router := s.newRouter()
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", opt.HTTP.Port),
		Handler: router,
	}

	go func() {
		logger.Info(context.Background(), "start server", "listen", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "start server", "error", err)
		}
	}()

	return s, nil
}

// Close 关闭服务
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) newRouter() chi.Router {
	router := chi.NewRouter()

	router.Use(recoverer)

	ac := s.auth
	router.Use(ac.Authorize)

	router.Post(`/session`, ac.LoginWithEmail())
	router.Post(`/register`, ac.Register())
	router.Delete(`/session`, ac.Logout())
	router.Get(`/login/oauth/{site}`, ac.LoginWithOauth())
	router.Post(`/login/oauth/{site}`, ac.VerifyOauth())
	router.Post(`/register/oauth`, ac.RegisterWithOauth())

	router.Group(func(router chi.Router) {
		router.Use(ac.DenyAnonymous)

		router.Get(`/session`, ac.MyIdentity())
		router.Put(`/my/password`, ac.ChangePassword())
	})

	return router
}

// NewHandler 把appHandler转换为http.Handler
//
// render函数用于将app handler返回的数据转换为http response数据。
//
// render函数的前两个入参类型为http.ResponseWriter和*http.Request，
// 后续入参类型就与appHandler的出参参数类型一致，但最后一个error参数可选，
// 如果声明了error入参，appHandler执行完毕后的error会传递给render函数，以实现自定义错误处理。
// render函数的出参可以是(any, error)和(error)两种，如果返回了数据，就会被下发到客户端，否则只会响应空结果。
//
// Example:
//
//	NewHandler(
//	  func(context.Context, args) error,
//	  // render函数的前两个参数必须是http.ResponseWriter和*http.Request
//	  func(http.ResponseWriter, *http.Request) error,
//	)
//
//	NewHandler(
//	  func(context.Context, args) error,
//	  func(http.ResponseWriter, *http.Request, error) error,
//	)
//
//	NewHandler(
//	  func(context.Context, args) (int, string, error),
//	  // 如果app handler返回了数据，render的入参也需要有相同的参数
//	  func(http.ResponseWriter, *http.Request, int, string) (any, error),
//	)
//
//	NewHandler(
//	  func(context.Context, args) (int, string, error),
//	  // render也可以选择接收app handler返回的错误，用于自定义错误处理
//	  func(http.ResponseWriter, *http.Request, int, string, error) (any, error),
//	)
//
//	NewHandler(
//	  func(context.Context, args) (int, string, error),
//	  // render可以不返回任何数据，这样服务器端会响应空消息
//	  func(http.ResponseWriter, *http.Request, int, string) error,
//	)
func NewHandler(appHandler any, render any) (http.Handler, error) {
	handlerFactor, err := newFuncFactor(appHandler)
	if err != nil {
		return nil, err
	} else if err = checkHandlerParameters(handlerFactor); err != nil {
		return nil, err
	}

	renderFactor, err := newFuncFactor(render)
	if err != nil {
		return nil, err
	} else if err = checkRenderParameters(handlerFactor, renderFactor); err != nil {
		return nil, err
	}

	// 是否需要把handler返回的error传递给render
	passErr := len(renderFactor.ins)-2 == len(handlerFactor.outs) &&
		errorT.AssignableTo(renderFactor.ins[len(renderFactor.ins)-1])

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerIns := []reflect.Value{
			reflect.ValueOf(r.Context()),
		}

		if len(handlerFactor.ins) > 1 {
			argT := handlerFactor.ins[1]
			isPtr := argT.Kind() == reflect.Pointer

			if isPtr {
				argT = argT.Elem()
			}

			handlerArgs := reflect.New(argT)
			if err := scanRequest(handlerArgs.Interface(), r); err != nil {
				sendResponse(w, withError(errBadRequest.WrapError(err)))
				return
			}

			if !isPtr {
				handlerArgs = handlerArgs.Elem()
			}

			handlerIns = append(handlerIns, handlerArgs)
		}

		handlerOuts := handlerFactor.v.Call(handlerIns)
		if !passErr {
			if handlerErr, _ := handlerOuts[len(handlerOuts)-1].Interface().(error); handlerErr != nil {
				sendResponse(w, withError(errUnexpectedException.WrapError(handlerErr)))
				return
			}
		}

		renderIns := []reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
		}

		if passErr {
			renderIns = append(renderIns, handlerOuts...)
		} else {
			renderIns = append(renderIns, handlerOuts[:len(handlerOuts)-1]...)
		}

		renderOuts := renderFactor.v.Call(renderIns)

		if renderErr, _ := renderOuts[len(renderOuts)-1].Interface().(error); renderErr != nil {
			if apiErr, ok := renderErr.(apiError); ok {
				sendResponse(w, withError(apiErr))
			} else {
				sendResponse(w, withError(errUnexpectedException.WrapError(renderErr)))
			}
			return
		}

		if len(renderOuts) > 1 {
			sendResponse(w, withData(renderOuts[0].Interface()))
			return
		}
	}), nil
}

// MustNewHandler 把appHandler转换为http.Handler，转换失败则panic
func MustNewHandler(appHandler any, render any) http.Handler {
	handler, err := NewHandler(appHandler, render)
	if err != nil {
		panic(err)
	}
	return handler
}

// NewVoidHandler 对不返回数据的app handler进行默认转换
func NewVoidHandler[T any](appHandler func(context.Context, T) error) (http.Handler, error) {
	return NewHandler(appHandler, func(http.ResponseWriter, *http.Request) error {
		return nil
	})
}

// MustNewVoidHandler 对不返回数据的app handler进行默认转换，转换失败则panic
func MustNewVoidHandler[T any](appHandler func(context.Context, T) error) http.Handler {
	handler, err := NewVoidHandler(appHandler)
	if err != nil {
		panic(err)
	}
	return handler
}

type funcFactor struct {
	v    reflect.Value
	ins  []reflect.Type
	outs []reflect.Type
}

func newFuncFactor(fn any) (*funcFactor, error) {
	v := reflect.ValueOf(fn)
	t := v.Type()

	if t.Kind() != reflect.Func {
		return nil, errors.New("not a function")
	}

	factor := &funcFactor{
		v:    v,
		ins:  slices.Collect(t.Ins()),
		outs: slices.Collect(t.Outs()),
	}

	return factor, nil
}

var (
	contextT            = reflect.TypeFor[context.Context]()
	errorT              = reflect.TypeFor[error]()
	httpResponseWriterT = reflect.TypeFor[http.ResponseWriter]()
	httpRequestT        = reflect.TypeFor[*http.Request]()
)

// 检查handler参数
//   - 入参数量为1至2个
//   - 第一个入参必须是context.Context
//   - 如果有第二个入参，必须是struct或struct指针类型
//   - 出参数量为1至2个
//   - 最后一个出参必须是error类型
func checkHandlerParameters(handler *funcFactor) error {
	if len(handler.ins) < 1 || len(handler.ins) > 2 {
		return errors.New("handler should accept 1 or 2 inputs")
	} else if len(handler.outs) < 1 {
		return errors.New("handler should have at least 1 output")
	}

	if !handler.ins[0].AssignableTo(contextT) {
		return errors.New("first handler input should be context.Context")
	} else if len(handler.ins) == 2 {
		argT := handler.ins[1]
		if argT.Kind() == reflect.Pointer {
			argT = argT.Elem()
		}
		if argT.Kind() != reflect.Struct {
			return errors.New("second handler input should be a struct or struct pointer")
		}
	}

	if !handler.outs[len(handler.outs)-1].AssignableTo(errorT) {
		return errors.New("last handler output should be error")
	}

	return nil
}

// 检查render参数
//   - 第一个入参必须是http.ResponseWriter
//   - 第二个入参必须是*http.Request
//   - 后续入参必须与handler出参类型一致，但error类型可选
//   - 出参数量为1至2个
//   - 最后一个出参必须是error类型
func checkRenderParameters(handler, render *funcFactor) error {
	if len(render.ins) < 2 {
		return errors.New("render should accept at least 2 inputs")
	} else if len(render.outs) < 1 {
		return errors.New("render should return at least 1 output")
	}

	if !render.ins[0].AssignableTo(httpResponseWriterT) {
		return errors.New("first render input should be http.ResponseWriter")
	}

	if !render.ins[1].AssignableTo(httpRequestT) {
		return errors.New("second render input should be *http.Request")
	}

	if renderIns, handlerOuts := len(render.ins), len(handler.outs); renderIns-2 < handlerOuts-1 || renderIns-2 > handlerOuts {
		return errors.New("render inputs do not match handler outputs")
	}

	for i, v := range render.ins[2:] {
		if !handler.outs[i].AssignableTo(v) {
			return fmt.Errorf("render input %d type %s does not match handler output %s", i+3, v, handler.outs[i])
		}
	}

	if len(render.ins)-2 == len(handler.outs) {
		if !errorT.AssignableTo(render.ins[len(render.ins)-1]) {
			return errors.New("last render input should be error")
		}
	}

	if !render.outs[len(render.outs)-1].AssignableTo(errorT) {
		return errors.New("last render output should be error")
	}

	return nil
}

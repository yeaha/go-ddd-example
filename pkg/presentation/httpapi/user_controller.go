package httpapi

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"time"

	"ddd-example/pkg/option"
	"ddd-example/pkg/user/app"
	"ddd-example/pkg/user/app/handler"
	"ddd-example/pkg/user/domain"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

var (
	visitorKey contextKey = "__VISITOR__"
)

type contextKey any

// userController 账号相关接口
type userController struct {
	App *app.Application
	opt *option.Options
}

func newUserController(opt *option.Options) *userController {
	return &userController{
		App: app.NewApplication(opt),
		opt: opt,
	}
}

// Authorize 获取访问者账号中间件
func (c *userController) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if payload, ok := c.readSessionToken(r); ok {
			user, newPayload, err := c.App.Authorize.Handle(r.Context(), payload)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), visitorKey, user))

				if newPayload != "" {
					c.writeSessionToken(newPayload, w)
				}
			} else if !errors.Is(err, domain.ErrSessionTokenExpired) {
				// 只记录错误，不中断请求
				logrus.WithError(err).Error("authorize visitor")
			}
		}

		next.ServeHTTP(w, r)
	})
}

// DenyAnonymous 禁止匿名访问中间件
func (c *userController) DenyAnonymous(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = mustVisitorFromCtx(r.Context())
		next.ServeHTTP(w, r)
	})
}

func (c *userController) writeSessionToken(token string, w http.ResponseWriter) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(token))
	http.SetCookie(w, &http.Cookie{
		Name:     "VISITOR",
		Value:    payload,
		Path:     "/",
		Expires:  time.Now().Add(3 * 31 * 24 * time.Hour),
		HttpOnly: true,
	})
}

func (c *userController) readSessionToken(r *http.Request) (string, bool) {
	if cookie, err := r.Cookie("VISITOR"); err == nil {
		data, err := base64.RawURLEncoding.DecodeString(cookie.Value)
		if err != nil {
			logrus.WithError(err).Debug("base64 decode session token")
			return "", false
		}

		return string(data), len(data) > 0
	}
	return "", false
}

// LoginWithEmail email登录
func (c *userController) LoginWithEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.LoginWithEmail{}
		mustScanJSON(&req, r.Body)

		_, token, err := c.App.LoginWithEmail.Handle(r.Context(), req)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrWrongPassword) {
				panic(errUnauthorized)
			}
			panic(errUnexpectedException.WrapError(err))
		}

		c.writeSessionToken(token, w)
		sendResponse(w, withStatusCode(http.StatusCreated))
	}
}

// Logout 退出登录
func (c *userController) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if user, ok := visitorFromCtx(r.Context()); ok {
			if err := c.App.Logout.Handle(r.Context(), user); err != nil {
				panic(errUnexpectedException.WrapError(err))
			}
		}

		sendResponse(w)
	}
}

// Register 账号注册
func (c *userController) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.Register{}
		mustScanJSON(&req, r.Body)

		_, token, err := c.App.Register.Handle(r.Context(), req)
		if err != nil {
			if errors.Is(err, domain.ErrEmailRegistered) {
				panic(errEmailRegistered)
			}
			panic(errUnexpectedException.WrapError(err))
		}

		c.writeSessionToken(token, w)
		sendResponse(w, withStatusCode(http.StatusCreated))
	}
}

// ChangePassword 修改密码
func (c *userController) ChangePassword() http.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) {
		req := handler.ChangePassword{
			User: mustVisitorFromCtx(r.Context()),
		}
		mustScanJSON(&req, r.Body)

		if err := c.App.ChangePassword.Handle(r.Context(), req); err != nil {
			if errors.Is(err, domain.ErrWrongPassword) {
				panic(errWrongPassword)
			}
			panic(errUnexpectedException.WrapError(err))
		}
	}
}

// MyIdentity 当前访问者信息
func (c *userController) MyIdentity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		visitor := mustVisitorFromCtx(r.Context())

		sendResponse(w, withData(visitor))
	}
}

// LoginWithOauth oauth三方登录，下发重定向地址
func (c *userController) LoginWithOauth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client, ok := c.opt.GetOauthClient(chi.URLParam(r, "site"))
		if !ok {
			panic(errOauthNotSupport)
		}

		req := struct {
			RedirectURI string `json:"redirect_uri" valid:"url,required"` // FIXME: 检查重定向地址域名有效性，防止钓鱼劫持
		}{}
		mustScanValues(&req, r.URL.Query())

		sendResponse(w, withData(mapAny{
			"next_url": client.AuthorizeURL(req.RedirectURI).String(),
		}))
	}
}

// VerifyOauth 验证三方登录
//
// 前端在三方站点验证完毕重定向回来之后，把三方站点回传的query string提交给服务器端做验证
func (c *userController) VerifyOauth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client, ok := c.opt.GetOauthClient(chi.URLParam(r, "site"))
		if !ok {
			panic(errOauthNotSupport)
		}

		req := handler.VerifyOauth{
			Client: client,
		}
		mustScanJSON(&req, r.Body)

		query, err := url.ParseQuery(req.RawQuery)
		if err != nil {
			panic(errBadRequest.WrapError(err))
		} else if code := query.Get("code"); code == "" {
			panic(errBadRequest.WrapError(err))
		}
		req.Query = query

		result, err := c.App.VerifyOauth.Handle(r.Context(), req)
		if err != nil {
			panic(errUnexpectedException.WrapError(err))
		} else if user := result.User; user != nil {
			// 下发会话凭证及登录账号信息
			c.writeSessionToken(result.SessionToken, w)

			sendResponse(w, withData(mapAny{
				"user": user,
			}))
			return
		} else if token := result.OauthToken; token != "" {
			// 下发三方验证token，用于后续注册或关联账号
			sendResponse(w, withData(mapAny{
				"oauth_token": token,
			}))
			return
		}

		// 不应该走到这里
		panic(errUnexpectedException.WrapError(errors.New("oops")))
	}
}

// RegisterWithOauth 三方账号绑定或注册
func (c *userController) RegisterWithOauth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.RegisterWithOauth{}
		mustScanJSON(&req, r.Body)

		user, token, err := c.App.RegisterWithOauth.Handle(r.Context(), req)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidOauthToken) {
				panic(errInvalidOauthToken)
			} else if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrWrongPassword) {
				panic(errUnauthorized)
			} else if errors.Is(err, domain.ErrEmailRegistered) {
				panic(errEmailRegistered)
			}

			panic(errUnexpectedException.WrapError(err))
		}

		c.writeSessionToken(token, w)
		sendResponse(w, withData(mapAny{
			"user": user,
		}))
	}
}

func visitorFromCtx(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(visitorKey).(*domain.User)
	return user, ok
}

func mustVisitorFromCtx(ctx context.Context) *domain.User {
	user, ok := visitorFromCtx(ctx)
	if !ok {
		panic(errUnauthorized)
	}
	return user
}

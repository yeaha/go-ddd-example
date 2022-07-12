package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/joyparty/httpkit"
	"github.com/sirupsen/logrus"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/option"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/handler"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

var (
	visitorKey contextKey = "__VISITOR__"
)

type contextKey any

// userController 账号相关接口
type userController struct {
	App *app.Application
}

func newUserController(opt *option.Options) *userController {
	return &userController{
		App: app.NewApplication(opt),
	}
}

// Authorize 获取访问者账号中间件
func (c *userController) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if payload, ok := c.readSessionToken(r); ok {
			user, token, err := c.App.Handlers.RetrieveToken.Handle(r.Context(), payload)
			if err != nil {
				// 只记录错误，不中断请求
				logrus.WithError(err).Debug("retrieve session token")
			} else if !token.IsExpired() {
				r = r.WithContext(context.WithValue(r.Context(), visitorKey, user))

				if token.NeedRenew() {
					payload, err = c.App.Handlers.RenewToken.Handle(r.Context(), user)
					if err != nil {
						logrus.WithError(err).Error("renew session token")
					} else {
						c.writeSessionToken(payload, w)
					}
				}
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
	http.SetCookie(w, &http.Cookie{
		Name:     "VISITOR",
		Value:    token,
		Expires:  time.Now().Add(3 * 31 * 24 * time.Hour),
		HttpOnly: true,
	})
}

func (c *userController) readSessionToken(r *http.Request) (string, bool) {
	if cookie, err := r.Cookie("VISITOR"); err == nil {
		payload := cookie.Value
		return payload, payload != ""
	}

	return "", false
}

// LoginWithEmail email登录
func (c *userController) LoginWithEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.LoginWithEmail{}
		httpkit.MustScanJSON(&req, r.Body)

		_, token, err := c.App.Handlers.LoginWithEmail.Handle(r.Context(), req)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrWrongPassword) {
				panic(httpkit.NewError(http.StatusUnauthorized))
			}
			panic(httpkit.WrapError(err))
		}
		c.writeSessionToken(token, w)
	}
}

// Logout 退出登录
func (c *userController) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if user, ok := visitorFromCtx(r.Context()); ok {
			if err := c.App.Handlers.Logout.Handle(r.Context(), user); err != nil {
				panic(httpkit.WrapError(err))
			}
		}
	}
}

// Register 账号注册
func (c *userController) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.Register{}
		httpkit.MustScanJSON(&req, r.Body)

		_, token, err := c.App.Handlers.Register.Handle(r.Context(), req)
		if err != nil {
			if errors.Is(err, domain.ErrEmailRegistered) {
				panic(httpkit.NewError(http.StatusConflict).WithJSON(httpkit.M{
					"error": "EMAIL_REGISTERED",
				}))
			}

			panic(httpkit.WrapError(err))
		}
		c.writeSessionToken(token, w)
	}
}

// ChangePassword 修改密码
func (c *userController) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.ChangePassword{
			User: mustVisitorFromCtx(r.Context()),
		}
		httpkit.MustScanJSON(&req, r.Body)

		if err := c.App.Handlers.ChangePassword.Handle(r.Context(), req); err != nil {
			if errors.Is(err, domain.ErrWrongPassword) {
				panic(httpkit.NewError(http.StatusNotAcceptable).WithJSON(httpkit.M{
					"error": "INCORRECT_OLD_PASSWORD",
				}))
			}
			panic(httpkit.WrapError(err))
		}
	}
}

// MyIdentity 当前访问者信息
func (c *userController) MyIdentity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		visitor := mustVisitorFromCtx(r.Context())
		httpkit.Render.JSON(w, http.StatusOK, visitor)
	}
}

func visitorFromCtx(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(visitorKey).(*domain.User)
	return user, ok
}

func mustVisitorFromCtx(ctx context.Context) *domain.User {
	user, ok := visitorFromCtx(ctx)
	if !ok {
		panic(httpkit.NewError(http.StatusUnauthorized).WithJSON(httpkit.M{
			"error": "DENY_ANONYMOUS",
		}))
	}
	return user
}

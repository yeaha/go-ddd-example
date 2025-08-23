package httpapi

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"time"

	"ddd-example/internal/app/handler"
	"ddd-example/internal/domain"
	"ddd-example/internal/option"
	"ddd-example/pkg/logger"

	"github.com/go-chi/chi/v5"
)

var visitorKey contextKey = "__VISITOR__"

type contextKey any

// authController 账号相关接口
type authController struct {
	// revive:disable:struct-tag

	opt *option.Options `do:""`

	authorize         *handler.AuthorizeHandler         `do:""`
	changePassword    *handler.ChangePasswordHandler    `do:""`
	loginWithEmail    *handler.LoginWithEmailHandler    `do:""`
	logout            *handler.LogoutHandler            `do:""`
	register          *handler.RegisterHandler          `do:""`
	registerWithOauth *handler.RegisterWithOauthHandler `do:""`
	verifyOauth       *handler.VerifyOauthHandler       `do:""`

	// revive:enable:struct-tag
}

// Authorize 获取访问者账号中间件
func (c *authController) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if payload, ok := c.readSessionToken(r); ok {
			account, newPayload, err := c.authorize.Handle(r.Context(), payload)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), visitorKey, account))

				if newPayload != "" {
					c.writeSessionToken(newPayload, w)
				}
			} else if !errors.Is(err, domain.ErrSessionTokenExpired) {
				// 只记录错误，不中断请求
				logger.Error(r.Context(), "authorize visitor", "error", err)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// DenyAnonymous 禁止匿名访问中间件
func (c *authController) DenyAnonymous(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = mustVisitorFromCtx(r.Context())
		next.ServeHTTP(w, r)
	})
}

func (c *authController) writeSessionToken(token string, w http.ResponseWriter) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(token))
	http.SetCookie(w, &http.Cookie{
		Name:     "VISITOR",
		Value:    payload,
		Path:     "/",
		Expires:  time.Now().Add(3 * 31 * 24 * time.Hour),
		HttpOnly: true,
	})
}

func (c *authController) readSessionToken(r *http.Request) (string, bool) {
	if cookie, err := r.Cookie("VISITOR"); err == nil {
		data, err := base64.RawURLEncoding.DecodeString(cookie.Value)
		if err != nil {
			// 出错了不中断请求，打印错误日志，作为匿名访问处理
			logger.Error(r.Context(), "base64 decode session token", "error", err)
			return "", false
		}

		return string(data), len(data) > 0
	}
	return "", false
}

// LoginWithEmail email登录
func (c *authController) LoginWithEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.LoginWithEmail{}
		mustScanJSON(&req, r.Body)

		_, token, err := c.loginWithEmail.Handle(r.Context(), req)
		if err != nil {
			if errors.Is(err, domain.ErrAccountNotFound) || errors.Is(err, domain.ErrWrongPassword) {
				panic(errUnauthorized)
			}
			panic(errUnexpectedException.WrapError(err))
		}

		c.writeSessionToken(token, w)
		sendResponse(w, withStatusCode(http.StatusCreated))
	}
}

// Logout 退出登录
func (c *authController) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if account, ok := visitorFromCtx(r.Context()); ok {
			if err := c.logout.Handle(r.Context(), account); err != nil {
				panic(errUnexpectedException.WrapError(err))
			}
		}

		sendResponse(w)
	}
}

// Register 账号注册
func (c *authController) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.Register{}
		mustScanJSON(&req, r.Body)

		_, token, err := c.register.Handle(r.Context(), req)
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
func (c *authController) ChangePassword() http.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) {
		req := handler.ChangePassword{
			Account: mustVisitorFromCtx(r.Context()),
		}
		mustScanJSON(&req, r.Body)

		if err := c.changePassword.Handle(r.Context(), req); err != nil {
			if errors.Is(err, domain.ErrWrongPassword) {
				panic(errWrongPassword)
			}
			panic(errUnexpectedException.WrapError(err))
		}
	}
}

// MyIdentity 当前访问者信息
func (c *authController) MyIdentity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		visitor := mustVisitorFromCtx(r.Context())

		sendResponse(w, withData(visitor))
	}
}

// LoginWithOauth oauth三方登录，下发重定向地址
func (c *authController) LoginWithOauth() http.HandlerFunc {
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
func (c *authController) VerifyOauth() http.HandlerFunc {
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

		result, err := c.verifyOauth.Handle(r.Context(), req)
		if err != nil {
			panic(errUnexpectedException.WrapError(err))
		} else if account := result.Account; account != nil {
			// 下发会话凭证及登录账号信息
			c.writeSessionToken(result.SessionToken, w)

			sendResponse(w, withData(mapAny{
				"account": account,
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
func (c *authController) RegisterWithOauth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := handler.RegisterWithOauth{}
		mustScanJSON(&req, r.Body)

		account, token, err := c.registerWithOauth.Handle(r.Context(), req)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidOauthToken) {
				panic(errInvalidOauthToken)
			} else if errors.Is(err, domain.ErrAccountNotFound) || errors.Is(err, domain.ErrWrongPassword) {
				panic(errUnauthorized)
			} else if errors.Is(err, domain.ErrEmailRegistered) {
				panic(errEmailRegistered)
			}

			panic(errUnexpectedException.WrapError(err))
		}

		c.writeSessionToken(token, w)
		sendResponse(w, withData(mapAny{
			"account": account,
		}))
	}
}

func visitorFromCtx(ctx context.Context) (*domain.Account, bool) {
	account, ok := ctx.Value(visitorKey).(*domain.Account)
	return account, ok
}

func mustVisitorFromCtx(ctx context.Context) *domain.Account {
	account, ok := visitorFromCtx(ctx)
	if !ok {
		panic(errUnauthorized)
	}
	return account
}

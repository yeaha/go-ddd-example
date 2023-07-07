package handler

import (
	"context"
	"fmt"

	"ddd-example/pkg/user/app/adapter"
	"ddd-example/pkg/user/app/event"
	"ddd-example/pkg/user/app/service"
	"ddd-example/pkg/user/domain"
	"ddd-example/pkg/user/infra"
	"ddd-example/pkg/utils/oauth"

	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	uuid "github.com/satori/go.uuid"
)

// RegisterWithOauth 三方账号注册，参数
type RegisterWithOauth struct {
	OauthToken     string `json:"oauth_token" valid:",required"`
	Email          string `json:"email" valid:"email,required"`
	VerifyPassword string `json:"verify_password" valid:",optional"` // 绑定账号，需要提供密码
}

// RegisterWithOauthHandler 三方账号注册
type RegisterWithOauthHandler struct {
	DB         *sqlx.DB
	Session    *service.SessionTokenService
	OauthToken *service.OauthTokenService
}

// Handle 三方登录，绑定或注册新账号
func (h *RegisterWithOauthHandler) Handle(ctx context.Context, args RegisterWithOauth) (user *domain.User, sessionToken string, err error) {
	vendorUser, err := h.OauthToken.Retrieve(ctx, args.OauthToken)
	if err != nil {
		err = fmt.Errorf("retrieve vendor user from cache, %w", err)
		return
	}

	var events []any
	if err = entity.Transaction(h.DB, func(tx *sqlx.Tx) error {
		user, events, err = h.handle(
			ctx, args, vendorUser,
			service.NewUserService(tx),
			infra.NewOauthDBRepository(tx),
		)
		return err
	}); err != nil {
		return
	}

	event.Publish(events...)

	sessionToken, err = h.Session.Generate(ctx, user)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
	}
	event.Publish(event.Login{
		User: user,
	})
	return
}

func (h *RegisterWithOauthHandler) handle(
	ctx context.Context,
	args RegisterWithOauth,
	vendorUser *oauth.User,
	userService *service.UserService,
	oauthRepos adapter.OauthRepository,
) (user *domain.User, events []any, err error) {
	if args.VerifyPassword != "" {
		user, err = userService.Authorize(ctx, args.Email, args.VerifyPassword)
	} else {
		user, err = userService.Create(ctx, args.Email, uuid.NewV4().String())

		if err == nil {
			events = append(events, event.Register{
				User: user,
			})
		}
	}

	if err != nil {
		return
	}

	if err = oauthRepos.Bind(ctx, user.ID, vendorUser.Vendor, vendorUser.ID); err != nil {
		err = fmt.Errorf("bound vendor user, %w", err)
	}
	return
}

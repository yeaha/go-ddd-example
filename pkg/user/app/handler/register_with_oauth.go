package handler

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	uuid "github.com/satori/go.uuid"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/service"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/utils/oauth"
)

// RegisterWithOauth 三方账号注册，参数
type RegisterWithOauth struct {
	VendorToken    string `json:"vendor_token" valid:",required"`
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
	vendorUser, err := h.OauthToken.Retrieve(ctx, args.VendorToken)
	if err != nil {
		if errors.Is(err, domain.ErrMissingCache) {
			err = domain.ErrInvalidVendorToken
			return
		}
		err = fmt.Errorf("retrieve vendor user from cache, %w", err)
		return
	}

	if err = entity.Transaction(h.DB, func(tx *sqlx.Tx) error {
		user, err = h.handle(
			ctx, args, vendorUser,
			service.NewUserService(tx),
			service.NewOauthService(tx),
		)
		return err
	}); err != nil {
		return
	}

	sessionToken, err = h.Session.Generate(ctx, user)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
	}
	return
}

func (h *RegisterWithOauthHandler) handle(
	ctx context.Context,
	args RegisterWithOauth,
	vendorUser *oauth.User,
	userService *service.UserService,
	oauthService *service.OauthService,
) (user *domain.User, err error) {
	if args.VerifyPassword != "" {
		user, err = userService.Authorize(ctx, args.Email, args.VerifyPassword)
	} else {
		user, err = userService.Create(ctx, args.Email, uuid.NewV4().String())
	}

	if err != nil {
		return
	}

	if err = oauthService.Bind(ctx, user, vendorUser); err != nil {
		err = fmt.Errorf("bound vendor user, %w", err)
	}
	return
}

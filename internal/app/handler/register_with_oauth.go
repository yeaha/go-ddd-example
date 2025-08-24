package handler

import (
	"context"
	"fmt"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/app/event"
	"ddd-example/internal/app/service"
	"ddd-example/internal/domain"
	"ddd-example/internal/infra"
	"ddd-example/pkg/oauth"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
)

// RegisterWithOauth 三方账号注册，参数
type RegisterWithOauth struct {
	OauthToken     string `json:"oauth_token" validte:"required"`
	Email          string `json:"email" validate:"email"`
	VerifyPassword string `json:"verify_password"` // 绑定账号，需要提供密码
}

// RegisterWithOauthHandler 三方账号注册
type RegisterWithOauthHandler struct {
	DB         *sqlx.DB                     `do:""`
	Session    *service.SessionTokenService `do:""`
	OauthToken *service.OauthTokenService   `do:""`
}

// Handle 三方登录，绑定或注册新账号
func (h *RegisterWithOauthHandler) Handle(ctx context.Context, args RegisterWithOauth) (account *domain.Account, sessionToken string, err error) {
	vendorUser, err := h.OauthToken.Retrieve(ctx, args.OauthToken)
	if err != nil {
		err = fmt.Errorf("retrieve vendor account from cache, %w", err)
		return
	}

	var events []any
	if err = entity.Transaction(h.DB, func(tx *sqlx.Tx) error {
		account, events, err = h.handle(
			ctx, args, vendorUser,
			service.NewAccountService(tx),
			infra.NewOauthRepositoryTx(tx),
		)
		return err
	}); err != nil {
		return
	}

	event.Publish(events...)

	sessionToken, err = h.Session.Generate(ctx, account)
	if err != nil {
		err = fmt.Errorf("generate session token, %w", err)
	}
	event.Publish(event.Login{
		Account: account,
	})
	return
}

func (h *RegisterWithOauthHandler) handle(
	ctx context.Context,
	args RegisterWithOauth,
	vendorUser *oauth.User,
	accountService *service.AccountService,
	oauthRepos adapter.OauthRepository,
) (account *domain.Account, events []any, err error) {
	if args.VerifyPassword != "" {
		account, err = accountService.Authorize(ctx, args.Email, args.VerifyPassword)
	} else {
		account, err = accountService.Create(ctx, args.Email, uuid.New().String())

		if err == nil {
			events = append(events, event.Register{
				Account: account,
			})
		}
	}

	if err != nil {
		return
	}

	if err = oauthRepos.Bind(ctx, account.ID, vendorUser.Vendor, vendorUser.ID); err != nil {
		err = fmt.Errorf("bound vendor user, %w", err)
	}
	return
}

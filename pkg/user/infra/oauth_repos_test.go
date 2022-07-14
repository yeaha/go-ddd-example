//go:build dbtest
// +build dbtest

package infra

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

func TestOauthDBRepository(t *testing.T) {
	if err := entity.Transaction(testDB, func(tx *sqlx.Tx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var (
			repos = NewOauthDBRepository(tx)

			userID    = uuid.NewV4()
			vendor    = "facebook"
			venderUID = uuid.NewV4().String()
		)

		if err := (testCases{
			{
				Name: "Bind()",
				Fn: func() error {
					return repos.Bind(ctx, userID, vendor, venderUID)
				},
			},
			{
				Name: "Find()",
				Fn: func() error {
					_, err := repos.Find(ctx, "google", uuid.NewV4().String())
					require.ErrorIs(t, err, domain.ErrUserNotFound)

					uid, err := repos.Find(ctx, vendor, venderUID)
					require.True(t, uuid.Equal(userID, uid))
					return err
				},
			},
		}).Execute(); err != nil {
			return err
		}

		return errAutoRollback
	}); !errors.Is(err, errAutoRollback) {
		t.Fatalf("oauth repository, %v", err)
	}
}

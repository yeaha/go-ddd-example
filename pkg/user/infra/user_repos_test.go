//go:build dbtest
// +build dbtest

package infra

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"
	"github.com/stretchr/testify/require"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/utils/database"

	// postgresql database driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	errAutoRollback = errors.New("auto rollback")
	testDB          *sqlx.DB
)

func init() {
	dsn := os.Getenv("TESTDB")
	if dsn == "" {
		panic(fmt.Errorf("require ENV %q", "TESTDB"))
	}

	db, err := database.NewDB(database.Option{
		Driver: "pgx",
		DSN:    dsn,
	})
	if err != nil {
		panic(fmt.Errorf("connect test database, %w", err))
	}
	testDB = db.Unsafe()

	entity.DefaultCacher = cache.NewMemoryCache()
}

type (
	testCase struct {
		Name string
		Fn   func() error
	}

	testCases []testCase
)

func (tc testCases) Execute() error {
	for _, v := range tc {
		if err := v.Fn(); err != nil {
			return fmt.Errorf("%s, %w", v.Name, err)
		}
	}
	return nil
}

func TestUserDBRepository(t *testing.T) {
	if err := entity.Transaction(testDB, func(tx *sqlx.Tx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var (
			user  *domain.User
			repos = NewUserDBRepository(tx)
		)

		if err := (testCases{
			{
				Name: "Create()",
				Fn: func() error {
					user = &domain.User{}
					user.SetEmail("yangyi@juwang.cn")
					require.NoError(t, user.SetPassword("abcdef"))

					return repos.Create(ctx, user)
				},
			},
			{
				Name: "FindByEmail",
				Fn: func() error {
					_, err := repos.FindByEmail(ctx, "yangyi@qq.com")
					require.ErrorIs(t, err, domain.ErrUserNotFound)

					_, err = repos.FindByEmail(ctx, "yangyi@juwang.cn")
					return err
				},
			},
			{
				Name: "Save()",
				Fn: func() error {
					require.NoError(t, user.RefreshSessionSalt())
					return repos.Save(ctx, user)
				},
			},
		}).Execute(); err != nil {
			return err
		}

		return errAutoRollback
	}); !errors.Is(err, errAutoRollback) {
		t.Fatalf("users repository, %v", err)
	}
}

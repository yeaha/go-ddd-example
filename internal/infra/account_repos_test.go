//go:build dbtest
// +build dbtest

package infra

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ddd-example/internal/domain"

	"github.com/joyparty/entity"
)

func TestAccountRepository(t *testing.T) {
	if err := entity.Transaction(testDB, func(tx entity.DB) (err error) {
		defer func() {
			err = cmp.Or(err, errRollbackTest)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		email := "test@test.com"
		repos := newAccountDBRepository(tx)

		return testTable{
			{
				Name: "Create",
				Func: func() error {
					account := &domain.Account{}
					if err := account.SetEmail(email); err != nil {
						return fmt.Errorf("set email, %w", err)
					}

					if err := account.SetPassword("abcdef"); err != nil {
						return fmt.Errorf("set password, %w", err)
					}

					return repos.Create(ctx, account)
				},
			},
			{
				Name: "FindByEmail",
				Func: func() error {
					if _, err := repos.FindByEmail(ctx, "test@test.net"); err == nil {
						return errors.New("expected error for non-existent account")
					} else if !errors.Is(err, domain.ErrAccountNotFound) {
						return fmt.Errorf("expected domain.ErrAccountNotFound, got %v", err)
					}

					_, err := repos.FindByEmail(ctx, email)
					return err
				},
			},
		}.Execute()
	}); !errors.Is(err, errRollbackTest) {
		t.Fatalf("account repository, %v", err)
	}
}

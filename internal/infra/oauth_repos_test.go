//go:build dbtest
// +build dbtest

package infra

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ddd-example/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
)

func TestOauthDBRepository(t *testing.T) {
	if err := entity.Transaction(testDB, func(tx *sqlx.Tx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var (
			accountID = uuid.New()
			vendor    = "joyparty"
			vendorUID = uuid.New().String()
		)

		repos := NewOauthDBRepository(tx)

		table := testTable{
			{
				Name: "Bind",
				Func: func() error {
					return repos.Bind(ctx, accountID, vendor, vendorUID)
				},
			},
			{
				Name: "Find",
				Func: func() error {
					if _, err := repos.Find(ctx, "foobar", uuid.New().String()); err == nil {
						return errors.New("expected error for non-existent binding")
					} else if !errors.Is(err, domain.ErrAccountNotFound) {
						return fmt.Errorf("expected domain.ErrAccountNotFound, got %v", err)
					}

					if uid, err := repos.Find(ctx, vendor, vendorUID); err != nil {
						return err
					} else if uid != accountID {
						return fmt.Errorf("expected %s, got %s", accountID, uid)
					}

					return nil
				},
			},
		}

		if err := table.Execute(); err != nil {
			return err
		}

		return errRollbackTest
	}); !errors.Is(err, errRollbackTest) {
		t.Fatalf("oauth repository, %v", err)
	}
}

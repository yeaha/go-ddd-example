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

	"github.com/google/uuid"
	"github.com/joyparty/entity"
)

func TestOauthDBRepository(t *testing.T) {
	if err := entity.Transaction(testDB, func(db entity.DB) (err error) {
		defer func() {
			err = cmp.Or(err, errRollbackTest)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var (
			accountID = uuid.New()
			vendor    = "joyparty"
			vendorUID = uuid.New().String()
		)

		repos := NewOauthRepository(db)

		return testTable{
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
		}.Execute()
	}); !errors.Is(err, errRollbackTest) {
		t.Fatalf("oauth repository, %v", err)
	}
}

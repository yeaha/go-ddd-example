//go:build dbtest
// +build dbtest

package infra

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"ddd-example/internal/domain"
	"ddd-example/pkg/database"

	"github.com/jmoiron/sqlx"
	"github.com/joyparty/entity"
	"github.com/joyparty/entity/cache"
	"github.com/stretchr/testify/suite"

	// postgresql database driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	testDB *sqlx.DB
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

func TestUserDBRepository(t *testing.T) {
	suite.Run(t, &userRepositoryTestSuite{})
}

type userRepositoryTestSuite struct {
	suite.Suite
	repos *UserDBRepository
	tx    *sqlx.Tx

	ctx struct {
		Email string
	}
}

func (s *userRepositoryTestSuite) SetupSuite() {
	tx, err := testDB.BeginTxx(context.Background(), &sql.TxOptions{})
	s.Require().NoError(err)

	s.tx = tx
	s.repos = NewUserDBRepository(tx)

	s.ctx.Email = "test@test.com"
}

func (s *userRepositoryTestSuite) TearDownSuite() {
	s.Require().NoError(s.tx.Rollback())
}

func (s *userRepositoryTestSuite) Test1_Create() {
	require := s.Require()
	user := &domain.User{}

	require.NoError(user.SetEmail(s.ctx.Email))
	require.NoError(user.SetPassword("abcdef"))
	require.NoError(s.repos.Create(context.Background(), user))
}

func (s *userRepositoryTestSuite) Test2_FindByEmail() {
	require := s.Require()

	_, err := s.repos.FindByEmail(context.Background(), "test@test.net")
	require.ErrorIs(err, domain.ErrUserNotFound)

	_, err = s.repos.FindByEmail(context.Background(), s.ctx.Email)
	require.NoError(err)
}

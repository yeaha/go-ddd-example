//go:build dbtest
// +build dbtest

package infra

import (
	"context"
	"database/sql"
	"testing"

	"ddd-example/internal/domain"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
)

func TestOauthDBRepository(t *testing.T) {
	suite.Run(t, &oauthRepositoryTestSuite{})
}

type oauthRepositoryTestSuite struct {
	suite.Suite
	tx    *sqlx.Tx
	repos *OauthDBRepository

	ctx struct {
		AccountID uuid.UUID
		Vendor    string
		VendorUID string
	}
}

func (s *oauthRepositoryTestSuite) SetupSuite() {
	tx, err := testDB.BeginTxx(context.Background(), &sql.TxOptions{})
	s.Require().NoError(err)

	s.tx = tx
	s.repos = NewOauthDBRepository(tx)

	s.ctx.AccountID = uuid.NewV4()
	s.ctx.Vendor = "facebook"
	s.ctx.VendorUID = uuid.NewV4().String()
}

func (s *oauthRepositoryTestSuite) TearDownSuite() {
	s.Require().NoError(s.tx.Rollback())
}

func (s *oauthRepositoryTestSuite) Test1_Bind() {
	err := s.repos.Bind(context.Background(), s.ctx.AccountID, s.ctx.Vendor, s.ctx.VendorUID)
	s.Require().NoError(err)
}

func (s *oauthRepositoryTestSuite) Test2_Find() {
	require := s.Require()

	_, err := s.repos.Find(context.Background(), "google", uuid.NewV4().String())
	require.ErrorIs(err, domain.ErrAccountNotFound)

	uid, err := s.repos.Find(context.Background(), s.ctx.Vendor, s.ctx.VendorUID)
	require.True(uuid.Equal(s.ctx.AccountID, uid))
	require.NoError(err)
}

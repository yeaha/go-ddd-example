package infra

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/domain"

	"github.com/google/uuid"
)

var _ = Describe("oauthDBRepository", func() {
	var (
		repos     adapter.OauthRepository
		accountID uuid.UUID
		vendor    string
		vendorUID string
	)

	BeforeEach(func() {
		repos = NewOauthRepository(beginTestTx())
		accountID = uuid.New()
		vendor = "joyparty"
		vendorUID = uuid.New().String()
	})

	Describe("Bind", func() {
		It("binds the vendor account to the account id", func(ctx SpecContext) {
			Expect(repos.Bind(ctx, accountID, vendor, vendorUID)).To(Succeed())
		}, specTimeout)
	})

	Describe("Find", func() {
		When("the binding exists", func() {
			BeforeEach(func(ctx SpecContext) {
				Expect(repos.Bind(ctx, accountID, vendor, vendorUID)).To(Succeed())
			})

			It("returns the bound account id", func(ctx SpecContext) {
				got, err := repos.Find(ctx, vendor, vendorUID)
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(accountID))
			}, specTimeout)
		})

		When("no binding matches", func() {
			It("returns domain.ErrAccountNotFound", func(ctx SpecContext) {
				_, err := repos.Find(ctx, "foobar", uuid.New().String())
				Expect(err).To(MatchError(domain.ErrAccountNotFound))
			}, specTimeout)
		})
	})
})

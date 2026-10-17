package infra

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/domain"

	"github.com/google/uuid"
)

var _ = Describe("accountDBRepository", func() {
	const (
		email        = "test@test.com"
		updatedEmail = "updated@test.com"
	)

	var repos adapter.AccountRepository

	BeforeEach(func() {
		repos = newAccountDBRepository(beginTestTx())
	})

	newAccount := func() *domain.Account {
		GinkgoHelper()

		account := &domain.Account{}
		Expect(account.SetEmail(email)).To(Succeed())
		Expect(account.SetPassword("abcdef")).To(Succeed())
		return account
	}

	// createAccount 创建账号并返回带 ID 的领域对象，供需要 ID 的 spec 复用
	createAccount := func(ctx SpecContext) *domain.Account {
		GinkgoHelper()

		Expect(repos.Create(ctx, newAccount())).To(Succeed())

		account, err := repos.FindByEmail(ctx, email)
		Expect(err).NotTo(HaveOccurred())
		Expect(account).NotTo(BeNil())
		return account
	}

	Describe("Create", func() {
		It("persists the account", func(ctx SpecContext) {
			Expect(repos.Create(ctx, newAccount())).To(Succeed())

			account, err := repos.FindByEmail(ctx, email)
			Expect(err).NotTo(HaveOccurred())
			Expect(account).NotTo(BeNil())
			Expect(account.Email).To(Equal(email))
			Expect(account.ComparePassword("abcdef")).To(BeTrue())
		}, specTimeout)
	})

	Describe("Find", func() {
		When("the account exists", func() {
			var account *domain.Account

			BeforeEach(func(ctx SpecContext) {
				account = createAccount(ctx)
			})

			It("returns the account", func(ctx SpecContext) {
				got, err := repos.Find(ctx, account.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(got).NotTo(BeNil())
				Expect(got.ID).To(Equal(account.ID))
				Expect(got.Email).To(Equal(email))

				// 第二次会命中 entity 的内存缓存，结果应当一致
				cached, err := repos.Find(ctx, account.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cached).NotTo(BeNil())
				Expect(cached.ID).To(Equal(account.ID))
				Expect(cached.Email).To(Equal(email))
			}, specTimeout)
		})

		When("no account matches the id", func() {
			It("returns domain.ErrAccountNotFound", func(ctx SpecContext) {
				_, err := repos.Find(ctx, uuid.New())
				Expect(err).To(MatchError(domain.ErrAccountNotFound))
			}, specTimeout)
		})
	})

	Describe("FindByEmail", func() {
		When("the account exists", func() {
			BeforeEach(func(ctx SpecContext) {
				Expect(repos.Create(ctx, newAccount())).To(Succeed())
			})

			It("returns the matching account", func(ctx SpecContext) {
				account, err := repos.FindByEmail(ctx, email)
				Expect(err).NotTo(HaveOccurred())
				Expect(account).NotTo(BeNil())
				Expect(account.Email).To(Equal(email))
			}, specTimeout)
		})

		When("no account matches the email", func() {
			It("returns domain.ErrAccountNotFound", func(ctx SpecContext) {
				_, err := repos.FindByEmail(ctx, "test@test.net")
				Expect(err).To(MatchError(domain.ErrAccountNotFound))
			}, specTimeout)
		})
	})

	Describe("Update", func() {
		When("the account exists", func() {
			var account *domain.Account

			BeforeEach(func(ctx SpecContext) {
				account = createAccount(ctx)

				// 预热缓存，验证 Update 之后按 ID 读不到旧数据
				_, err := repos.Find(ctx, account.ID)
				Expect(err).NotTo(HaveOccurred())
			})

			It("persists the change and invalidates the cached account", func(ctx SpecContext) {
				Expect(account.SetEmail(updatedEmail)).To(Succeed())
				Expect(repos.Update(ctx, account)).To(Succeed())

				updated, err := repos.FindByEmail(ctx, updatedEmail)
				Expect(err).NotTo(HaveOccurred())
				Expect(updated).NotTo(BeNil())
				Expect(updated.ID).To(Equal(account.ID))

				cached, err := repos.Find(ctx, account.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cached).NotTo(BeNil())
				Expect(cached.Email).To(Equal(updatedEmail))
			}, specTimeout)
		})
	})
})

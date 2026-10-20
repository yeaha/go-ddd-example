package service

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ddd-example/internal/domain"

	"github.com/google/uuid"
)

var _ = Describe("SessionTokenService", func() {
	const salt = "54381095jfepwoqrp2"

	var (
		service SessionTokenService
		account *domain.Account
		token   SessionToken
	)

	BeforeEach(func() {
		account = &domain.Account{
			ID:          uuid.New(),
			SessionSalt: salt,
		}
		token = newSessionToken(account)
	})

	It("encodes a non-empty payload", func() {
		Expect(service.encode(token, salt)).NotTo(BeEmpty())
	})

	Describe("round-tripping a token", func() {
		var (
			payload string
			decoded SessionToken
			err     error
		)

		JustBeforeEach(func() {
			payload = service.encode(token, salt)
			decoded, err = service.decode(payload)
		})

		It("decodes back to the original token and re-encodes to the same payload", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(decoded).To(Equal(token))
			Expect(service.encode(decoded, salt)).To(Equal(payload))
		})

		It("produces a different payload with a different salt", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(service.encode(decoded, "abcfaof")).NotTo(Equal(payload))
		})
	})
})

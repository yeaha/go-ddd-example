package domain

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Account", func() {
	var account *Account

	BeforeEach(func() {
		account = &Account{}
	})

	Describe("password", func() {
		BeforeEach(func() {
			Expect(account.SetPassword("abcdefg")).To(Succeed())
		})

		It("only matches the password it was set with", func() {
			Expect(account.ComparePassword("abcdefg")).To(BeTrue())
			Expect(account.ComparePassword("abcdef")).To(BeFalse())
		})
	})
})

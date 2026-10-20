package e2e

import (
	"net/http"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth API", func() {
	var e *httpexpect.Expect

	BeforeEach(func() {
		e = newExpect()
	})

	Describe("注册", func() {
		It("注册成功并建立会话", func() {
			email := uniqueEmail()
			registerAccount(e, email, "secret123")

			body := e.GET(`/session`).Expect().
				Status(http.StatusOK).
				JSON().Object()

			expectErrno(body, 0)
			data := body.Value("data").Object()
			data.Value("email").String().IsEqual(email)
			id, err := uuid.Parse(data.Value("id").String().Raw())
			Expect(err).NotTo(HaveOccurred())
			Expect(id).NotTo(Equal(uuid.Nil))
		})

		DescribeTable("非法注册参数",
			func(payload map[string]any) {
				body := e.POST(`/register`).
					WithJSON(payload).
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object()

				expectErrno(body, 40000)
			},
			Entry("email格式错误", map[string]any{
				"email":    "not-an-email",
				"password": "secret123",
			}),
			Entry("缺少密码", map[string]any{
				"email": uniqueEmail(),
			}),
		)

		It("拒绝重复注册的email", func() {
			email := uniqueEmail()
			registerAccount(e, email, "secret123")

			body := e.POST(`/register`).
				WithJSON(map[string]string{
					"email":    email,
					"password": "another-password",
				}).
				Expect().
				Status(http.StatusConflict).
				JSON().Object()

			expectErrno(body, 40002)
		})
	})

	Describe("登录", func() {
		It("使用正确密码登录并建立会话", func() {
			email := uniqueEmail()
			registerAccount(newExpect(), email, "secret123")

			e.POST(`/session`).
				WithJSON(map[string]string{
					"email":    email,
					"password": "secret123",
				}).
				Expect().
				Status(http.StatusCreated)

			body := e.GET(`/session`).Expect().
				Status(http.StatusOK).
				JSON().Object()

			expectErrno(body, 0)
			body.Value("data").Object().Value("email").String().IsEqual(email)
		})

		It("拒绝错误密码", func() {
			email := uniqueEmail()
			registerAccount(newExpect(), email, "secret123")

			body := e.POST(`/session`).
				WithJSON(map[string]string{
					"email":    email,
					"password": "wrong-password",
				}).
				Expect().
				Status(http.StatusUnauthorized).
				JSON().Object()

			expectErrno(body, 40001)
		})

		It("拒绝不存在的账号", func() {
			body := e.POST(`/session`).
				WithJSON(map[string]string{
					"email":    uniqueEmail(),
					"password": "secret123",
				}).
				Expect().
				Status(http.StatusUnauthorized).
				JSON().Object()

			expectErrno(body, 40001)
		})

		It("匿名访问当前账号信息返回401", func() {
			body := e.GET(`/session`).Expect().
				Status(http.StatusUnauthorized).
				JSON().Object()

			expectErrno(body, 40001)
		})
	})

	Describe("退出登录", func() {
		It("登录后退出并使会话失效", func() {
			email := uniqueEmail()
			registerAccount(e, email, "secret123")

			e.DELETE(`/session`).Expect().
				Status(http.StatusOK)

			body := e.GET(`/session`).Expect().
				Status(http.StatusUnauthorized).
				JSON().Object()

			expectErrno(body, 40001)
		})

		It("匿名退出登录也返回成功", func() {
			e.DELETE(`/session`).Expect().
				Status(http.StatusOK)
		})
	})

	Describe("修改密码", func() {
		It("匿名访问返回401", func() {
			body := e.PUT(`/my/password`).
				WithJSON(map[string]string{
					"old_password": "old-secret",
					"new_password": "new-secret",
				}).
				Expect().
				Status(http.StatusUnauthorized).
				JSON().Object()

			expectErrno(body, 40001)
		})

		It("旧密码错误返回406", func() {
			email := uniqueEmail()
			registerAccount(e, email, "old-secret")

			body := e.PUT(`/my/password`).
				WithJSON(map[string]string{
					"old_password": "wrong-secret",
					"new_password": "new-secret",
				}).
				Expect().
				Status(http.StatusNotAcceptable).
				JSON().Object()

			expectErrno(body, 40003)
		})

		It("修改密码后旧密码失效", func() {
			email := uniqueEmail()
			registerAccount(e, email, "old-secret")

			// 成功时handler不写响应体
			e.PUT(`/my/password`).
				WithJSON(map[string]string{
					"old_password": "old-secret",
					"new_password": "new-secret",
				}).
				Expect().
				Status(http.StatusOK).
				Body().IsEmpty()

			e.DELETE(`/session`).Expect().
				Status(http.StatusOK)

			e.POST(`/session`).
				WithJSON(map[string]string{
					"email":    email,
					"password": "new-secret",
				}).
				Expect().
				Status(http.StatusCreated)

			body := e.POST(`/session`).
				WithJSON(map[string]string{
					"email":    email,
					"password": "old-secret",
				}).
				Expect().
				Status(http.StatusUnauthorized).
				JSON().Object()

			expectErrno(body, 40001)
		})
	})
})

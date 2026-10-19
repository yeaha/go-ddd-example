package e2e

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"ddd-example/internal/bootstrap"
	"ddd-example/internal/option"
	"ddd-example/pkg/oauth"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// baseURL 进程内启动的http服务地址
var baseURL string

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

// 所有ginkgo进程共享同一个http服务：由进程1启动进程内http服务，服务地址通过
// SynchronizedBeforeSuite广播给其余进程。进程1注册的cleanup会等所有进程
// 结束后才执行，因此临时目录与服务生命周期覆盖整个测试运行。
var _ = SynchronizedBeforeSuite(func() []byte {
	// e2e只关心接口行为，屏蔽被测服务日志，保持输出精简
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	tmpDir, err := os.MkdirTemp("", "ddd-example-e2e-*")
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	opt := &option.Options{DBDir: tmpDir}
	opt.HTTP.Addr = "127.0.0.1:0"
	opt.Oauth = map[string]oauth.Options{
		"facebook": {ClientID: "test-client-id", ClientSecret: "test-client-secret"},
	}

	app, err := bootstrap.New(opt)
	Expect(err).NotTo(HaveOccurred())

	addr, err := app.Start(context.Background())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() {
		Expect(app.Shutdown()).To(Succeed())
	})

	return []byte("http://" + addr)
}, func(data []byte) {
	baseURL = string(data)
})

// newExpect 创建一个独立的http客户端，cookie jar互不共享
func newExpect() *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  baseURL,
		Reporter: GinkgoT(),
		Client: &http.Client{
			Jar:     httpexpect.NewCookieJar(),
			Timeout: 5 * time.Second,
		},
	})
}

func uniqueEmail() string {
	return fmt.Sprintf("e2e-%s@example.com", uuid.NewString())
}

func expectErrno(object *httpexpect.Object, code int) {
	GinkgoHelper()

	object.Value("errno").Number().IsEqual(code)
}

func registerAccount(e *httpexpect.Expect, email, password string) {
	GinkgoHelper()

	e.POST(`/register`).
		WithJSON(map[string]string{
			"email":    email,
			"password": password,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().
		Value("errno").Number().IsEqual(0)
}

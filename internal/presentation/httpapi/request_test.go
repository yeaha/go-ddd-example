package httpapi

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type scanRequestPayload struct {
	Page  int    `json:"page"`
	Name  string `json:"name"`
	Email string `json:"email" validate:"omitempty,email"`
}

var _ = Describe("scanRequest", func() {
	newRequest := func(method, target, contentType, body string) *http.Request {
		GinkgoHelper()

		var reader io.Reader
		if body != "" {
			reader = strings.NewReader(body)
		}

		req := httptest.NewRequest(method, target, reader)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		return req
	}

	DescribeTable("scanning an http request into a payload",
		func(method, target, contentType, body string, want *scanRequestPayload, wantErr string) {
			payload := &scanRequestPayload{}
			err := scanRequest(payload, newRequest(method, target, contentType, body))

			if want != nil {
				Expect(err).NotTo(HaveOccurred())
				Expect(payload).To(Equal(want))
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(wantErr))
			}
		},
		Entry("ok - GET query", http.MethodGet, "/?page=1&name=alice", "", "", &scanRequestPayload{Page: 1, Name: "alice"}, ""),
		Entry("ok - GET empty query", http.MethodGet, "/", "", "", &scanRequestPayload{}, ""),
		Entry("ok - DELETE query", http.MethodDelete, "/?page=2&name=bob", "", "", &scanRequestPayload{Page: 2, Name: "bob"}, ""),
		Entry("bad - GET invalid int", http.MethodGet, "/?page=abc", "", "", nil, "decode query string"),
		Entry("bad - unsupported method", http.MethodOptions, "/", "", "", nil, "unsupported http method"),
		Entry("ok - POST json", http.MethodPost, "/", "application/json", `{"page":3,"name":"c","email":"a@b.c"}`, &scanRequestPayload{Page: 3, Name: "c", Email: "a@b.c"}, ""),
		Entry("ok - PUT json", http.MethodPut, "/", "application/json", `{"page":4,"name":"d"}`, &scanRequestPayload{Page: 4, Name: "d"}, ""),
		Entry("bad - POST bad json", http.MethodPost, "/", "application/json", `{bad`, nil, "json decode"),
		Entry("bad - POST missing content type", http.MethodPost, "/", "", `{"page":1}`, nil, "parse content-type"),
		Entry("bad - POST content type without slash", http.MethodPost, "/", "application", `{"page":1}`, nil, "unsupported content type"),
		Entry("bad - POST unsupported content type", http.MethodPost, "/", "text/plain", "page=1", nil, "unsupported content type"),
		Entry("ok - POST form", http.MethodPost, "/", "application/x-www-form-urlencoded", "page=5&name=e", &scanRequestPayload{Page: 5, Name: "e"}, ""),
		Entry("bad - GET invalid email", http.MethodGet, "/?page=1&name=f&email=bad", "", "", nil, "Key: 'scanRequestPayload.Email' Error:Field validation for 'Email' failed on the 'email' tag"),
	)

	It("ok - POST multipart form", func() {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		Expect(mw.WriteField("page", "6")).To(Succeed())
		Expect(mw.WriteField("name", "g")).To(Succeed())
		Expect(mw.Close()).To(Succeed())

		payload := &scanRequestPayload{}
		err := scanRequest(payload, newRequest(http.MethodPost, "/", mw.FormDataContentType(), buf.String()))
		Expect(err).NotTo(HaveOccurred())
		Expect(payload).To(Equal(&scanRequestPayload{Page: 6, Name: "g"}))
	})
})

package httpapi

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type scanRequestPayload struct {
	Page  int    `json:"page"`
	Name  string `json:"name"`
	Email string `json:"email" validate:"omitempty,email"`
}

func TestScanRequest(t *testing.T) {
	newRequest := func(method, target, contentType, body string) *http.Request {
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

	assertScan := func(t *testing.T, payload any, req *http.Request, want *scanRequestPayload, wantErr string) {
		t.Helper()

		err := scanRequest(payload, req)

		if want != nil {
			if err != nil {
				t.Fatalf("scan should pass, got %v", err)
			} else if !reflect.DeepEqual(payload, want) {
				t.Fatalf("scan should decode %+v, got %+v", want, payload)
			}
		} else if err == nil {
			t.Fatal("scan should fail, got nil")
		} else if !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("scan should fail with %q, got %q", wantErr, err)
		}
	}

	tests := []struct {
		name        string
		method      string
		target      string
		contentType string
		body        string
		want        *scanRequestPayload
		wantErr     string
	}{
		{
			name:        "ok - GET query",
			method:      http.MethodGet,
			target:      "/?page=1&name=alice",
			contentType: "",
			body:        "",
			want:        &scanRequestPayload{Page: 1, Name: "alice"},
			wantErr:     "",
		},
		{
			name:        "ok - GET empty query",
			method:      http.MethodGet,
			target:      "/",
			contentType: "",
			body:        "",
			want:        &scanRequestPayload{},
			wantErr:     "",
		},
		{
			name:        "ok - DELETE query",
			method:      http.MethodDelete,
			target:      "/?page=2&name=bob",
			contentType: "",
			body:        "",
			want:        &scanRequestPayload{Page: 2, Name: "bob"},
			wantErr:     "",
		},
		{
			name:        "bad - GET invalid int",
			method:      http.MethodGet,
			target:      "/?page=abc",
			contentType: "",
			body:        "",
			want:        nil,
			wantErr:     "decode values",
		},
		{
			name:        "bad - unsupported method",
			method:      http.MethodOptions,
			target:      "/",
			contentType: "",
			body:        "",
			want:        nil,
			wantErr:     "unsupported http method",
		},
		{
			name:        "ok - POST json",
			method:      http.MethodPost,
			target:      "/",
			contentType: "application/json",
			body:        `{"page":3,"name":"c","email":"a@b.c"}`,
			want:        &scanRequestPayload{Page: 3, Name: "c", Email: "a@b.c"},
			wantErr:     "",
		},
		{
			name:        "ok - PUT json",
			method:      http.MethodPut,
			target:      "/",
			contentType: "application/json",
			body:        `{"page":4,"name":"d"}`,
			want:        &scanRequestPayload{Page: 4, Name: "d"},
			wantErr:     "",
		},
		{
			name:        "bad - POST bad json",
			method:      http.MethodPost,
			target:      "/",
			contentType: "application/json",
			body:        `{bad`,
			want:        nil,
			wantErr:     "json decode",
		},
		{
			name:        "bad - POST missing content type",
			method:      http.MethodPost,
			target:      "/",
			contentType: "",
			body:        `{"page":1}`,
			want:        nil,
			wantErr:     "parse content-type",
		},
		{
			name:        "bad - POST content type without slash",
			method:      http.MethodPost,
			target:      "/",
			contentType: "application",
			body:        `{"page":1}`,
			want:        nil,
			wantErr:     "unsupported content type",
		},
		{
			name:        "bad - POST unsupported content type",
			method:      http.MethodPost,
			target:      "/",
			contentType: "text/plain",
			body:        "page=1",
			want:        nil,
			wantErr:     "unsupported content type",
		},
		{
			name:        "ok - POST form",
			method:      http.MethodPost,
			target:      "/",
			contentType: "application/x-www-form-urlencoded",
			body:        "page=5&name=e",
			want:        &scanRequestPayload{Page: 5, Name: "e"},
			wantErr:     "",
		},
		{
			name:        "bad - GET invalid email",
			method:      http.MethodGet,
			target:      "/?page=1&name=f&email=bad",
			contentType: "",
			body:        "",
			want:        nil,
			wantErr:     "validate payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertScan(
				t,
				&scanRequestPayload{},
				newRequest(tt.method, tt.target, tt.contentType, tt.body),
				tt.want, tt.wantErr,
			)
		})
	}

	t.Run("ok - POST multipart form", func(t *testing.T) {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)

		if err := mw.WriteField("page", "6"); err != nil {
			t.Fatal(err)
		} else if err = mw.WriteField("name", "g"); err != nil {
			t.Fatal(err)
		}
		mw.Close()

		assertScan(
			t,
			&scanRequestPayload{},
			newRequest(http.MethodPost, "/", mw.FormDataContentType(), buf.String()),
			&scanRequestPayload{Page: 6, Name: "g"}, "",
		)
	})
}

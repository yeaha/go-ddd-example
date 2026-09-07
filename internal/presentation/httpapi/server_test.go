package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testArgs struct {
	Page int `json:"page"`
}

func factor(t *testing.T, fn any) *funcFactor {
	t.Helper()

	f, err := newFuncFactor(fn)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func assertCheck(t *testing.T, err error, want string) {
	t.Helper()

	switch {
	case want == "" && err != nil:
		t.Fatalf("check should pass, got %q", err)
	case want != "" && err == nil:
		t.Fatalf("check should fail with %q, got nil", want)
	case want != "" && err.Error() != want:
		t.Fatalf("check should fail with %q, got %q", want, err)
	}
}

func TestCheckHandlerParameters(t *testing.T) {
	tests := []struct {
		name    string
		handler any
		wantErr string
	}{
		{"ok - no args", func(context.Context) error { return nil }, ""},
		{"ok - pointer args", func(context.Context, *testArgs) error { return nil }, ""},
		{"ok - struct args", func(context.Context, testArgs) error { return nil }, ""},
		{"ok - multi outputs", func(context.Context, *testArgs) (int, string, error) { return 0, "", nil }, ""},
		{"bad - no inputs", func() error { return nil }, "handler should accept 1 or 2 inputs"},
		{"bad - too many inputs", func(context.Context, *testArgs, *testArgs) error { return nil }, "handler should accept 1 or 2 inputs"},
		{"bad - no outputs", func(context.Context, *testArgs) {}, "handler should have at least 1 output"},
		{"bad - first input not context", func(*testArgs) error { return nil }, "first handler input should be context.Context"},
		{"bad - second input not struct", func(context.Context, int) error { return nil }, "second handler input should be a struct or struct pointer"},
		{"bad - last output not error", func(context.Context) string { return "" }, "last handler output should be error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCheck(t, checkHandlerParameters(factor(t, tt.handler)), tt.wantErr)
		})
	}
}

func TestCheckRenderParameters(t *testing.T) {
	tests := []struct {
		name    string
		handler any
		render  any
		wantErr string
	}{
		{
			"ok - error only",
			func(context.Context) error { return nil },
			func(http.ResponseWriter, *http.Request) error { return nil },
			"",
		},
		{
			"ok - pass error to render",
			func(context.Context) error { return nil },
			func(http.ResponseWriter, *http.Request, error) error { return nil },
			"",
		},
		{
			"ok - data without error input",
			func(context.Context) (string, error) { return "", nil },
			func(http.ResponseWriter, *http.Request, string) error { return nil },
			"",
		},
		{
			"ok - data with error input",
			func(context.Context) (string, error) { return "", nil },
			func(http.ResponseWriter, *http.Request, string, error) (any, error) { return nil, nil },
			"",
		},
		{
			"bad - too few inputs",
			func(context.Context) error { return nil },
			func(*http.Request) error { return nil },
			"render should accept at least 2 inputs",
		},
		{
			"bad - first input not ResponseWriter",
			func(context.Context) error { return nil },
			func(*testArgs, *http.Request) error { return nil },
			"first render input should be http.ResponseWriter",
		},
		{
			"bad - second input not request",
			func(context.Context) error { return nil },
			func(http.ResponseWriter, *testArgs) error { return nil },
			"second render input should be *http.Request",
		},
		{
			"bad - missing inputs",
			func(context.Context) (string, error) { return "", nil },
			func(http.ResponseWriter, *http.Request) error { return nil },
			"render inputs do not match handler outputs",
		},
		{
			"bad - extra inputs",
			func(context.Context) error { return nil },
			func(http.ResponseWriter, *http.Request, string, string) error { return nil },
			"render inputs do not match handler outputs",
		},
		{
			"bad - input type mismatch",
			func(context.Context) (string, error) { return "", nil },
			func(http.ResponseWriter, *http.Request, int) error { return nil },
			"render input 3 type int does not match handler output string",
		},
		{
			"bad - last output not error",
			func(context.Context) (string, error) { return "", nil },
			func(http.ResponseWriter, *http.Request, string) string { return "" },
			"last render output should be error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, render := factor(t, tt.handler), factor(t, tt.render)
			assertCheck(t, checkRenderParameters(handler, render), tt.wantErr)
		})
	}
}

func TestNewHandlerConstruction(t *testing.T) {
	tests := []struct {
		name       string
		appHandler any
		render     any
		wantErr    string
	}{
		{
			name:       "ok - 2 inputs",
			appHandler: func(context.Context, *testArgs) error { return nil },
			render:     func(http.ResponseWriter, *http.Request) error { return nil },
			wantErr:    "",
		},
		{
			name:       "ok - data out, render without error",
			appHandler: func(context.Context) (string, error) { return "", nil },
			render:     func(http.ResponseWriter, *http.Request, string) error { return nil },
			wantErr:    "",
		},
		{
			name:       "ok - pass error to render",
			appHandler: func(context.Context) error { return nil },
			render:     func(http.ResponseWriter, *http.Request, error) error { return nil },
			wantErr:    "",
		},
		{
			name:       "ok - render returns data",
			appHandler: func(context.Context) error { return nil },
			render:     func(http.ResponseWriter, *http.Request) (any, error) { return nil, nil },
			wantErr:    "",
		},
		{
			name:       "bad - handler not a function",
			appHandler: "not a function",
			render:     func(http.ResponseWriter, *http.Request) error { return nil },
			wantErr:    "not a function",
		},
		{
			name:       "bad - handler invalid",
			appHandler: func() error { return nil },
			render:     func(http.ResponseWriter, *http.Request) error { return nil },
			wantErr:    "handler should accept 1 or 2 inputs",
		},
		{
			name:       "bad - render not a function",
			appHandler: func(context.Context) error { return nil },
			render:     "not a function",
			wantErr:    "not a function",
		},
		{
			name:       "bad - render invalid",
			appHandler: func(context.Context) error { return nil },
			render:     func(http.ResponseWriter) error { return nil },
			wantErr:    "render should accept at least 2 inputs",
		},
		{
			name:       "bad - 2 inputs, second not struct",
			appHandler: func(context.Context, int) error { return nil },
			render:     func(http.ResponseWriter, *http.Request) error { return nil },
			wantErr:    "second handler input should be a struct or struct pointer",
		},
		{
			name:       "bad - 3 inputs",
			appHandler: func(context.Context, *testArgs, *testArgs) error { return nil },
			render:     func(http.ResponseWriter, *http.Request) error { return nil },
			wantErr:    "handler should accept 1 or 2 inputs",
		},
		{
			name:       "bad - last handler output not error",
			appHandler: func(context.Context) string { return "" },
			render:     func(http.ResponseWriter, *http.Request) error { return nil },
			wantErr:    "last handler output should be error",
		},
		{
			name:       "bad - render inputs do not match handler outputs",
			appHandler: func(context.Context) (string, error) { return "", nil },
			render:     func(http.ResponseWriter, *http.Request) error { return nil },
			wantErr:    "render inputs do not match handler outputs",
		},
		{
			name:       "bad - handler checked before render",
			appHandler: func() error { return nil },
			render:     func(http.ResponseWriter) error { return nil },
			wantErr:    "handler should accept 1 or 2 inputs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHandler(tt.appHandler, tt.render)
			assertCheck(t, err, tt.wantErr)
		})
	}
}

func TestMustNewHandler(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("should panic on invalid handler")
		}
	}()

	MustNewHandler("not a function", func(http.ResponseWriter, *http.Request) error { return nil })
}

func TestMustNewVoidHandler(t *testing.T) {
	handler := MustNewVoidHandler(func(context.Context, *testArgs) error { return nil })
	if handler == nil {
		t.Fatal("handler should not be nil")
	}
}

func TestNewVoidHandler(t *testing.T) {
	sentinelErr := errors.New("sentinel")

	t.Run("ok - pointer args", func(t *testing.T) {
		var gotPage int
		handler := MustNewVoidHandler(func(ctx context.Context, args *testArgs) error {
			gotPage = args.Page
			return nil
		})

		w := serveRequest(t, handler, http.MethodPost, "/", `{"page": 3}`, "application/json")

		if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if w.Body.Len() > 0 {
			t.Fatalf("body should be empty, got %q", w.Body.String())
		} else if gotPage != 3 {
			t.Fatalf("args should be parsed page 3, got %d", gotPage)
		}
	})

	t.Run("ok - struct args", func(t *testing.T) {
		var gotPage int
		handler := MustNewVoidHandler(func(ctx context.Context, args testArgs) error {
			gotPage = args.Page
			return nil
		})

		w := serveRequest(t, handler, http.MethodPost, "/", `{"page": 7}`, "application/json")

		if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if w.Body.Len() > 0 {
			t.Fatalf("body should be empty, got %q", w.Body.String())
		} else if gotPage != 7 {
			t.Fatalf("args should be parsed page 7, got %d", gotPage)
		}
	})

	t.Run("bad - handler error responds 500", func(t *testing.T) {
		handler := MustNewVoidHandler(func(ctx context.Context, args *testArgs) error {
			return sentinelErr
		})

		w := serveRequest(t, handler, http.MethodPost, "/", `{"page": 1}`, "application/json")

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status should be 500, got %d", w.Code)
		}

		v := parseResp(t, w)
		if v.Errno != 50000 {
			t.Fatalf("errno should be 50000, got %d", v.Errno)
		}
	})
}

func TestNewHandlerServeParams(t *testing.T) {
	sentinelErr := errors.New("sentinel")

	type ctxKey struct{}

	t.Run("ok - handler first input is request context", func(t *testing.T) {
		key := ctxKey{}
		value := "ctx-value"

		handler := MustNewHandler(
			func(ctx context.Context) error {
				if got := ctx.Value(key); got != value {
					return errors.New("context value mismatch")
				}
				return nil
			},
			func(w http.ResponseWriter, r *http.Request) error { return nil },
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), key, value))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if w.Body.Len() > 0 {
			t.Fatalf("body should be empty, got %q", w.Body.String())
		}
	})

	t.Run("ok - handler second input is request args pointer", func(t *testing.T) {
		handler := MustNewHandler(
			func(ctx context.Context, args *testArgs) (int, error) {
				if args == nil {
					return 0, errors.New("args should not be nil")
				}
				return args.Page, nil
			},
			func(w http.ResponseWriter, r *http.Request, page int) (any, error) {
				return mapAny{"page": page}, nil
			},
		)

		w := serveRequest(t, handler, http.MethodGet, "/", "", "")
		v := parseResp(t, w)

		if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if m, _ := v.Data.(map[string]any); m["page"] != float64(0) {
			t.Fatalf("data should echo page 0, got %+v", v.Data)
		}
	})

	t.Run("ok - handler second input is request args struct", func(t *testing.T) {
		handler := MustNewHandler(
			func(ctx context.Context, args testArgs) (int, error) {
				return args.Page, nil
			},
			func(w http.ResponseWriter, r *http.Request, page int) (any, error) {
				return mapAny{"page": page}, nil
			},
		)

		w := serveRequest(t, handler, http.MethodPost, "/", `{"page": 3}`, "application/json")
		v := parseResp(t, w)

		if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if m, _ := v.Data.(map[string]any); m["page"] != float64(3) {
			t.Fatalf("data should echo page 3, got %+v", v.Data)
		}
	})

	t.Run("ok - handler data outputs forwarded, no error input", func(t *testing.T) {
		var id int
		var name string

		handler := MustNewHandler(
			func(ctx context.Context) (int, string, error) { return 1, "alice", nil },
			func(w http.ResponseWriter, r *http.Request, i int, s string) error {
				id, name = i, s
				return nil
			},
		)

		w := serveRequest(t, handler, http.MethodGet, "/", "", "")

		if id != 1 || name != "alice" {
			t.Fatalf("render should receive data outs (%d, %q), got (%d, %q)", 1, "alice", id, name)
		} else if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if w.Body.Len() > 0 {
			t.Fatalf("body should be empty, got %q", w.Body.String())
		}
	})

	t.Run("ok - handler data and error forwarded with error input", func(t *testing.T) {
		var id int
		var got error

		handler := MustNewHandler(
			func(ctx context.Context) (int, error) { return 1, sentinelErr },
			func(w http.ResponseWriter, r *http.Request, i int, err error) (any, error) {
				id, got = i, err
				return mapAny{"echo": i}, nil
			},
		)

		w := serveRequest(t, handler, http.MethodGet, "/", "", "")
		v := parseResp(t, w)

		if id != 1 {
			t.Fatalf("render should receive data out %d, got %d", 1, id)
		} else if !errors.Is(got, sentinelErr) {
			t.Fatalf("render should receive %v, got %v", sentinelErr, got)
		} else if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if m, _ := v.Data.(map[string]any); m["echo"] != float64(1) {
			t.Fatalf("data should echo 1, got %+v", v.Data)
		}
	})

	t.Run("ok - handler nil data nil error forwarded with error input", func(t *testing.T) {
		var s string
		var got error

		handler := MustNewHandler(
			func(ctx context.Context) (string, error) { return "", nil },
			func(w http.ResponseWriter, r *http.Request, v string, err error) (any, error) {
				s, got = v, err
				return mapAny{"echo": v}, nil
			},
		)

		w := serveRequest(t, handler, http.MethodGet, "/", "", "")
		v := parseResp(t, w)

		if s != "" {
			t.Fatalf("render should receive empty data, got %q", s)
		} else if got != nil {
			t.Fatalf("render should receive nil error, got %v", got)
		} else if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if m, _ := v.Data.(map[string]any); m["echo"] != "" {
			t.Fatalf("data should echo empty string, got %+v", v.Data)
		}
	})

	t.Run("ok - error-only handler forwards only error with error input", func(t *testing.T) {
		var got error

		handler := MustNewHandler(
			func(ctx context.Context) error { return sentinelErr },
			func(w http.ResponseWriter, r *http.Request, err error) error {
				got = err
				return nil
			},
		)

		w := serveRequest(t, handler, http.MethodGet, "/", "", "")

		if !errors.Is(got, sentinelErr) {
			t.Fatalf("render should receive %v, got %v", sentinelErr, got)
		} else if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if w.Body.Len() > 0 {
			t.Fatalf("body should be empty, got %q", w.Body.String())
		}
	})

	t.Run("bad - handler error without error input drops data, render skipped", func(t *testing.T) {
		renderCalled := false

		handler := MustNewHandler(
			func(ctx context.Context) (string, error) { return "value", sentinelErr },
			func(w http.ResponseWriter, r *http.Request, v string) error {
				renderCalled = true
				return nil
			},
		)

		w := serveRequest(t, handler, http.MethodGet, "/", "", "")
		v := parseResp(t, w)

		if renderCalled {
			t.Fatal("render should not be called")
		} else if w.Code != http.StatusInternalServerError {
			t.Fatalf("status should be 500, got %d", w.Code)
		} else if v.Errno != errUnexpectedException.code || v.Error != errUnexpectedException.message {
			t.Fatalf("response should be errUnexpectedException, got %+v", v)
		}
	})

	t.Run("bad - render error - 500", func(t *testing.T) {
		handler := MustNewHandler(
			func(ctx context.Context) error { return nil },
			func(w http.ResponseWriter, r *http.Request) error { return sentinelErr },
		)

		w := serveRequest(t, handler, http.MethodGet, "/", "", "")
		v := parseResp(t, w)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status should be 500, got %d", w.Code)
		} else if v.Errno != errUnexpectedException.code {
			t.Fatalf("errno should be %d, got %d", errUnexpectedException.code, v.Errno)
		}
	})

	t.Run("ok - all nil, no data written - empty 200", func(t *testing.T) {
		handler := MustNewHandler(
			func(ctx context.Context) error { return nil },
			func(w http.ResponseWriter, r *http.Request) error { return nil },
		)

		w := serveRequest(t, handler, http.MethodGet, "/", "", "")

		if w.Code != http.StatusOK {
			t.Fatalf("status should be 200, got %d", w.Code)
		} else if w.Body.Len() > 0 {
			t.Fatalf("body should be empty, got %q", w.Body.String())
		}
	})
}

func serveRequest(t *testing.T, handler http.Handler, method, target, body, contentType string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func parseResp(t *testing.T, w *httptest.ResponseRecorder) apiResponse {
	t.Helper()

	var v apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("unmarshal response %q: %v", w.Body.String(), err)
	}
	return v
}

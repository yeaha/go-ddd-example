package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"ddd-example/pkg/logger"
)

// recoverer 捕获接口层抛出的错误
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(
			logger.NewContext(
				r.Context(),
				logger.FromContext(r.Context()).
					With("@request", fmt.Sprintf("%s %s", r.Method, r.URL.Path)),
			),
		)

		defer func() {
			if v := recover(); v != nil {
				switch v := v.(type) {
				case apiError:
					if err := errors.Unwrap(v); err != nil {
						if code := v.StatusCode(); code >= http.StatusInternalServerError {
							logger.Error(r.Context(), "recover http error", "error", err)
						} else {
							logger.Debug(r.Context(), "recover http error", "error", err)
						}
					}

					sendResponse(w, withError(v))
				default:
					logger.Error(r.Context(), "recover panic",
						"error", v,
						"trace", string(debug.Stack()),
					)

					sendResponse(w, withError(errUnexpectedException))
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

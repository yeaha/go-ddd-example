package httpapi

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// recoverer 捕获接口层抛出的错误
func recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					switch v := v.(type) {
					case apiError:
						if causeErr := errors.Unwrap(v); causeErr != nil {
							entry := logger.With()
							if f, ok := v.WrapLine(); ok {
								entry = entry.With(
									"file", fmt.Sprintf("%s:%d", f.File, f.Line),
									"func", f.Function,
								)
							} else {
								entry = entry.With(
									"method", r.Method,
									"uri", r.URL.Path,
								)
							}

							if code := v.StatusCode(); code >= http.StatusInternalServerError {
								entry.Error("recover http error")
							} else {
								entry.Debug("recover http error")
							}
						}

						sendResponse(w, withError(v))
					default:
						logger.Error("recover panic",
							"method", r.Method,
							"uri", r.URL.Path,
							"error", v,
						)

						sendResponse(w, withError(errUnexpectedException))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

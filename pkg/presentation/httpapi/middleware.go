package httpapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// recoverer 捕获接口层抛出的错误
func recoverer(logger logrus.FieldLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					switch v := v.(type) {
					case apiError:
						if causeErr := errors.Unwrap(v); causeErr != nil {
							entry := logger.WithError(causeErr)
							if f, ok := v.WrapLine(); ok {
								entry = entry.WithFields(logrus.Fields{
									logrus.FieldKeyFile: fmt.Sprintf("%s:%d", f.File, f.Line),
									logrus.FieldKeyFunc: f.Function,
								})
							} else {
								entry = entry.WithFields(logrus.Fields{
									"method": r.Method,
									"uri":    r.URL.Path,
								})
							}

							if code := v.StatusCode(); code >= http.StatusInternalServerError {
								entry.Error("recover http error")
							} else {
								entry.Debug("recover http error")
							}
						}

						sendResponse(w, withError(v))
					default:
						logrus.WithField("error", v).
							WithFields(logrus.Fields{
								"method": r.Method,
								"url":    r.URL.Path,
							}).
							Error("recover panic")

						sendResponse(w, withError(errUnexpectedException))
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

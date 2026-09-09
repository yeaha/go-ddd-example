package httpapi

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

var (
	requestDecoder = schema.NewDecoder()

	requestValidator = validator.New(
		validator.WithRequiredStructEnabled(),
	)
)

const defaultMaxMemory = 32 << 20 // 32MB

func init() {
	requestDecoder.IgnoreUnknownKeys(true)
	requestDecoder.SetAliasTag("json")
}

func scanRequest(payload any, r *http.Request) (err error) {
	defer func() {
		if err == nil {
			if err = validatePayload(payload); err != nil {
				err = fmt.Errorf("validate payload, %w", err)
			}
		}
	}()

	switch r.Method {
	default:
		return fmt.Errorf("unsupported http method %s", r.Method)

	case http.MethodGet, http.MethodDelete:
		if values := r.URL.Query(); len(values) > 0 {
			if err := requestDecoder.Decode(payload, values); err != nil {
				return fmt.Errorf("decode query string, %w", err)
			}
		}

	case http.MethodPost, http.MethodPut, http.MethodPatch:
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			return fmt.Errorf("parse content-type, %w", err)
		}

		switch mediaType {
		default:
			return fmt.Errorf("unsupported content type: %s", mediaType)

		case "application/json":
			if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
				return fmt.Errorf("json decode, %w", err)
			}

		case "application/x-www-form-urlencoded", "multipart/form-data":
			if mediaType == "multipart/form-data" {
				if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
					return fmt.Errorf("parse multipart form, %w", err)
				}
			} else {
				if err := r.ParseForm(); err != nil {
					return fmt.Errorf("parse form, %w", err)
				}
			}

			if values := r.Form; len(values) > 0 {
				if err := requestDecoder.Decode(payload, values); err != nil {
					return fmt.Errorf("decode form, %w", err)
				}
			}
		}
	}

	return nil
}

func mustScanRequest(payload any, r *http.Request) {
	if err := scanRequest(payload, r); err != nil {
		panic(errBadRequest.WrapError(err))
	}
}

func validatePayload(payload any) error {
	if v, ok := payload.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return requestValidator.Struct(payload)
}

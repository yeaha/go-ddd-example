package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"

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

func scanJSON(dst any, input io.Reader) error {
	if err := json.NewDecoder(input).Decode(dst); err != nil {
		return fmt.Errorf("json decode, %w", err)
	}

	if err := validatePayload(dst); err != nil {
		return fmt.Errorf("validate payload, %w", err)
	}
	return nil
}

func mustScanJSON(dst any, input io.Reader) {
	if err := scanJSON(dst, input); err != nil {
		panic(errBadRequest.WrapError(err))
	}
}

func scanValues(dst any, values url.Values) error {
	if len(values) > 0 {
		if err := requestDecoder.Decode(dst, values); err != nil {
			return fmt.Errorf("decode values, %w", err)
		}
	}

	if err := validatePayload(dst); err != nil {
		return fmt.Errorf("validate payload, %w", err)
	}
	return nil
}

func mustScanValues(dst any, values url.Values) {
	if err := scanValues(dst, values); err != nil {
		panic(errBadRequest.WrapError(err))
	}
}

func scanRequest(payload any, r *http.Request) error {
	switch r.Method {
	default:
		return fmt.Errorf("unsupported http method %s", r.Method)

	case http.MethodGet, http.MethodDelete:
		return scanValues(payload, r.URL.Query())

	case http.MethodPost, http.MethodPut, http.MethodPatch:
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			return fmt.Errorf("parse content-type, %w", err)
		}

		switch mediaType {
		default:
			return fmt.Errorf("unsupported content type: %s", mediaType)

		case "application/json":
			return scanJSON(payload, r.Body)

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

			return scanValues(payload, r.Form)
		}
	}
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

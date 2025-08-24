package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
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

func init() {
	requestDecoder.IgnoreUnknownKeys(true)
	requestDecoder.SetAliasTag("json")
}

func mustScanJSON(dst any, input io.Reader) {
	if err := scanJSON(dst, input); err != nil {
		panic(errBadRequest.WrapError(err))
	}
}

func mustScanValues(dst any, values url.Values) {
	if err := scanValues(dst, values); err != nil {
		panic(errBadRequest.WrapError(err))
	}
}

func scanJSON(dst any, input io.Reader) error {
	if err := json.NewDecoder(input).Decode(dst); err != nil {
		return fmt.Errorf("json decode, %w", err)
	}

	if v, ok := dst.(interface {
		Validate() error
	}); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validate values, %w", err)
		}
	}

	return requestValidator.Struct(dst)
}

func scanValues(dst any, values url.Values) error {
	if err := requestDecoder.Decode(dst, values); err != nil {
		return fmt.Errorf("decode values, %w", err)
	}

	if v, ok := dst.(interface {
		Validate() error
	}); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validate values, %w", err)
		}
	}

	return requestValidator.Struct(dst)
}

package httpapi

import (
	"context"
	"ddd-example/pkg/logger"
	"encoding/json"
	"net/http"
)

type mapAny map[string]any

// apiResponse 定义了http所有接口的统一下行数据结构
type apiResponse struct {
	Errno int    `json:"errno"`
	Error string `json:"error"`
	Data  any    `json:"data"`

	statusCode int
}

func (r apiResponse) StatusCode() int {
	if n := r.statusCode; n > 0 {
		return n
	}

	return http.StatusOK
}

type apiResponseOption func(*apiResponse)

func withData(data any) apiResponseOption {
	return func(ar *apiResponse) {
		ar.Data = data
	}
}

func withError(err apiError) apiResponseOption {
	return func(ar *apiResponse) {
		ar.Errno = err.code
		ar.Error = err.message
		ar.statusCode = err.StatusCode()

		if data, ok := err.Data(); ok {
			ar.Data = data
		}
	}
}

func withStatusCode(code int) apiResponseOption {
	return func(ar *apiResponse) {
		ar.statusCode = code
	}
}

func sendResponse(w http.ResponseWriter, options ...apiResponseOption) {
	response := &apiResponse{}

	for _, fn := range options {
		fn(response)
	}

	if response.Data == nil {
		response.Data = struct{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode())

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(context.Background(), "send response", "error", err)
	}
}

package httpapi

import (
	"net/http"
)

var (
	errBadRequest        = newAPIError(40000, "上行参数不符合要求", http.StatusBadRequest)
	errUnauthorized      = newAPIError(40001, "账号验证失败", http.StatusUnauthorized)
	errEmailRegistered   = newAPIError(40002, "Email已经被注册", http.StatusConflict)
	errWrongPassword     = newAPIError(40003, "密码验证错误", http.StatusNotAcceptable)
	errOauthNotSupport   = newAPIError(40004, "不支持的oauth服务", http.StatusNotFound)
	errInvalidOauthToken = newAPIError(40005, "oauth凭证无效", http.StatusNotAcceptable)

	errUnexpectedException = newAPIError(50000, "服务器端未知错误", http.StatusInternalServerError)
)

type apiError struct {
	code    int
	message string
	status  int
	cause   error
	data    any
}

func newAPIError(code int, message string, status int) apiError {
	return apiError{
		code:    code,
		message: message,
		status:  status,
	}
}

func (err apiError) Error() string {
	return err.message
}

func (err apiError) Data() (any, bool) {
	return err.data, err.data != nil
}

func (err apiError) StatusCode() int {
	if v := err.status; v > 0 {
		return v
	}

	return http.StatusInternalServerError
}

func (err apiError) Clone() apiError {
	return apiError{
		code:    err.code,
		message: err.message,
		status:  err.status,
		cause:   err.cause,
		data:    err.data,
	}
}

func (err apiError) Unwrap() error {
	return err.cause
}

func (err apiError) WrapError(cause error) apiError {
	clone := err.Clone()
	clone.cause = cause

	return clone
}

func (err apiError) WithData(data any) apiError {
	clone := err.Clone()
	clone.data = data

	return clone
}

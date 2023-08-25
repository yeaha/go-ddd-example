package domain

import "errors"

var (
	// ErrAccountNotFound 账号不存在
	ErrAccountNotFound = errors.New("account not found")
	// ErrWrongPassword 密码错误
	ErrWrongPassword = errors.New("wrong password")
	// ErrEmailRegistered email已注册
	ErrEmailRegistered = errors.New("email has been registered")
	// ErrInvalidSessionToken 无效的会话凭证
	ErrInvalidSessionToken = errors.New("invalid session token")
	// ErrSessionTokenExpired 会话凭证已过期
	ErrSessionTokenExpired = errors.New("session token expired")
	// ErrMissingCache 缓存不存在
	ErrMissingCache = errors.New("missing cache")
	// ErrInvalidOauthToken 无效的三方验证信息缓存凭证
	ErrInvalidOauthToken = errors.New("invalid oauth token")
)

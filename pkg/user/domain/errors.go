package domain

import "errors"

var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")
	// ErrWrongPassword 密码错误
	ErrWrongPassword = errors.New("wrong password")
	// ErrEmailRegistered email已注册
	ErrEmailRegistered = errors.New("email has been registered")
	// ErrInvalidSessionToken 无效的会话凭证
	ErrInvalidSessionToken = errors.New("invalid session token")
	// ErrSessionTokenExpired 会话凭证已过期
	ErrSessionTokenExpired = errors.New("session token expired")
)

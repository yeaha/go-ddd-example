package event

import "ddd-example/internal/domain"

// Login 账号登录
type Login struct {
	User *domain.User
}

// Register 账号注册
type Register struct {
	User *domain.User
}

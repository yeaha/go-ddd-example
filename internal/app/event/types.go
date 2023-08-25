package event

import "ddd-example/internal/domain"

// Login 账号登录
type Login struct {
	Account *domain.Account
}

// Register 账号注册
type Register struct {
	Account *domain.Account
}

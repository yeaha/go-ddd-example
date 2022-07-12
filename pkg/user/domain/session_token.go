package domain

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

var (
	defaultTokenExpire = 30 * 24 * time.Hour // 30天
	defaultTokenRenew  = 7 * 24 * time.Hour
)

// SessionToken 会话凭证
type SessionToken struct {
	UserID uuid.UUID
	Expire int64 // 凭证过期时间
}

// NewSessionToken 生成会话凭证
func NewSessionToken(user *User) SessionToken {
	return SessionToken{
		UserID: user.ID,
		Expire: time.Now().Add(defaultTokenExpire).Unix(),
	}
}

// ExpireTime 过期时间
func (token SessionToken) ExpireTime() time.Time {
	if n := token.Expire; n > 0 {
		return time.Unix(n, 0)
	}
	return time.Time{}
}

// IsExpired 是否过期
func (token SessionToken) IsExpired() bool {
	if t := token.ExpireTime(); !t.IsZero() {
		return t.Before(time.Now())
	}
	return true
}

// NeedRenew 是否需要延期
func (token SessionToken) NeedRenew() bool {
	if t := token.ExpireTime(); !t.IsZero() {
		return t.Before(time.Now().Add(defaultTokenRenew))
	}
	return false
}

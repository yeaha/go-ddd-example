package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/domain"

	"github.com/google/uuid"
)

var (
	tokenExpire = 30 * 24 * time.Hour // 30天
	tokenRenew  = 7 * 24 * time.Hour
)

// SessionTokenService 会话凭证逻辑
type SessionTokenService struct {
	Accounts adapter.AccountRepository `do:""`
}

// Generate 构造会话凭证
func (s *SessionTokenService) Generate(ctx context.Context, account *domain.Account) (payload string, err error) {
	if err := s.Suspend(ctx, account); err != nil {
		return "", err
	}

	return s.Renew(account)
}

// Renew 构造会话凭证，但不刷新session salt
func (s *SessionTokenService) Renew(account *domain.Account) (payload string, err error) {
	token := newSessionToken(account)
	return s.encode(token, account.SessionSalt), nil
}

// Suspend 使指定账号会话失效
func (s *SessionTokenService) Suspend(ctx context.Context, account *domain.Account) error {
	// 通过替换sesion salt达到会话失效的目的
	if err := account.RefreshSessionSalt(); err != nil {
		return fmt.Errorf("refresh session token, %w", err)
	} else if err := s.Accounts.Update(ctx, account); err != nil {
		return fmt.Errorf("save account, %w", err)
	}
	return nil
}

// Retrieve 恢复凭证内的信息
func (s *SessionTokenService) Retrieve(ctx context.Context, payload string) (*domain.Account, SessionToken, error) {
	token, err := s.decode(payload)
	if err != nil {
		return nil, token, fmt.Errorf("decode token, %w", err)
	}

	account, err := s.Accounts.Find(ctx, token.AccountID)
	if err != nil {
		return nil, token, err
	}

	if s.encode(token, account.SessionSalt) != payload {
		return nil, token, errors.New("invalid token signature")
	}
	return account, token, nil
}

// 构造包含签名的token字符串
func (s *SessionTokenService) encode(token SessionToken, salt string) string {
	payload := fmt.Sprintf("%s,%d", token.AccountID, token.Expire)
	signature := s.sign(payload, salt)

	return fmt.Sprintf("%s;%s", payload, signature)
}

func (s *SessionTokenService) decode(payload string) (SessionToken, error) {
	payload, _, ok := strings.Cut(payload, ";")
	if !ok {
		return SessionToken{}, domain.ErrInvalidSessionToken
	}

	id, expire, ok := strings.Cut(payload, ",")
	if !ok {
		return SessionToken{}, domain.ErrInvalidSessionToken
	}

	accountID, err := uuid.Parse(id)
	if err != nil {
		return SessionToken{}, fmt.Errorf("invalid account id, %w", err)
	}

	expireTime, err := strconv.Atoi(expire)
	if err != nil {
		return SessionToken{}, fmt.Errorf("invalid expire time, %w", err)
	}

	return SessionToken{AccountID: accountID, Expire: int64(expireTime)}, nil
}

// 计算数字签名
func (s *SessionTokenService) sign(payload string, salt string) string {
	signature := fmt.Sprintf("%s,%s", payload, salt)
	return fmt.Sprintf("%x", md5.Sum([]byte(signature)))
}

// SessionToken 会话凭证
type SessionToken struct {
	AccountID uuid.UUID
	Expire    int64 // 凭证过期时间
}

// newSessionToken 生成会话凭证
func newSessionToken(account *domain.Account) SessionToken {
	return SessionToken{
		AccountID: account.ID,
		Expire:    time.Now().Add(tokenExpire).Unix(),
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
		return t.Before(time.Now().Add(tokenRenew))
	}
	return false
}

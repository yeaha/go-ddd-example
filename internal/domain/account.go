package domain

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Account 系统账号
type Account struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	PasswordSalt string    `json:"-"`
	SessionSalt  string    `json:"-"`
}

// SetPassword 设置密码
func (a *Account) SetPassword(password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return errors.New("empty password")
	}

	if err := a.refreshPasswordSalt(); err != nil {
		return fmt.Errorf("refresh password salt, %w", err)
	}

	a.Password = newPassword(password, a.PasswordSalt)
	return nil
}

var emailRegexp = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")

// SetEmail 设置email
func (a *Account) SetEmail(email string) error {
	email = NormalizeEmail(email)

	if !emailRegexp.Match([]byte(email)) {
		return errors.New("invalid email")
	}
	a.Email = email

	return nil
}

// ComparePassword 验证密码是否一致
func (a *Account) ComparePassword(password string) bool {
	return password != "" &&
		a.Password == newPassword(password, a.PasswordSalt)
}

// RefreshSessionSalt 更新会话签名盐，更新后同一账号的其它会话会自动失效
func (a *Account) RefreshSessionSalt() error {
	s, err := newSalt(8)
	if err != nil {
		return err
	}
	a.SessionSalt = s
	return nil
}

func (a *Account) refreshPasswordSalt() error {
	s, err := newSalt(8)
	if err != nil {
		return err
	}
	a.PasswordSalt = s
	return nil
}

func newSalt(length int) (string, error) {
	data := make([]byte, length)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", data), nil
}

func newPassword(password string, salt string) string {
	data := append([]byte(password), []byte(salt)...)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// NormalizeEmail 规范化email输入
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

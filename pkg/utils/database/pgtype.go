package database

import (
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	uuid "github.com/satori/go.uuid"
)

// ToUnix 转换为unix timestamp
func ToUnix(t pgtype.Timestamptz) int64 {
	if t.Status == pgtype.Present {
		return t.Time.Unix()
	}
	return 0
}

// SetTimestamptz 设置timestamp with time zone字段值
func SetTimestamptz(dst *pgtype.Timestamptz, t any) error {
	switch v := t.(type) {
	case time.Time:
		if !v.IsZero() {
			return dst.Set(v)
		}
	case int64:
		if v > 0 {
			return dst.Set(time.Unix(v, 0))
		}
	default:
		return fmt.Errorf("cannot set %T", t)
	}

	if dst.Status != pgtype.Null {
		return dst.Set(nil)
	}
	return nil
}

// SetUUID 设置uuid字段值
func SetUUID(dst *pgtype.UUID, src any) error {
	switch v := src.(type) {
	case uuid.UUID:
		if !uuid.Equal(v, uuid.Nil) {
			return dst.Set(v)
		}
	case uuid.NullUUID:
		if v.Valid {
			return dst.Set(v.UUID)
		}
	default:
		return fmt.Errorf("cannot set %T", src)
	}

	if dst.Status != pgtype.Null {
		return dst.Set(nil)
	}
	return nil
}

// SetVarchar 设置varchar字段值
func SetVarchar(dst *pgtype.Varchar, src string) error {
	if src != "" {
		return dst.Set(src)
	} else if dst.Status != pgtype.Null {
		return dst.Set(nil)
	}
	return nil
}

// SetText 设置text字段值
func SetText(dst *pgtype.Text, src string) error {
	if src != "" {
		return dst.Set(src)
	} else if dst.Status != pgtype.Null {
		return dst.Set(nil)
	}
	return nil
}

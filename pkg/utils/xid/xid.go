package xid

import (
	"strings"

	"github.com/google/uuid"
)

// NewUUID 生成全局唯一ID
func NewUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

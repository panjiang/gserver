package comm

import (
	"encoding/binary"
	"time"
)

// 协议格式 [head,body[code,data]]
const (
	bodySizeLen = 4 // 主体大小长度
	codeLen     = 2 // 协议码: uint16 2字节

	headTotalLen = bodySizeLen + codeLen
)

// Message 通信消息
type Message struct {
	Code      uint16
	bodySize  uint32
	Body      []byte
	CreatedAt time.Time
}

// NewMessage 创建包实例，仅头部数据，主体数据后续添加
func NewMessage(b []byte) *Message {
	start := 0
	end := start + codeLen
	code := binary.BigEndian.Uint16(b[start:end])

	start = end
	end = start + bodySizeLen
	bodySize := binary.BigEndian.Uint32(b[start:end])

	return &Message{Code: code, bodySize: bodySize}
}

// Pack 组包
func (m *Message) pack() []byte {
	b := make([]byte, headTotalLen)
	start := 0
	end := start + codeLen
	binary.BigEndian.PutUint16(b[start:end], m.Code)

	start = end
	end = start + bodySizeLen
	binary.BigEndian.PutUint32(b[start:end], uint32(len(m.Body)))

	return append(b, m.Body...)
}

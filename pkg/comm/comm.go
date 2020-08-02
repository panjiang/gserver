package comm

import (
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

type msgProcess func(code uint16, body []byte)

// Comm 协议通信控制
type Comm struct {
	conn        net.Conn
	readTimeout time.Duration
	cache       []byte
	buf         []byte
}

// Send 发送消息
func (c *Comm) Send(code uint16, body []byte) error {
	msg := &Message{Code: code, Body: body}
	_, err := c.conn.Write(msg.pack())
	if err != nil {
		log.Error().Err(err).Msg("send")
	}
	return err
}

// ReadOne 读取一个请求
func (c *Comm) ReadOne() (*Message, error) {
	var msg *Message
	expectSize := headTotalLen
	for {
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))

		n, err := c.conn.Read(c.buf)
		if err != nil {
			return nil, err
		}

		// 未读全，继续读
		// 先读包长，再读包体
		c.cache = append(c.cache, c.buf[:n]...)
		if len(c.cache) < expectSize {
			continue
		}

		if msg == nil {
			msg = NewMessage(c.cache[:headTotalLen])
			expectSize = int(msg.bodySize)
			c.cache = c.cache[headTotalLen:]
		}
		if len(c.cache) < expectSize {
			continue
		}

		// 截取包体
		msg.Body = c.cache[:expectSize]
		msg.CreatedAt = time.Now()

		// 保留剩余数据
		c.cache = c.cache[expectSize:]

		return msg, nil
	}
}

// Close 手动关闭
func (c *Comm) Close() {
	if c.conn == nil {
		return
	}
	c.conn.Close()
	c.conn = nil
}

// NewComm 创建协议通信控制实例
func NewComm(conn net.Conn, readTimeout time.Duration) *Comm {
	return &Comm{
		conn:        conn,
		readTimeout: readTimeout,
		cache:       make([]byte, 0),
		buf:         make([]byte, 1024),
	}
}

package comm

import (
	"net"
	"time"
)

// Comm 协议通信控制
type Comm struct {
	conn     net.Conn
	readChan chan *Message
	quit     chan bool
}

// Msgs 已接收的消息
func (c *Comm) Msgs() <-chan *Message {
	return c.readChan
}

// Send 发送消息
func (c *Comm) Send(code uint16, body []byte) error {
	msg := &Message{Code: code, Body: body}
	_, err := c.conn.Write(msg.Pack())
	return err
}

func (c *Comm) reader() {
	data := make([]byte, 0)
	buf := make([]byte, 1024)

	expectSize := headTotalLen
	var msg *Message
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			break
		}

		// 未读全，继续读
		// 先读包长，再读包体
		data = append(data, buf[:n]...)
		if len(data) < expectSize {
			continue
		}

		if msg == nil {
			msg = NewMessage(data[:headTotalLen])
			expectSize = int(msg.bodySize)
			data = data[headTotalLen:]
		}
		if len(data) < expectSize {
			continue
		}

		// 截取包体
		msg.Body = data[:expectSize]
		msg.CreatedAt = time.Now()
		c.readChan <- msg
		// log.Debug().Uint16("code", msg.Code).Msg("recv")

		// 保留剩余数据
		data = data[expectSize:]

		// 重置包实例
		msg = nil
	}
	close(c.readChan)
	c.quit <- true
}

func (c *Comm) main() {
Loop:
	for {
		select {
		case <-c.quit:
			c.conn.Close()
			break Loop
		}
	}
	close(c.quit)
}

// Close 手动关闭
func (c *Comm) Close() {
	c.conn.Close()
}

// NewComm 创建协议通信控制实例
func NewComm(conn net.Conn) *Comm {
	c := &Comm{
		conn:     conn,
		readChan: make(chan *Message),
		quit:     make(chan bool),
	}
	go c.reader()
	go c.main()

	return c
}

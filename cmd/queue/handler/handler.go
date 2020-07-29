package handler

import (
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/panjiang/gserver/cmd/queue/dao"
	"github.com/panjiang/gserver/cmd/queue/models"
	"github.com/panjiang/gserver/pkg/comm"
	"github.com/panjiang/gserver/pkg/config"
	"github.com/rs/zerolog/log"
)

type hub interface {
	Add(id string, hd *Handler)
	Rem(id string)
}

// Handler 客户端请求控制类
type Handler struct {
	conf *config.Config
	comm *comm.Comm
	user *models.User
	dao  *dao.Dao
	hub  hub
}

// Send 发送数据
func (h *Handler) Send(code uint16, b []byte) {
	h.comm.Send(code, b)
}

func (h *Handler) setUser(user *models.User) {
	h.user = user
	h.hub.Add(user.ID, h)
}

func (h *Handler) process() {
	msgs := h.comm.Msgs()

	ticker := time.NewTicker(time.Second * 10)
Loop:
	for {
		select {
		case <-ticker.C:
			// 连接建立后，10s未发请求，断开连接
			if h.user == nil {
				h.comm.Close()
				break Loop
			}
		case msg, ok := <-msgs:
			if !ok {
				msgs = nil
				break Loop
			}

			process, ok := FindProcess(msg.Code)
			if !ok {
				log.Debug().Uint16("code", msg.Code).Msg("not found")
				break
			}

			resp, err := process(h, msg.Body)
			if err != nil {
				log.Debug().Err(err).Msg("process error")
				break
			}

			out, err := proto.Marshal(resp)
			if err != nil {
				break
			}

			h.comm.Send(msg.Code, out)
			log.Debug().Dur("duration", time.Now().Sub(msg.CreatedAt)).Uint16("code", msg.Code).Str("user", h.user.ID).Send()
		}
	}

	ticker.Stop()
	if h.user != nil {
		h.hub.Rem(h.user.ID)
	}
}

// NewHandler 创建请求请求控制实例
func NewHandler(hub hub, conf *config.Config, dao *dao.Dao, conn net.Conn) *Handler {
	h := &Handler{
		conf: conf,
		comm: comm.NewComm(conn),
		dao:  dao,
		hub:  hub,
	}
	go h.process()
	return h
}

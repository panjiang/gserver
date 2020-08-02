package handler

import (
	"fmt"
	"io"
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
	cm   *comm.Comm
	user *models.User
	dao  *dao.Dao
	hub  hub
}

// Send 发送数据
func (h *Handler) Send(code uint16, b []byte) {
	h.cm.Send(code, b)
}

func (h *Handler) setUser(user *models.User) {
	h.user = user
	h.hub.Add(user.ID, h)
}

func (h *Handler) handleRequest(msg *comm.Message) (err error) {
	process, ok := FindProcess(msg.Code)
	if !ok {
		err = fmt.Errorf("code not found: %d", msg.Code)
		return
	}

	resp, err := process(h, msg.Body)
	if err != nil {
		return
	}

	out, err := proto.Marshal(resp)
	if err != nil {
		return
	}

	if err = h.cm.Send(msg.Code, out); err != nil {
		return
	}

	log.Debug().
		Dur("duration", time.Now().Sub(msg.CreatedAt)).
		Uint16("code", msg.Code).
		Str("user", h.user.ID).Send()

	return
}

func (h *Handler) reader() {
	for {
		msg, err := h.cm.ReadOne()
		if err != nil {
			if err != io.EOF {
				log.Error().Err(err).Msg("ReadOne")
			}
			break
		}
		if err = h.handleRequest(msg); err != nil {
			log.Error().Err(err).Msg("msg")
			break
		}
	}

	if h.user != nil {
		h.hub.Rem(h.user.ID)
	}

	h.cm.Close()
}

// NewHandler 创建请求请求控制实例
func NewHandler(hub hub, conf *config.Config, dao *dao.Dao, conn net.Conn) *Handler {
	h := &Handler{
		conf: conf,
		cm:   comm.NewComm(conn, time.Second*10),
		dao:  dao,
		hub:  hub,
	}
	go h.reader()
	return h
}

package handler

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"github.com/panjiang/gserver/api/queue"
	"github.com/panjiang/gserver/cmd/queue/codes"
	"github.com/panjiang/gserver/cmd/queue/models"
)

func init() {
	register(codes.TokenRequest, requestToken)
}

func requestToken(h *Handler, in []byte) (resp proto.Message, err error) {
	req := &queue.RequestTokenReq{}
	if err = proto.Unmarshal(in, req); err != nil {
		return
	}
	// 请求参数校验
	if req.Id == "" {
		err = errors.New("invalid id")
		return
	}

	// 关联用户
	if h.user == nil {
		h.setUser(&models.User{ID: req.Id})
	} else if h.user.ID != req.Id {
		err = errors.New("id changed")
		return
	}

	// 存在有效期内的token
	token, err := h.dao.TokenGet(req.Id)
	if err != redis.Nil {
		// 异常
		if err != nil {
			return
		}
		// err == nil, 存在
		resp = &queue.RequestTokenResp{
			OldToken: token,
		}
		return
	}

	// 是否已经在请求集合中
	rank, err := h.dao.TokenRequestRank(req.Id)
	if err != nil {
		if err != redis.Nil {
			return
		}
		// 不存在，添加进去
		score := float64(time.Now().UnixNano() / int64(time.Microsecond))
		_, err = h.dao.TokenRequestAdd(req.Id, score)
		if err != nil {
			return
		}

		// 获取所处排名
		rank, err = h.dao.TokenRequestRank(req.Id)
		if err != nil {
			return
		}
	}

	number, seconds := 0, 0
	n := h.conf.Queue.Limit

	// 小于每秒限制，说明不用排队
	if rank > n {
		number = rank - 1
		seconds = number / n
	}

	// 返回前面排队人数
	resp = &queue.RequestTokenResp{
		FrontNumber: int32(number),
		WaitSeconds: int32(seconds),
	}

	return
}

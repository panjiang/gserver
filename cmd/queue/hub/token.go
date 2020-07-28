package hub

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/panjiang/gserver/api/queue"
	"github.com/panjiang/gserver/cmd/queue/codes"
	"github.com/panjiang/gserver/pkg/utils/xid"
	"github.com/rs/zerolog/log"
)

func issueTokenOnce(h *hub) (err error) {
	// 取出前n名请求的id
	conf := h.conf.Queue
	n := conf.Limit - 1
	lifetime := conf.TokenLifetime
	elements, err := h.dao.TokenRequestPeek(n)
	if err != nil {
		return
	}

	log.Debug().Int("n", len(elements)).Int("hds", len(h.hds)).Msg("issueTokenOnce")
	if len(elements) == 0 {
		return
	}

	// 为其生成token
	tokens := map[string]string{}
	for _, ele := range elements {
		token := xid.NewUUID()
		tokens[ele.Member.(string)] = token
	}

	// 批量存储到redis，并设置有效期
	dur := time.Second * time.Duration(lifetime)
	if err = h.dao.TokenRelate(tokens, dur); err != nil {
		return
	}

	// 从请求集合中删除
	maxScore := elements[len(elements)-1].Score
	if err = h.dao.TokenRequestRem(maxScore); err != nil {
		return
	}
	log.Debug().Msg("issueTokenOnce#1")

	// 通知客户端
	for id, token := range tokens {
		push := &queue.RequestTokenPush{
			Token: token,
		}
		b, _ := proto.Marshal(push)
		h.Notice(id, codes.TokenPush, b)
	}
	log.Debug().Msg("issueTokenOnce#2")

	return
}

func (h *hub) issueTokenTask() {
	for {
		err := issueTokenOnce(h)
		if err != nil {
			log.Error().Err(err).Msg("issue token once")
		}
		time.Sleep(time.Second)
	}
}

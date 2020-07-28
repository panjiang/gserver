package handler

import (
	"github.com/golang/protobuf/proto"
	"github.com/rs/zerolog/log"
)

// ProcessFunc 协议处理函数模板
type ProcessFunc func(h *Handler, in []byte) (resp proto.Message, err error)

var processFuncs map[uint16]ProcessFunc

func init() {
	processFuncs = make(map[uint16]ProcessFunc)
}

func register(code uint16, h ProcessFunc) {
	if _, ok := processFuncs[code]; ok {
		log.Panic().Msgf("register code:%d duplicately", code)
	}

	processFuncs[code] = h
}

// FindProcess 按code查找处理函数
func FindProcess(code uint16) (ProcessFunc, bool) {
	h, ok := processFuncs[code]
	return h, ok
}

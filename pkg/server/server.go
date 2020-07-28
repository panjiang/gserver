package server

import (
	"net"

	"github.com/rs/zerolog/log"
)

// TCPHandlerHub .
type TCPHandlerHub interface {
	AcceptConn(conn net.Conn)
}

// TCPServer 主服务
type TCPServer struct {
	addr           string
	handlerFactory TCPHandlerHub
}

// Run 启动服务
func (s *TCPServer) Run() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Err(err).Msg("accept connection")
			continue
		}

		// 控制连接
		s.handlerFactory.AcceptConn(conn)
	}
}

// NewTCPServer 创建新服务
func NewTCPServer(addr string, handlerFactory TCPHandlerHub) *TCPServer {
	server := &TCPServer{
		addr:           addr,
		handlerFactory: handlerFactory,
	}
	return server
}

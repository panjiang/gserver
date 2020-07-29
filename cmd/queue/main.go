package main

import (
	"flag"

	"github.com/panjiang/gserver/cmd/queue/hub"
	"github.com/panjiang/gserver/pkg/config"
	"github.com/panjiang/gserver/pkg/server"
	"github.com/rs/zerolog/log"
)

const name = "QueueService"

func main() {
	flag.Parse()

	// 解析配置文件
	conf, err := config.Parse(name)
	if err != nil {
		log.Fatal().Err(err).Msg("parse config")
	}

	hub, err := hub.NewHub(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("create hub")
	}

	// 启动服务
	log.Info().Str("name", name).Str("addr", conf.Queue.Addr).Msg("run server")
	svr := server.NewTCPServer(conf.Queue.Addr, hub)
	if err := svr.Run(); err != nil {
		log.Fatal().Err(err).Msg("run server")
	}
}

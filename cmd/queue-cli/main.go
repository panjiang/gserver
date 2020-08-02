package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/panjiang/gserver/api/queue"
	"github.com/panjiang/gserver/cmd/queue/codes"
	"github.com/panjiang/gserver/pkg/comm"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Client 客户端类
type Client struct {
	id   string
	cm   *comm.Comm
	wg   *sync.WaitGroup
	done chan struct{}
}

func (c *Client) requestToken(isFirst bool) {
	out, err := proto.Marshal(&queue.RequestTokenReq{
		Id: c.id,
	})
	if err != nil {
		panic(err)
	}

	if err := c.cm.Send(codes.TokenRequest, out); err != nil {
		panic(err)
	}
}

func (c *Client) gotToken(token string, isNew bool) {
	log.Debug().Str("id", c.id).Str("token", token).Bool("isNew", isNew).Msg("got token")
	c.done <- struct{}{}
}

func (c *Client) requestTokenResp(in []byte) {
	var resp queue.RequestTokenResp
	if err := proto.Unmarshal(in, &resp); err != nil {
		panic(err)
	}

	if resp.OldToken != "" {
		c.gotToken(resp.OldToken, false)
		return
	}

	// 需要排队，每秒请求一次状态，直到排到
	if resp.FrontNumber > 0 {
		log.Debug().Str("id", c.id).Int32("frontNumber", resp.FrontNumber).Msg("token request")
		time.Sleep(time.Second)
		c.requestToken(false)
	}
}

func (c *Client) requestTokenPush(in []byte) {
	var resp queue.RequestTokenPush
	if err := proto.Unmarshal(in, &resp); err != nil {
		panic(err)
	}
	c.gotToken(resp.Token, true)
}

func (c *Client) process(msg *comm.Message) {
	switch msg.Code {
	case codes.TokenRequest:
		c.requestTokenResp(msg.Body)
	case codes.TokenPush:
		c.requestTokenPush(msg.Body)
	}
}

func (c *Client) run() {
Loop:
	for {
		select {
		case <-c.done:
			break Loop
		default:
		}

		msg, err := c.cm.ReadOne()
		if err != nil {
			panic(err)
		}
		c.process(msg)
	}

	c.cm.Close()
	c.wg.Done()
}

func newClient(server string, id string, wg *sync.WaitGroup) {
	d := net.Dialer{Timeout: time.Second * 10}
	conn, err := d.Dial("tcp", server)
	if err != nil {
		panic(err)
	}

	cm := comm.NewComm(conn, time.Second*60)
	c := &Client{
		id:   id,
		wg:   wg,
		cm:   cm,
		done: make(chan struct{}, 1),
	}

	// 发起请求
	go func() {
		c.requestToken(true)
	}()

	// 处理接收
	c.run()
}

var server string // 服务器地址
var clientsNumber int

func init() {
	flag.StringVar(&server, "s", "localhost:8080", "server addr")
	flag.IntVar(&clientsNumber, "n", 3000, "number of clients")
}

func main() {
	flag.Parse()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var wg sync.WaitGroup
	for i := 0; i < clientsNumber; i++ {
		wg.Add(1)

		// 客户端唯一ID
		clientID := fmt.Sprintf("%d", i)

		go newClient(server, clientID, &wg)
		time.Sleep(time.Microsecond * 100)
	}

	wg.Wait()

	time.Sleep(time.Second * 1)
}

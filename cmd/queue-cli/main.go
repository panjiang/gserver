package main

import (
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
	ID   string
	comm *comm.Comm
	wg   *sync.WaitGroup
	done chan bool
}

func (c *Client) requestToken() {
	out, err := proto.Marshal(&queue.RequestTokenReq{
		Id: c.ID,
	})
	if err != nil {
		panic(err)
	}

	c.comm.Send(codes.TokenRequest, out)
}

func (c *Client) gotToken(token string, isNew bool) {
	log.Debug().Str("id", c.ID).Str("token", token).Bool("isNew", isNew).Msg("got token")
	c.done <- true
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
		log.Debug().Str("id", c.ID).Int32("frontNumber", resp.FrontNumber).Msg("token request")
		time.Sleep(time.Second)
		c.requestToken()
	}
}

func (c *Client) requestTokenPush(in []byte) {
	var resp queue.RequestTokenPush
	if err := proto.Unmarshal(in, &resp); err != nil {
		panic(err)
	}
	c.gotToken(resp.Token, true)
}

func (c *Client) recv() {
	msgs := c.comm.Msgs()
Loop:
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				msg = nil
				break Loop
			}
			switch msg.Code {
			case codes.TokenRequest:
				c.requestTokenResp(msg.Body)
			case codes.TokenPush:
				c.requestTokenPush(msg.Body)
			}
		case <-c.done:
			break Loop
		}
	}
	c.comm.Close()
	c.wg.Done()
}

func newClient(server string, id string, wg *sync.WaitGroup) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		panic(err)
	}
	comm := comm.NewComm(conn)
	c := &Client{ID: id, comm: comm, wg: wg, done: make(chan bool, 1)}

	// 处理接收
	go c.recv()

	// 发起请求
	c.requestToken()
}

const (
	server        = "localhost:8080" // 服务器地址
	clientsNumber = 2000             // 机器人数量
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	wg := &sync.WaitGroup{}
	for i := 0; i < clientsNumber; i++ {
		wg.Add(1)

		// 客户端唯一ID
		clientID := fmt.Sprintf("%d", i)

		go newClient(server, clientID, wg)
		time.Sleep(time.Millisecond * 10)
	}
	wg.Wait()
}

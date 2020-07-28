package hub

import (
	"net"
	"sync"

	"github.com/panjiang/gserver/cmd/queue/dao"
	"github.com/panjiang/gserver/cmd/queue/handler"
	"github.com/panjiang/gserver/pkg/config"
	"github.com/panjiang/gserver/pkg/server"
)

type hub struct {
	conf *config.Config
	dao  *dao.Dao
	hds  map[string]*handler.Handler
	mu   sync.RWMutex
}

func (h *hub) AcceptConn(conn net.Conn) {
	handler.NewHandler(h, h.conf, h.dao, conn)
}

func (h *hub) Add(id string, hd *handler.Handler) {
	h.mu.Lock()
	h.hds[id] = hd
	h.mu.Unlock()
}

func (h *hub) Rem(id string) {
	h.mu.Lock()
	delete(h.hds, id)
	h.mu.Unlock()
}

func (h *hub) Notice(id string, code uint16, data []byte) {
	h.mu.RLock()
	hd, ok := h.hds[id]
	h.mu.RUnlock()
	if !ok {
		return
	}

	hd.Send(code, data)
}

// NewHub 创建逻辑控制实例中心
func NewHub(conf *config.Config) (server.TCPHandlerHub, error) {
	dao, err := dao.New(conf)
	if err != nil {
		return nil, err
	}

	h := &hub{
		dao:  dao,
		conf: conf,
		hds:  make(map[string]*handler.Handler),
		mu:   sync.RWMutex{},
	}
	go h.issueTokenTask()

	return h, nil
}

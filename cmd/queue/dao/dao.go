package dao

import (
	"github.com/go-redis/redis"
	"github.com/panjiang/gserver/pkg/config"
	"github.com/panjiang/gserver/pkg/redisdb"
)

type Dao struct {
	rdb redis.UniversalClient
	key *redisdb.Keyer
}

// New create instance
func New(conf *config.Config) (*Dao, error) {
	rdb, err := redisdb.Connect(conf.Redis)
	if err != nil {
		return nil, err
	}

	keyer := redisdb.NewKeyer(conf.Name)

	d := &Dao{
		rdb: rdb,
		key: keyer,
	}

	return d, nil
}

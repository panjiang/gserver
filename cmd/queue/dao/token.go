package dao

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/panjiang/gserver/pkg/redisdb"
	"github.com/panjiang/gserver/pkg/utils/xstrconv"
)

// TokenRequestRank 获取所在排名
func (d *Dao) TokenRequestRank(id string) (int, error) {
	key := d.key.Gen(kpTokenRequst)
	rank, err := d.rdb.ZRank(key, id).Result()
	if err != nil {
		return 0, err
	}
	return int(rank) + 1, err
}

// TokenRequestAdd 添加请求到集合
func (d *Dao) TokenRequestAdd(id string, score float64) (int64, error) {
	key := d.key.Gen(kpTokenRequst)
	return d.rdb.ZAdd(key, redis.Z{Score: score, Member: id}).Result()
}

// TokenRequestPeek 取出排名前n的元素
func (d *Dao) TokenRequestPeek(n int) ([]redis.Z, error) {
	key := d.key.Gen(kpTokenRequst)
	stop := int64(n)
	return d.rdb.ZRangeWithScores(key, 0, stop).Result()
}

// TokenRequestRem 删除小于maxScore的元素
func (d *Dao) TokenRequestRem(maxScore float64) error {
	key := d.key.Gen(kpTokenRequst)
	_, err := d.rdb.ZRemRangeByScore(key, "-inf", xstrconv.FormatFloat64(maxScore)).Result()
	return err
}

// TokenRelate 关联token和客户端id
func (d *Dao) TokenRelate(tokens map[string]string, dur time.Duration) error {
	pipe := d.rdb.TxPipeline()
	for clientID, token := range tokens {
		key := redisdb.GlobalKey(kpTokenIssue, clientID)
		pipe.Set(key, token, dur).Result()
	}
	_, err := pipe.Exec()
	return err
}

// TokenGet 获取已生成的token
func (d *Dao) TokenGet(id string) (string, error) {
	key := redisdb.GlobalKey(kpTokenIssue, id)
	return d.rdb.Get(key).Result()
}

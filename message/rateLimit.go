package message

import (
	"crypto/md5"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

type rateLimit struct {
	cache   *redis.Client
	msgSec  time.Duration
	sameSec time.Duration
}

func newRateLimit(cache *redis.Client) *rateLimit {
	return &rateLimit{
		cache:   cache,
		msgSec:  time.Second,
		sameSec: 10 * time.Second,
	}
}

func (r *rateLimit) perSec(mid int64) error {
	key := fmt.Sprintf("rate_%d_%d", mid, time.Now().Unix())
	ex, err := r.cache.SetNX(key, 1, r.msgSec).Result()
	if err != nil {
		return err
	}
	if !ex {
		return errors.ErrRateMsg
	}
	return nil
}

func (r *rateLimit) sameMsg(msg ProducerMessage) error {
	key := fmt.Sprintf("rate_msg_%s", md5.Sum([]byte(msg.Uid+msg.Message)))
	cut, err := r.cache.Incr(key).Result()
	if err != nil {
		return err
	}
	if cut == 1 {
		_, err := r.cache.Expire(key, r.sameSec).Result()
		if err != nil {
			log.Error("set rate same msg for redis", zap.Error(err), zap.Int64("mid", msg.Mid))
			if _, err := r.cache.Del(key).Result(); err != nil {
				log.Error("del rate same msg for redis", zap.Error(err), zap.Int64("mid", msg.Mid))
			}
		}
	}
	if cut >= 3 {
		return errors.ErrRateSameMsg
	}
	return nil
}

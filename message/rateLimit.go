package message

import (
	"crypto/md5"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/ht-co/cpw/micro/log"
	"gitlab.com/ht-co/wowtch/live/alakazam/errors"
	"go.uber.org/zap"
	"time"
)

const (
	rateKey = prefix + ":rate_%d_%d"

	rateMsgKey = prefix + ":rate_msg_%s"
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
	key := fmt.Sprintf(rateKey, mid, time.Now().Unix())
	ex, err := r.cache.SetNX(key, 1, r.msgSec).Result()
	if err != nil {
		return err
	}
	if !ex {
		return errors.ErrRateMsg
	}
	return nil
}

func (r *rateLimit) sameMsg(message string, uid string) error {
	key := fmt.Sprintf(rateMsgKey, md5.Sum([]byte(uid+message)))
	cut, err := r.cache.Incr(key).Result()
	if err != nil {
		return err
	}
	if cut == 1 {
		_, err := r.cache.Expire(key, r.sameSec).Result()
		if err != nil {
			log.Error("set rate same msg for redis", zap.Error(err), zap.String("uid", uid))
			if _, err := r.cache.Del(key).Result(); err != nil {
				log.Error("del rate same msg for redis", zap.Error(err), zap.String("uid", uid))
			}
		}
	}
	if cut >= 3 {
		return errors.ErrRateSameMsg
	}
	return nil
}

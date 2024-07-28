package ioc

import (
	"github.com/MuxiKeStack/be-evaluation/pkg/limiter"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitLimiter(cmd redis.Cmdable) limiter.Limiter {
	// 最近一秒的内的请求不能多余以前
	return limiter.NewRedisSlideWindowLimiter(cmd, time.Second, 1000)
}

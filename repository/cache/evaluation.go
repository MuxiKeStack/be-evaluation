package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/MuxiKeStack/be-evaluation/domain"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"math/rand/v2"
	"strconv"
	"time"
)

var ErrKeyNotExists = redis.Nil

const (
	filedScore    = "score"
	filedRaterCnt = "rater_cnt"
)

var (
	//go:embed lua/composite_score_update_rating.lua
	updateRatingLuaScript string
	//go:embed lua/composite_score_add_rating.lua
	addRatingLuaScript string
	//go:embed lua/composite_score_delete_rating.lua
	deleteRatingLuaScript string
)

type EvaluationCache interface {
	GetCompositeScore(ctx context.Context, courseId int64) (domain.CompositeScore, error)
	SetCompositeScore(ctx context.Context, courseId int64, cs domain.CompositeScore) error
	UpdateRatingIfCompositeScorePresent(ctx context.Context, courseId int64, oldRating uint8, newRating uint8) error
	AddRatingIfCompositeScorePresent(ctx context.Context, courseId int64, starRating uint8) error
	DeleteRatingIfCompositeScorePresent(ctx context.Context, courseId int64, starRating uint8) error
}

type RedisEvaluationCache struct {
	cmd redis.Cmdable
	g   singleflight.Group
}

func NewRedisEvaluationCache(cmd redis.Cmdable) EvaluationCache {
	return &RedisEvaluationCache{cmd: cmd}
}

func (cache *RedisEvaluationCache) GetCompositeScore(ctx context.Context, courseId int64) (domain.CompositeScore, error) {
	key := cache.compositeScoreKey(courseId)
	data, err := cache.cmd.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.CompositeScore{}, err
	}
	if len(data) == 0 {
		return domain.CompositeScore{}, ErrKeyNotExists
	}
	score, _ := strconv.ParseFloat(data[filedScore], 64)
	raterCnt, _ := strconv.ParseInt(data[filedRaterCnt], 10, 64)
	return domain.CompositeScore{
		CourseId: courseId,
		Score:    score,
		RaterCnt: raterCnt,
	}, nil
}

func (cache *RedisEvaluationCache) SetCompositeScore(ctx context.Context, courseId int64, cs domain.CompositeScore) error {
	key := cache.compositeScoreKey(courseId)
	// 使用singleflight, 防止缓存击穿
	_, err, _ := cache.g.Do(key, func() (interface{}, error) {
		err := cache.cmd.HSet(ctx, key, filedScore, cs.Score, filedRaterCnt, cs.RaterCnt).Err()
		if err != nil {
			return nil, err
		}
		n := rand.IntN(181) // 随机偏移的秒数[0, 180]，防止缓存雪崩
		return nil, cache.cmd.Expire(ctx, key, time.Minute*15+time.Second*time.Duration(n)).Err()
	})
	return err
}

func (cache *RedisEvaluationCache) UpdateRatingIfCompositeScorePresent(ctx context.Context, courseId int64, oldRating uint8, newRating uint8) error {
	key := cache.compositeScoreKey(courseId)
	return cache.cmd.Eval(ctx, updateRatingLuaScript, []string{key}, oldRating, newRating).Err()
}

func (cache *RedisEvaluationCache) AddRatingIfCompositeScorePresent(ctx context.Context, courseId int64, starRating uint8) error {
	key := cache.compositeScoreKey(courseId)
	return cache.cmd.Eval(ctx, addRatingLuaScript, []string{key}, starRating).Err()
}

func (cache *RedisEvaluationCache) DeleteRatingIfCompositeScorePresent(ctx context.Context, courseId int64, starRating uint8) error {
	key := cache.compositeScoreKey(courseId)
	return cache.cmd.Eval(ctx, deleteRatingLuaScript, []string{key}, starRating).Err()
}

func (cache *RedisEvaluationCache) compositeScoreKey(courseId int64) string {
	return fmt.Sprintf("kstack:evaluation:composite_score:%d", courseId)
}

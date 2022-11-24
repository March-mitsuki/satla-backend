package rdbtools

import (
	"context"

	"github.com/go-redis/redis/v9"
)

type RdbOk int

const (
	NoKey RdbOk = iota
	OrgErr
	Empty
	Ok
)

// 一个get redis时的便利函数, 目前还没用到
func GetValueAndCheck(rdb *redis.Client, ctx context.Context, key string) (string, RdbOk) {
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return val, NoKey
	} else if err != nil {
		return val, OrgErr
	} else if val == "" {
		return val, Empty
	} else {
		return val, Ok
	}
}

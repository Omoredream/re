package Utils

import (
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

func RedisHPExpire(ctx g.Ctx, client *gredis.Redis, key string, ttlMS int64, fields ...string) (err error) {
	v, err := client.Do(ctx, "HPExpire", append([]any{key, ttlMS, "FIELDS", len(fields)}, gconv.Interfaces(fields)...)...)
	if err != nil {
		return
	}

	results := v.Int64s()
	for i, result := range results {
		switch result {
		case -2:
			err = gerror.Newf("提供的哈希键 %s 不存在, 或提供的哈希键中不存在 %s", key, fields[i])
			return
		case 0:
			err = gerror.Newf("未满足指定的 TTL 条件")
			return
		case 1:
			// 正常
		case 2:
			// 删除
		}
	}

	return
}

func RedisHPTtl(ctx g.Ctx, client *gredis.Redis, key string, fields ...string) (TTLs []int64, err error) {
	v, err := client.Do(ctx, "HPTtl", append([]any{key, "FIELDS", len(fields)}, gconv.Interfaces(fields)...)...)
	if err != nil {
		return
	}

	results := v.Int64s()
	TTLs = make([]int64, len(results))
	for i, result := range results {
		switch result {
		case -2:
			// err = gerror.Newf("提供的哈希键 %s 不存在, 或提供的哈希键中不存在 %s", key, fields[i])
			TTLs[i] = 0
		case -1:
			// 永不过期
			TTLs[i] = -1
		default:
			// 正常
			TTLs[i] = result
		}
	}

	return
}

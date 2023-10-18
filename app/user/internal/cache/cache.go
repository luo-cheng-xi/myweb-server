package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"math/rand"
	"myweb/app/user/internal/conf"
	"strconv"
	"time"
)

const LockKey = "lock:key:"

type Cache struct {
	Cli    *redis.Client
	logger *zap.Logger
}

type RedisData struct {
	Value      any
	Expiration int64 //过期时间
}

func NewCache(config *conf.Config, logger *zap.Logger) *Cache {
	logger.Debug("cache init",
		zap.String("Addr", config.RedisConf.Addr),
		zap.String("Password", config.RedisConf.Password),
		zap.String("DB", strconv.Itoa(config.RedisConf.DB)))
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisConf.Addr,
		Password: config.RedisConf.Password,
		DB:       config.RedisConf.DB,
	})
	return &Cache{
		Cli:    client,
		logger: logger,
	}
}

// Set 向Redis存入一个键值对
func (c Cache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	str, _ := json.Marshal(value)
	err := c.Cli.Set(ctx, key, str, expiration).Err()
	if err != nil {
		c.logger.Error("缓存存储出错", zap.String("cause", err.Error()))
		return err
	}
	return nil
}

func (c Cache) SetWithExtraRandExpr(ctx context.Context, key string, value any, expiration time.Duration) error {
	if expiration != 0 {
		expiration += time.Duration(rand.Int()%3600) * time.Second
		c.logger.Debug("计算出的新expiration = ", zap.String("", expiration.String()))
	}
	err := c.Set(ctx, key, value, expiration)
	if err != nil {
		return err
	}
	return nil
}

func (c Cache) SetWithLogicalExpire(ctx context.Context, key string, value any, expiration int64) {
	//设置逻辑过期
	rdata := RedisData{
		Value:      value,
		Expiration: expiration,
	}
	//写入Redis
	c.Cli.Set(ctx, key, rdata, 0)
}

// Get tar 要传引用。返回true表示找到了，其他情况为false
func (c Cache) Get(ctx context.Context, key string, tar any) (bool, error) {
	ret, err := c.Cli.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		c.logger.Error("缓存读取出错", zap.String("cause", err.Error()))
		return false, err
	}
	err = json.Unmarshal([]byte(ret), &tar)
	if err != nil {
		c.logger.Error("反序列化出现错误", zap.String("cause", err.Error()))
		return false, err
	}
	return true, nil
}

func (c Cache) Exist(ctx context.Context, key string) (bool, error) {
	ret, err := c.Cli.Exists(ctx, key).Result()
	if err != nil {
		c.logger.Error("查询key出错", zap.String("cause", err.Error()))
		return false, err
	}
	return ret == 1, nil
}

func (c Cache) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		err := c.Cli.Del(ctx, key).Err()
		if err != nil && err != redis.Nil {
			return err
		}
	}

	return nil
}

// QueryWithMutex 包含互斥锁机制的获取获取方法
// Return:返回false则需要进行数据库操作，否则不需要
// Param:ctx context.Context类型，key 待查询value的key,tar 用于接收缓存数据，需要加`&`
func (c Cache) QueryWithMutex(ctx context.Context, key string, tar any) (bool, error) {
	// 从redis中获取信息
	get := c.Cli.Get(ctx, key)
	err := get.Err()
	val := get.Val()
	// 找到了缓存信息
	if err != redis.Nil {
		// 为空信息
		if val == "" {
			return true, errors.New("查询信息为空，不需要继续进行查询了")
		}
		// 不为空，返回
		err := json.Unmarshal([]byte(val), &tar)
		if err != nil {
			c.logger.Debug("信息解析出错")
			return true, err
		}
		return true, nil
	}
	// 没有找到缓存信息
	// 获取互斥锁
	lockKey := LockKey + key
	// 获取到了则进行数据库操作
	if c.tryLock(ctx, lockKey) {
		c.logger.Debug("取得锁，准许进行数据库操作")
		return false, nil
	}
	// 未获取到则继续执行
	time.Sleep(50 * time.Millisecond)
	c.logger.Debug("未持有锁，等待")
	return c.QueryWithMutex(ctx, key, &tar)
}

// TryLock 尝试拿锁，如果拿到锁就是true,没拿到就是false
func (c Cache) tryLock(ctx context.Context, key string) bool {
	flag := c.Cli.SetNX(ctx, key, "1", 3*time.Second)
	return flag.Val()
}

func (c Cache) unlock(ctx context.Context, key string) {
	c.Cli.Del(ctx, key)
}

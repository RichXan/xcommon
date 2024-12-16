package xcache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/richxan/xpkg/xlog"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type RedisConfig struct {
	Addresses    []string      `yaml:"addresses"` //非哨兵模式时，配置多个地址表示集群模式
	Password     string        `yaml:"password"`
	MasterName   string        `yaml:"master"` //有配置master表示哨兵模式
	Db           int           `yaml:"db"`
	IsConsole    bool          `yaml:"is_console"` //是否打印日志
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	MaxRetries   int           `yaml:"max_retries"`
}

type RedisClient struct {
	rdb     redis.UniversalClient
	logger  xlog.Logger
	redSync *redsync.Redsync // 分布式锁对象
	mutex   sync.Mutex
}

func NewRedisClient(masterName string, addresses []string, password string, logger *xlog.Logger) (*RedisClient, error) {
	gRedisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        addresses,
		DB:           0,
		Password:     password,
		MaxRetries:   3,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		PoolSize:     20,
		MinIdleConns: 10,
		MasterName:   masterName,
	})
	if err := gRedisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("init redis error and error is %v", err)
	}
	c := &RedisClient{
		rdb:    gRedisClient,
		logger: *logger,
	}
	return c, nil
}

func NewRedisClientByConfig(config *RedisConfig, logger *xlog.Logger) (*RedisClient, error) {
	addresses := config.Addresses
	dialTimeout := time.Second * 5
	readTimeout := time.Second * 15
	writeTimeout := time.Second * 15
	poolSize := 20
	minIdleConns := 10
	maxRetries := 3
	if config.DialTimeout > 0 {
		dialTimeout = config.DialTimeout
	}
	if config.ReadTimeout > 0 {
		readTimeout = config.ReadTimeout
	}
	if config.WriteTimeout > 0 {
		writeTimeout = config.WriteTimeout
	}
	if config.PoolSize > 0 {
		poolSize = config.PoolSize
	}
	if config.MinIdleConns > 0 {
		minIdleConns = config.MinIdleConns
	}
	if config.MaxRetries > 0 {
		maxRetries = config.MaxRetries
	}
	gRedisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        addresses,
		DB:           config.Db,
		Password:     config.Password,
		MaxRetries:   maxRetries,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MasterName:   config.MasterName,
	})
	if err := gRedisClient.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("init redis error and error is %v", err)
	}
	if logger == nil {
		logger = xlog.NewLogger(xlog.LoggerConfig{
			Level:       zerolog.LevelInfoValue,
			Directory:   "./logs",
			ProjectName: "xproject",
			LoggerName:  "redis",
			MaxSize:     100,
			MaxBackups:  10,

			SaveLoggerAsFile: true,
		})
	}
	instance := &RedisClient{
		rdb:    gRedisClient,
		logger: *logger,
	}
	instance.newRedisSync(gRedisClient)
	return instance, nil
}

func (r *RedisClient) Client() redis.UniversalClient {
	return r.rdb
}

// 创建分布式锁对象
func (r *RedisClient) newRedisSync(redisClient redis.UniversalClient) *redsync.Redsync {
	if r.redSync == nil {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		if redisClient == nil {
			panic("redis client is nil")
		}
		r.redSync = redsync.New(goredis.NewPool(redisClient))
	}
	return r.redSync
}

// 创建一个影子redis对象，用于分布式锁
func (r *RedisClient) Shadow(logger *xlog.Logger) *RedisClient {
	if r.rdb == nil {
		panic("redis client is nil")
	}
	instance := &RedisClient{r.rdb, *logger, nil, sync.Mutex{}}
	instance.newRedisSync(r.rdb)
	return instance
}

func (r *RedisClient) Set(k, v string, expiration, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	st := time.Now()
	err := r.rdb.Set(ctx, k, v, expiration).Err()
	r.logger.Info().Str("key", k).Str("value", v).Any("error", err).Int("cost(ms)", int(time.Since(st).Milliseconds())).Msg("set redis finish")
	return err
}

func (r *RedisClient) Get(k string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	st := time.Now()
	v, e := r.rdb.Get(ctx, k).Bytes()
	vs := ""
	if v != nil {
		vs = string(v)
	}
	r.logger.Info().Str("key", k).Any("value", vs).Any("error", e).Int("cost(ms)", int(time.Since(st).Milliseconds())).Msg("get redis finish")
	return v, e
}

func (r *RedisClient) Exists(k string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	st := time.Now()
	v, e := r.rdb.Exists(ctx, k).Result()
	r.logger.Info().Str("key", k).Int64("Exists", v).Any("error", e).Int("cost(ms)", int(time.Since(st).Milliseconds())).Msg("key exists finish")
	return v > 0
}

func (r *RedisClient) Delete(k string, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	st := time.Now()
	v, e := r.rdb.Del(ctx, k).Result()
	r.logger.Info().Str("key", k).Int64("Deleted", v).Any("error", e).Int("cost(ms)", int(time.Since(st).Milliseconds())).Msg("key exists finish")
	return v > 0, e
}

// 竞争的key，成功后执行fn，fn执行完毕后释放锁
func (r *RedisClient) RedLockFunc(key string, fn func() error, options ...redsync.Option) (err error) {
	s := time.Now().UnixMilli()
	redSyncMutex := r.redSync.NewMutex(key, options...)
	if err = redSyncMutex.Lock(); err != nil {
		r.logger.Error().Msgf("redlock %s error %v", key, err)
		return errors.New("busy business, please try again later")
	}
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			r.logger.Error().Msgf("redlock %s error %v", key, recoverErr)
		}
		exists := r.Exists(key, time.Second*30)
		if exists { //存在key才释放，防止出现超时key失效异常
			if ok, err1 := redSyncMutex.Unlock(); !ok || err1 != nil {
				r.logger.Error().Msgf("redlock %s unlock error %v", key, err1)
				err = err1
			} else {
				r.logger.Info().Msgf("redlock %s unlock cost %d", key, time.Now().UnixMilli()-s)
			}
		}
	}()
	err = fn()
	if err != nil {
		r.logger.Error().Msgf("redlock %s business error %v", key, err)
	}
	return err
}

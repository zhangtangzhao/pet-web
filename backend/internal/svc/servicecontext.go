package svc

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"pet/backend/internal/config"
	"pet/backend/internal/model"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Rdb    *redis.Client

	payMtx    sync.Mutex
	payClient *core.Client
	payErr    error
	payReady  bool
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := model.NewDB(c.Database.DataSource)
	if err != nil {
		logx.Must(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Pass,
		DB:       c.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logx.Errorf("redis ping failed: %v", err)
	}
	return &ServiceContext{
		Config: c,
		DB:     db,
		Rdb:    rdb,
	}
}

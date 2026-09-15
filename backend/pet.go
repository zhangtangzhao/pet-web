package main

import (
	"context"
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"pet/backend/internal/common"
	"pet/backend/internal/config"
	"pet/backend/internal/handler"
	"pet/backend/internal/logic/notify"
	"pet/backend/internal/logic/trade"
	"pet/backend/internal/svc"
)

var configFile = flag.String("f", "etc/pet-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv()) // 支持 yaml 中 ${ENV_VAR} 注入
	if err := c.Validate(); err != nil {
		logx.Must(err)
	}

	if err := common.InitIDNode(c.Snowflake.Node); err != nil {
		logx.Must(err)
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 超时未支付订单自动关闭（每分钟扫描）
	go trade.StartOrderCloser(ctx)

	// 微信通知投递器（客服回复离线提醒 + 订单事件，每 10s 扫描）
	go notify.StartNotifier(context.Background(), ctx)

	logx.Infof("pet-api 启动于 %s:%d (mode=%s)", c.Host, c.Port, c.Mode)
	server.Start()
}

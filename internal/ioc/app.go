package ioc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/JrMarcco/easy-grpc/registry"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	AppFxOpt    = fx.Provide(InitApp)
	AppFxInvoke = fx.Invoke(AppLifecycle)
)

// App 应用整体的封装，组合了 grpc.Server
type App struct {
	*grpc.Server

	timeout         time.Duration
	registry        registry.Registry
	serviceInstance registry.ServiceInstance

	logger *zap.Logger
}

func (app *App) Start() error {
	ln, err := net.Listen("tcp", ":50501")
	if err != nil {
		return err
	}

	go func() {
		if serveErr := app.Serve(ln); serveErr != nil {
			panic(serveErr)
		}
		app.logger.Info("[kuryr] successfully started grpc server")
	}()

	if app.registry != nil {
		// 注册服务到注册中心
		registerCtx, cancel := context.WithTimeout(context.Background(), app.timeout)
		registerErr := app.registry.Register(registerCtx, app.serviceInstance)
		cancel()

		if registerErr != nil {
			return fmt.Errorf("[kuryr] failed to register service instance: %w", registerErr)
		}
		app.logger.Info("[kuryr] successfully registered service instance]")
	}
	return nil
}

func (app *App) Stop() error {
	if app.registry != nil {
		// 从注册中心注销服务
		unregisterCtx, cancel := context.WithTimeout(context.Background(), app.timeout)
		err := app.registry.Unregister(unregisterCtx, app.serviceInstance)
		cancel()

		if err != nil {
			app.logger.Error("[kuryr] failed to unregister service instance", zap.Error(err))
		}
		app.logger.Info("[kuryr] successfully unregistered service instance]")

		_ = app.registry.Close()
	}

	// 优雅推出
	app.logger.Info("[kuryr] gracefully stopping grpc server ...")
	app.GracefulStop()
	app.logger.Info("[kuryr] grpc server stopped")
	return nil
}

// InitApp 初始化 app
func InitApp(grpcServer *grpc.Server, r registry.Registry, logger *zap.Logger) *App {
	type config struct {
		Name        string `mapstructure:"name"`
		Addr        string `mapstructure:"addr"`
		Group       string `mapstructure:"group"`
		Timeout     int    `mapstructure:"timeout"`
		ReadWeight  int32  `mapstructure:"read_weight"`
		WriteWeight int32  `mapstructure:"write_weight"`
	}

	cfg := config{}
	if err := viper.UnmarshalKey("app", &cfg); err != nil {
		panic(err)
	}

	si := registry.ServiceInstance{
		Name:        cfg.Name,
		Addr:        cfg.Addr,
		Group:       cfg.Group,
		ReadWeight:  uint32(cfg.ReadWeight),
		WriteWeight: uint32(cfg.WriteWeight),
	}

	return &App{
		Server:          grpcServer,
		timeout:         time.Duration(cfg.Timeout) * time.Millisecond,
		registry:        r,
		serviceInstance: si,
		logger:          logger,
	}
}

// AppLifecycle app 生命周期 hook
func AppLifecycle(lc fx.Lifecycle, app *App) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return app.Start()
		},
		OnStop: func(ctx context.Context) error {
			return app.Stop()
		},
	})
}

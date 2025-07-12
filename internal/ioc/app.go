package ioc

import (
	"context"
	"net"

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
	}()
	return nil
}

func (app *App) Stop() error {
	// 优雅推出
	app.GracefulStop()
	return nil
}

// InitApp 初始化 app
func InitApp(grpcServer *grpc.Server) *App {
	return &App{
		Server: grpcServer,
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

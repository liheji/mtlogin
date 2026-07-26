package app

import (
	"context"
	"errors"
	"mtlogin/internal/domain"
	"mtlogin/internal/refresh"
	"mtlogin/internal/runtime"
	"mtlogin/internal/scheduler"
	"mtlogin/internal/store"
	"mtlogin/pkg/apicore"
	"mtlogin/pkg/apikit/apimtauth"
	"mtlogin/pkg/apikit/apimtbrowse"
	"mtlogin/pkg/apikit/apinotify"
	"mtlogin/pkg/conf"
	"mtlogin/pkg/log"
	"mtlogin/pkg/util"
	"sync"
)

// App 负责启动期依赖装配和进程退出时的资源关闭。
//
// 所有依赖只在进程启动时构建一次；配置变更需要重启进程后生效。
type App struct {
	store     *store.LevelDBStore
	runtime   *runtime.Runtime
	scheduler *scheduler.Scheduler
	notify    *apinotify.Manager
	transport closeableClient
	closeOnce sync.Once
	closeErr  error
}

type closeableClient interface {
	Close() error
}

// New 装配应用依赖，并从固定 conf/app.toml / conf/notify.toml 加载配置。
func New(ctx context.Context) (*App, error) {
	cfg, err := conf.LoadApp(conf.AppConfigPath)
	if err != nil {
		return nil, mapConfigError(err)
	}
	authCfg, browseCfg, err := buildMTeamConfigs(cfg.MT)
	if err != nil {
		return nil, err
	}
	notifyCfg := apinotify.Config{}
	if loaded, err := apinotify.LoadConfig(apinotify.NotifyConfigPath); err != nil {
		log.Warnf("load notify config failed; only base notifier enabled: %v", err)
	} else {
		notifyCfg = *loaded
	}
	store, err := store.OpenDefault()
	if err != nil {
		return nil, err
	}

	transport, err := apicore.NewTLSClient()
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	authClient, err := apimtauth.NewClient(authCfg, transport)
	if err != nil {
		_ = transport.Close()
		_ = store.Close()
		return nil, err
	}
	browseClient, err := apimtbrowse.NewClient(browseCfg, transport)
	if err != nil {
		_ = transport.Close()
		_ = store.Close()
		return nil, err
	}
	notifyOptions := []apinotify.ManagerOption{
		apinotify.SystemProxy(cfg.System.Proxy),
	}
	nm := apinotify.NewManager(notifyCfg, transport, notifyOptions...)
	refreshClient := browseClientAdapter{client: browseClient}
	refreshCfg := refresh.Config{
		Timeout: util.SecondsDuration(cfg.Refresh.Timeout),
	}
	authManager := refresh.NewAuthManager(refresh.AuthConfig{
		PasswordAuth: refresh.PasswordAuthConfig{
			Username: authCfg.PasswordAuth.Username, Password: authCfg.PasswordAuth.Password, TotpSecret: authCfg.PasswordAuth.TotpSecret,
		},
		TokenAuth: domain.AuthState{
			Token: authCfg.TokenAuth.Auth, DID: authCfg.TokenAuth.DID, VisitorID: authCfg.TokenAuth.VisitorID,
		},
		RetryOnAuthFailure: cfg.Refresh.RetryOnAuthFailure,
	}, authClientAdapter{client: authClient}, store)
	svc := refresh.NewService(refreshCfg, refreshClient, authManager, store, notifyAdapter{manager: nm})
	rt := runtime.New(ctx)
	rt.Publish(svc)
	sched, err := scheduler.New(scheduler.Config{
		Crontab:  cfg.App.Crontab,
		MinDelay: util.SecondsDuration(cfg.App.MinDelay),
		MaxDelay: util.SecondsDuration(cfg.App.MaxDelay),
	}, rt)
	if err != nil {
		_ = nm.Close()
		_ = transport.Close()
		_ = store.Close()
		return nil, err
	}
	a := &App{store: store, runtime: rt, scheduler: sched, notify: nm, transport: transport}
	log.Info("app assembled successfully")
	return a, nil
}

func (a *App) RunOnce(ctx context.Context) error {
	return a.runtime.TryRun(ctx)
}

func (a *App) Start(ctx context.Context) error {
	if err := a.scheduler.Start(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}
	a.closeOnce.Do(func() {
		if a.runtime != nil {
			a.runtime.Stop()
		}
		if a.scheduler != nil {
			a.scheduler.Stop()
		}
		var errs []error
		if a.notify != nil {
			errs = append(errs, a.notify.Close())
		}
		if a.transport != nil {
			errs = append(errs, a.transport.Close())
		}
		if a.store != nil {
			errs = append(errs, a.store.Close())
		}
		a.closeErr = errors.Join(errs...)
	})
	return a.closeErr
}

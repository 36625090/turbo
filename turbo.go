package turbo

import (
	"github.com/36625090/turbo/authorities"
	"github.com/36625090/turbo/config"
	"github.com/36625090/turbo/logging"
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/option"
	"github.com/36625090/turbo/server"
	"github.com/36625090/turbo/transport"
	"github.com/36625090/turbo/utils"
	"github.com/go-various/consul"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	_ "runtime/pprof"
)

type Turbo interface {

	//Initialize 服务初始化
	Initialize(func(*server.TurboContext)) error

	//InitializeBackend 注册后端逻辑端点
	InitializeBackend(string, logical.Factory, *logical.BackendContext) error

	//InitializeAuthorization 注册自定义验证接口
	InitializeAuthorization(authorities.Authorization) error

	//InitializeSigner 注册自定义加签验签
	InitializeSigner(transport.Signer)

	Start() error
	Stop()
}

//Default default
func Default(opts *option.Options, factories map[string]logical.Factory) (Turbo, error) {
	if opts.Pprof {
		go func() {
			log.Println(http.ListenAndServe(opts.PprofAddr, nil))
		}()
	}

	logger, err := logging.NewLogger(opts.App, opts.Log)
	if err != nil {
		return nil, err
	}

	globalConfig := &config.GlobalConfig{}

	if err := initializeLocalConfig(opts, globalConfig); err != nil {
		return nil, err
	}

	logger.Trace("initialize config", "config", utils.JSONPrettyDump(globalConfig))

	authorization, err := initializeAuthorization(opts.App, globalConfig, err)
	if err != nil {
		return nil, err
	}

	var client consul.Client
	if opts.UseConsul {
		client, err = initializeConsul(opts, logger)
		if nil != err {
			return nil, err
		}
		if err := initCentralConfig(client, globalConfig); err != nil {
			return nil, err
		}
	}

	inv := server.NewServer(opts, globalConfig, client, logger)

	if err := inv.InitializeAuthorization(authorization); err != nil {
		return nil, err
	}

	context := &logical.BackendContext{
		Logger:       logger,
		Application:  opts.App,
		XormConfig:   globalConfig.XormConfig,
		RedisConfig:  globalConfig.RedisConfig,
		AuthSettings: globalConfig.Authorization,
		TokenHandler: authorization.TokenHandler(),
		Consul:       client,
	}

	for name, factory := range factories {
		if err := inv.InitializeBackend(name, factory, context); err != nil {
			return nil, err
		}
	}
	return inv, nil
}

//initializeLocalConfig 加载本地配置
func initializeLocalConfig(opts *option.Options, globalConfig *config.GlobalConfig) error {
	if opts.ConfigFile != "" {
		bs, err := ioutil.ReadFile(opts.ConfigFile)
		if err != nil {
			return err
		}
		if err := hcl.Unmarshal(bs, globalConfig); err != nil {
			return err
		}
	}
	return nil
}

//initializeConsul 初始化consul
func initializeConsul(opts *option.Options, logger hclog.Logger) (consul.Client, error) {
	return consul.NewClient(opts.ConsulConfig(), logger)
}

//initCentralConfig 加载中心化配置
func initCentralConfig(client consul.Client, globalConfig *config.GlobalConfig) error {
	return client.LoadConfig(globalConfig)
}

//initializeAuthorization 初始化验证模块
func initializeAuthorization(app string, globalConfig *config.GlobalConfig, err error) (authorities.Authorization, error) {
	var tokenHandler authorities.TokenHandler

	if globalConfig.Authorization == nil {
		tokenHandler, _ = authorities.NewNoopTokenHandler()
		return authorities.NewAuthorization(&authorities.Settings{}, tokenHandler)
	}

	switch globalConfig.Authorization.AuthType {

	case authorities.AuthTypeJwt:
		tokenHandler, err = authorities.NewJwtTokenHandler(app, globalConfig.Authorization)
		if err != nil {
			return nil, err
		}
		return authorities.NewAuthorization(globalConfig.Authorization, tokenHandler)

	case authorities.AuthTypeRedis:
		tokenHandler, err = authorities.NewRedisTokenHandler(app, globalConfig.Authorization, globalConfig.RedisConfig)
		if err != nil {
			return nil, err
		}
		return authorities.NewAuthorization(globalConfig.Authorization, tokenHandler)

	default:
		tokenHandler, _ = authorities.NewNoopTokenHandler()
		return authorities.NewAuthorization(&authorities.Settings{}, tokenHandler)
	}

}

package turbo

import (
	"errors"
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
	Initialize() error

	//AddHandle 添加http方法
	AddHandle(path string, method logical.HttpMethod, handle func(*transport.Context, *server.HandlerParams) error)

	//RegisterBackend 注册后端逻辑端点
	RegisterBackend(string, logical.Factory, *logical.BackendContext) error

	//RegisterAuthorization 注册验证接口，如微注册则不验证
	RegisterAuthorization(authorization authorities.Authorization) error

	//Start 启动服务
	Start() error
	Stop()
	//AddLoggerSinks register sinks
	AddLoggerSinks(sinks ...hclog.SinkAdapter)
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

	authorization, err := initializeAuthorization(globalConfig, err)
	if err != nil {
		return nil, err
	}

	context := &logical.BackendContext{
		Logger:       logger,
		Application:  opts.App,
		XormConfig:   globalConfig.XormConfig,
		RedisConfig:  globalConfig.RedisConfig,
		AuthSettings: globalConfig.Authorization,
		TokenHandler: authorization.TokenHandler(),
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
	if err := inv.RegisterAuthorization(authorization); err != nil {
		return nil, err
	}

	for name, factory := range factories {
		if err := inv.RegisterBackend(name, factory, context); err != nil {
			return nil, err
		}
	}

	if err := inv.Initialize(); err != nil {
		return nil, err
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

func initializeAuthorization(globalConfig *config.GlobalConfig, err error) (authorities.Authorization, error) {
	var tokenHandler authorities.TokenHandler
	if globalConfig.Authorization.AuthType == authorities.AuthTypeJwt || globalConfig.Authorization.AuthType == "" {
		tokenHandler, err = authorities.NewJwtTokenHandler(globalConfig.Authorization)
	} else {
		tokenHandler, err = authorities.NewRedisTokenHandler(globalConfig.Authorization, globalConfig.RedisConfig)
	}
	if err != nil {
		return nil, err
	}

	authorization, err := authorities.NewAuthorization(globalConfig.Authorization, tokenHandler)
	if err != nil {
		return nil, errors.New("initialization authorization: " + err.Error())
	}
	return authorization, nil
}

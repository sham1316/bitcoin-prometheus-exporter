package config

import (
	"flag"
	"github.com/btcsuite/btcd/rpcclient"
	configParser "github.com/sham1316/configparser"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"sync"
)

var config *Config
var once sync.Once

var configPath *string

func init() {
	configPath = flag.String("config", "config.yml", "Configuration file path")
	flag.Parse()
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.DisableStacktrace = true
	zapLogger, _ := zapCfg.Build()
	zap.ReplaceGlobals(zapLogger)
	defer zapLogger.Sync() // flushes buffer, if any
}

func GetInstance() *Config {
	once.Do(func() {
		config = loadConfig(configPath)
		config.Logger = zap.S()
		config.Logger.Debugf("%+v\n", config)
	})
	return config
}

type Config struct {
	Logger     *zap.SugaredLogger
	BtcUser    string `default:"alice" env:"BTC_USER"`
	BtcPass    string `default:"DONT_USE_THIS_YOU_WILL_GET_ROBBED_8ak1gI25KFTvjovL3gAM967mies3E=" env:"BTC_PASS"`
	BtcUrl     string `default:"localhost:8332" env:"BTC_URL"`
	DisableTLS bool   `default:"false" env:"BTC_SSL"`
	Metrics    struct {
		MetricsHost   string `default:":8334" env:"METRICS_HOST"`
		MetricsPath   string `default:"/metrics" env:"METRICS_PATH"`
		MetricsPrefix string `default:"bitcoin" env:"METRICS_PREFIX"`
	}
	RpcConfig *rpcclient.ConnConfig
}

func loadConfig(configFile *string) *Config {
	config := Config{}
	_ = configParser.SetValue(&config, "default")
	confYamlFile, _ := ioutil.ReadFile(*configFile)
	_ = yaml.Unmarshal(confYamlFile, &config)
	_ = configParser.SetValue(&config, "env")
	config.RpcConfig = &rpcclient.ConnConfig{
		Host:         config.BtcUrl,
		User:         config.BtcUser,
		Pass:         config.BtcPass,
		DisableTLS:   config.DisableTLS,
		HTTPPostMode: true,
	}
	return &config
}

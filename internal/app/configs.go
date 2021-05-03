package app

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/mammadmodi/webis/internal/api/websocket"
	"github.com/mammadmodi/webis/pkg/logger"
	"github.com/mammadmodi/webis/pkg/redis"
	"io/ioutil"
	"time"
)

// Configs is struct that contains all configuration of all parts of application
type Configs struct {
	SockHubConfig       websocket.Configuration
	RedisConfigs        redis.Configs
	LoggingConfigs      logger.Configuration
	AuthMethod          string        `default:"jwt" split_words:"true"`
	JwtRsaPublicKeyPath string        `default:"/var/lib/webis/keys" split_words:"true"`
	Mode                string        `default:"debug"`
	Addr                string        `default:"127.0.0.1"`
	Port                int           `default:"8379"`
	GracefulTimeout     time.Duration `default:"15s" split_words:"true"`

	RSAPublicKey []byte
}

// NewConfiguration returns a configuration that is loaded with environment variables
func NewConfiguration() (*Configs, error) {
	config := new(Configs)
	err := envconfig.Process("webis", config)
	if err != nil {
		return nil, fmt.Errorf("error while processing global configs from env variables, error: %v", err)
	}

	// loading SockHub configs
	sockHubConfig := websocket.Configuration{}
	err = envconfig.Process("webis_sock", &sockHubConfig)
	if err != nil {
		return nil, fmt.Errorf("error while processing sockhub configs from env variables, error: %v", err)
	}
	config.SockHubConfig = sockHubConfig

	// loading logging configs
	loggingConfig := logger.Configuration{}
	err = envconfig.Process("webis_logging", &sockHubConfig)
	if err != nil {
		return nil, fmt.Errorf("error while processing logging configs from env variables, error: %v", err)
	}
	config.LoggingConfigs = loggingConfig

	// loading redis configs
	redisConfigs := redis.Configs{}
	err = envconfig.Process("webis_redis", &redisConfigs)
	if err != nil {
		return nil, fmt.Errorf("error while processing redis client configs from env variables, error: %v", err)
	}
	config.RedisConfigs = redisConfigs
	//
	//// Load RSA public keys to config object
	//rsaPubKey, err := getRSAPublicKeys(config.JwtRsaPublicKeyPath)
	//if err != nil {
	//	return nil, fmt.Errorf("error while loading rsa public keys to config, error: %v", err)
	//}
	//config.RSAPublicKey = rsaPubKey

	return config, nil
}

func getRSAPublicKeys(jwtRSAPath string) ([]byte, error) {
	key, err := ioutil.ReadFile(jwtRSAPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key from %s", jwtRSAPath)
	}
	return key, nil
}

// Package logger introduces different logger factories
package logger

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

// Configuration is generic struct for logging configs
type Configuration struct {
	Enabled             bool   `default:"true" split_words:"true"`
	Debug               bool   `default:"true" split_words:"true"`
	Level               string `default:"info" split_words:"true"`
	Pretty              bool   `default:"false" split_words:"true"`
	FileRedirectEnabled bool   `default:"false" split_words:"true"`
	FileRedirectPath    string `default:"/var/log" split_words:"true"`
	FileRedirectPrefix  string `default:"webis" split_words:"true"`

	// CoreFields are constant data that should be passed through all logs
	CoreFields map[string]interface{}
}

// GetConfigFromEnv tries to generate configuration from related
// environment variables with power of "kelseyhightower" library.
func GetConfigFromEnv(prefix string) (Configuration, error) {
	loggerConfig := Configuration{}
	if err := envconfig.Process(prefix, &loggerConfig); err != nil {
		return loggerConfig, fmt.Errorf("error while loading configs from env variables, error: %v", err)
	}

	return loggerConfig, nil
}

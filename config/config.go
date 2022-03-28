package config

import (
	"strings"
	"time"

	"github.com/mattermost/mattermost-app-chaosengine/store"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// App config for Mattermost app
type App struct {
	Type    apps.AppType
	RootURL string `mapstructure:"root_url"`
	Secret  string
}

// Options config to set to run the app.
type Options struct {
	Debug         bool
	App           App
	ListenAddress string `mapstructure:"address"`
	IsLocal       bool   `mapstructure:"local"`
	Environment   string
	Database      store.Config `mapstructure:"db"`
}

func (o *Options) Validate() error {
	return nil
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("chaos_engine")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	defaults := map[string]interface{}{
		"debug":       false,
		"environment": "dev",
		"address":     ":3000",

		// application settings if http or lambda
		"app.type":     apps.AppTypeHTTP,
		"app.root_url": "http://localhost:3000",
		"app.secret":   "secretkey",

		// database
		"db.rds.secret_name":   nil, // to be supported
		"db.idle_conns":        2,
		"db.max_open_conns":    1,
		"db.max_conn_lifetime": time.Hour,
	}

	for key, value := range defaults {
		viper.SetDefault(key, value)
	}
}

// Load will load the necessary config
func Load(logger logrus.FieldLogger) (Options, error) {
	if err := viper.ReadInConfig(); err != nil {
		logger.Warn(errors.Wrap(err, "unable to find config.yml. loading config from environment variables").Error())
	}

	var cfg Options
	if err := viper.Unmarshal(&cfg); err != nil {
		return Options{}, errors.Wrap(err, "failed to load")
	}
	if err := cfg.Validate(); err != nil {
		return Options{}, errors.Wrap(err, "failed to validate the config")
	}

	return cfg, nil
}

func SetDatabaseConfig(scheme string, url string, logger logrus.FieldLogger) (Options, error) {

	viper.Set("db.scheme", scheme)
	viper.Set("db.url", url)

	return Load(logger)
}

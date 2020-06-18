package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/go-redis/redis"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

var (
	instance *Config
	once     sync.Once
)

// GetInstance always returns the same instance of Config
func GetInstance() *Config {
	once.Do(func() {
		instance = New(false)
	})
	return instance
}

// GetTestInstance always returns the same instance of Config
func GetTestInstance() *Config {
	once.Do(func() {
		instance = New(true)
	})
	return instance
}

// Config ...
type Config struct {
	RedisClient *redis.Client
	Pipefile    string // Simulate the complex scoring pipeline
	Prefilter   string
}

// New Config
func New(test bool) *Config {
	// handle env
	if !test {
		err := gotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := gotenv.Load("../.env")
		if err != nil {
			log.Fatal(err)
		}
	}
	viper.AutomaticEnv()

	c := &Config{
		Pipefile:  viper.GetString("PIPE_FILENAME"),
		Prefilter: viper.GetString("PREFILTER_FILENAME"),
	}

	// handle logger
	c.initLogger()

	if !test {
		c.RedisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", viper.GetString("REDIS_HOST"), viper.GetInt("REDIS_PORT")),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		})

		if _, err := c.RedisClient.Ping().Result(); err != nil {
			logrus.Fatal(err)
		}
	}

	return c
}

func (c *Config) initLogger() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   viper.GetBool("LOG_FULL_TIMESTAMP"),
		TimestampFormat: viper.GetString("LOG_TIME_FORMAT"),
		ForceColors:     viper.GetBool("LOG_FORCE_COLORS"),
		DisableColors:   viper.GetBool("LOG_DISABLE_COLORS"),
	})
}

package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"

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
		if instance == nil {
			instance = New()
		}
	})
	return instance
}

// Config ...
type Config struct {
	RedisClient   *redis.Client
	Pipefile      string // Simulate the complex scoring pipeline
	Prefilter     string
	Postfilter    string
	AdminUser     string
	AdminPass     string
	ClientUser    string
	ClientPass    string
	RetryAttempts int
}

func new(omitRedis bool) *Config {
	c := &Config{}

	if c.Pipefile = viper.GetString("PIPE_FILENAME"); c.Pipefile == "" {
		log.Fatalf("failed to fetch config: %q", "PIPE_FILENAME")
	}

	if c.Prefilter = viper.GetString("PREFILTER_FILENAME"); c.Prefilter == "" {
		log.Fatalf("failed to fetch config: %q", "PREFILTER_FILENAME")
	}

	if c.Postfilter = viper.GetString("POSTFILTER_FILENAME"); c.Postfilter == "" {
		log.Fatalf("failed to fetch config: %q", "POSTFILTER_FILENAME")
	}

	if c.AdminUser = viper.GetString("ADMIN_USER"); c.AdminUser == "" {
		log.Fatalf("failed to fetch config: %q", "ADMIN_USER")
	}

	if c.AdminPass = viper.GetString("ADMIN_PASS"); c.AdminPass == "" {
		log.Fatalf("failed to fetch config: %q", "ADMIN_PASS")
	}

	if c.ClientUser = viper.GetString("CLIENT_USER"); c.ClientUser == "" {
		log.Fatalf("failed to fetch config: %q", "CLIENT_USER")
	}

	if c.ClientPass = viper.GetString("CLIENT_PASS"); c.ClientPass == "" {
		log.Fatalf("failed to fetch config: %q", "CLIENT_PASS")
	}

	c.RetryAttempts = viper.GetInt("RETRY_ATTEMPTS")

	// handle logger
	c.initLogger()

	if !omitRedis {
		c.RedisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", viper.GetString("REDIS_HOST"), viper.GetInt("REDIS_PORT")),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		})

		if _, err := c.RedisClient.Ping().Result(); err != nil {
			logrus.Fatal(errors.Wrap(err, "failed to connect to redis"))
		}
	}

	return c
}

// New Config
func New() *Config {
	// handle env
	_ = gotenv.Load()
	viper.AutomaticEnv()

	return new(false)
}

// NewTest Config
func NewTest() *Config {
	// handle env
	err := gotenv.Load("../.env")
	if err != nil {
		log.Fatal(err)
	}
	viper.AutomaticEnv()

	return new(true)
}

// NewTestDB Config
func NewTestDB() *Config {
	// handle env
	err := gotenv.Load("../.env")
	if err != nil {
		log.Fatal(err)
	}
	viper.AutomaticEnv()

	return new(false)
}

// OverrideInstance ..
// Allowing tests running both in docker and host machines/CI requires override of some configs.
// Specifically omitting redis or pointing to a different env.
func OverrideInstance(c *Config) {
	instance = c
}

func (c *Config) initLogger() {
	viper.SetDefault("log_level", "info")
	level, err := logrus.ParseLevel(viper.GetString("LOG_LEVEL"))
	if err != nil {
		log.Fatal("invalid log level")
	}

	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   viper.GetBool("LOG_FULL_TIMESTAMP"),
		TimestampFormat: viper.GetString("LOG_TIME_FORMAT"),
		ForceColors:     viper.GetBool("LOG_FORCE_COLORS"),
		DisableColors:   viper.GetBool("LOG_DISABLE_COLORS"),
	})
}

// DisableLogging enables to disable all logging for testing.
func (c *Config) DisableLogging() {
	logrus.SetOutput(ioutil.Discard)
}

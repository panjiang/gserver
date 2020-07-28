package config

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"

	"github.com/panjiang/gserver/pkg/redisdb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

var filename string
var debug bool

func init() {
	flag.StringVar(&filename, "conf", "config.yml", "yaml config file path")
	flag.BoolVar(&debug, "debug", true, "open debug log")
}

// Config 配置
type Config struct {
	Name  string          `yaml:"-"`
	Queue *queueSevice    `yaml:"queueService"`
	Redis *redisdb.Config `yaml:"redis"`
}

func (c *Config) validate() (err error) {
	if c.Queue == nil {
		err = errors.New("invalid conf")
		return
	}
	if err = c.Queue.validate(); err != nil {
		return
	}
	return
}

// Parse 从yaml文件中解析配置对象
func Parse(name string) (*Config, error) {
	// 日志级别
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if filename == "" {
		return nil, errors.New("must specify config file")
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	conf := Config{Name: name}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}

	return &conf, nil
}

package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type RabbitMQConf struct {
	Uri string `yaml:"uri"`

	Exchange struct {
		AutoDelete   bool                   `yaml:"auto-delete"`
		Durable      bool                   `yaml:"durable"`
		ExchangeName string                 `yaml:"exchange-name"`
		ExchangeType string                 `yaml:"exchange-type"`
		Arguments    map[string]interface{} `yaml:"arguments,omitempty"`
	}

	Queue struct {
		AutoDelete bool                   `yaml:"auto-delete"`
		Durable    bool                   `yaml:"durable"`
		Exclusive  bool                   `yaml:"exclusive"`
		QueueName  string                 `yaml:"queueName"`
		RoutingKey string                 `yaml:"routing-key"`
		Arguments  map[string]interface{} `yaml:"arguments,omitempty"`
	}
}

type Config struct {
	Workers int          `yaml:"workers"`
	SrcConf RabbitMQConf `yaml:"rabbitmq-src"`
	DstConf RabbitMQConf `yaml:"rabbitmq-dst"`
}

func LoadConfiguration(configFilePath string) (config *Config, err error) {
	configBuf, err := os.ReadFile(configFilePath)
	if err != nil {
		err = fmt.Errorf("%v", err)
	} else {
		fileExtension := filepath.Ext(configFilePath)
		if fileExtension == ".yaml" {
			config = new(Config)
			if err = yaml.Unmarshal(configBuf, &config); err != nil {
				err = fmt.Errorf("%v", err)
			}
		} else {
			err = fmt.Errorf("config file [%s] not support", configFilePath)
		}
	}
	return config, err
}

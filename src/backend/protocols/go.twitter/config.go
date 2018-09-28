// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC

package twitterworker

import (
	broker"github.com/CaliOpen/Caliopen/src/backend/brokers/go.twitter"
)

type WorkerConfig struct {
	TwitterAppKey string `mapstructure:"twitter_app_key"`
	TwitterAppSecret string `mapstructure:"twitter_app_secret"`
	BrokerConfig broker.BrokerConfig `mapstructure:"BrokerConfig"`
}

const (
	DefaultPollInterval = 10 // poll interval in seconds
)

var TwitterAppKey string
var TwitterAppSecret string
var AppConfig *WorkerConfig

func init() {
	TwitterAppKey ="TSpJhxzkNo43Q0d64Vz10a29I"
	TwitterAppSecret = "kU4mWmFXZPXdzztePHA0uwU57m6CpfbGp2TNzpwgmDiGyjUyxL"
	AppConfig = &WorkerConfig{}
}

package config

import (
	"time"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

type NodeCfg struct {
	RemoteAddrs        []string `envconfig:"remote_addrs"`
	SelfAddr           string   `envconfig:"self_addr"`
	PublicAddr         string   `envconfig:"public_addr"`
	RpcFaultPercentage int      `envconfig:"rpc_fault_percentage"`
	DBFaultPercentage  int      `envconfig:"db_fault_percentage"`
}

func GetNodeCfg(appPrefix string) NodeCfg {
	cfg := NodeCfg{}
	envconfig.MustProcess(appPrefix, &cfg)
	return cfg
}

type QuorumCfg struct {
	Write int
	Read  int
}

type Config struct {
	NodeID     string
	QuorumCfg  QuorumCfg
	ReadRetry  int
	WriteRetry int
	OpTimeout  time.Duration
}

func ReadFromEnv() *Config {
	return nil
}

type Opt func(*Config)

func WithOpTimeout(timeout time.Duration) Opt {
	return func(c *Config) {
		c.OpTimeout = timeout
	}
}

func WithRetries(read, write int) Opt {
	return func(c *Config) {
		c.ReadRetry = read
		c.WriteRetry = write
	}
}

func NewCfg(readQ, writeQ int, opts ...Opt) *Config {
	cfg := &Config{
		NodeID: uuid.New().String(),
		QuorumCfg: QuorumCfg{
			Write: writeQ,
			Read:  readQ,
		},
		ReadRetry:  2,
		WriteRetry: 2,
		OpTimeout:  time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

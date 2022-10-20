package xrds

type Config struct {
	Addr          string `yaml:"addr"`
	Password      string `yaml:"password"`
	DB            int    `yaml:"db"`
	MaxRetries    int    `yaml:"maxRetries"`
	PoolSize      int    `yaml:"poolSize"`
	MinIdleConns  int    `yaml:"minIdleConns"`
	MetricsEnable bool   `yaml:"metricsEnable"`
}

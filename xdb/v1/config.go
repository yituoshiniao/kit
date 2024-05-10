package v1

// Deprecated
// 兼容旧版本，新版本使用 Config 结构体
type Mysql = Config

type Config struct {
	Dsn         string `yaml:"dsn"`
	MaxIdle     int    `yaml:"maxIdle"`
	MaxOpen     int    `yaml:"maxOpen"`
	MaxLifetime int    `yaml:"maxLifetime"`
	LogMode     bool   `yaml:"logMode"`

	//普罗米修斯监控配置
	DBName          string `yaml:"dbName"`
	PushAddr        string `yaml:"pushAddr"`
	RefreshInterval int    `yaml:"refreshInterval"`
}

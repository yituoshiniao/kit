package xtrace

import "github.com/uber/jaeger-client-go/config"

// Config 增加 config 结构体别名，和 xdb、xrds 风格保持一致
type Config config.Configuration

// OtelConfig 增加 config 结构体别名，和 xdb、xrds 风格保持一致
type OtelConfig struct {
	// #常量配置
	SamplerParam float64
	// #数据服务器地址
	ReporterLocalAgentHostPort string
	// 服务名
	ServerName string
	// 是否安全模式
	Insecure string
}

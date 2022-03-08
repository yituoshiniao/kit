package xtrace

import "github.com/uber/jaeger-client-go/config"

// 增加 config 结构体别名，和 xdb、xrds 风格保持一致
type Config config.Configuration

//go:build prod

package config

import (
	"time"
)

const (
	// APP
	DefaultProg = "fss"
	// REDIS
	REDIS_HOST     = "10.5.0.4:6379"
	REDIS_PWD      = ""
	REDIS_DB_MAIN  = 0
	REDIS_DB_INCR  = 1
	REDIS_TTL_MAIN = 180 * time.Second // 3 Minute
	REDIS_TTL_INCR = 300 * time.Second // 5 Minute
	// CHI
	CHI_ADDR           = "10.5.0.5"
	CHI_PORT           = "8338"
	CHI_DefaultTimeOut = 30 * time.Second // 30 Seconds
)

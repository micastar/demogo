package global

import (
	"time"
)

const (
	DefaultTimeFormat = "2006-01-02"

	DisableKeepAlives = false

	DefaultTimeOut = 30 * time.Second

	Limit = 10

	DefaultStoreInterval = 3 * time.Hour

	GetIdInterval = 30 * time.Minute

	DefaultSendInterval = 1 * time.Hour

	DefaultPausingInterval = 20 * time.Minute
)

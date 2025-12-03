package channels

import (
	"context"
	"time"

	"github.com/forhsd/alert/errors"
	"github.com/matcornic/hermes/v2"
)

// Channel 告警渠道接口
type Channel interface {
	// Send 发送告警消息
	Send(ctx context.Context, title string, content []*errors.ErrorDetail) error
	// Close 关闭渠道
	Close() error
	// Name 渠道名称
	Name() string
}

// BaseChannel 基础渠道
type BaseChannel struct {
	hermes.Hermes
	Name       string
	Timeout    time.Duration
	RetryTimes int
}

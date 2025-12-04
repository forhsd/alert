package channels

import (
	"context"
	"fmt"
	"time"

	"github.com/forhsd/alert/errors"
	"github.com/matcornic/hermes/v2"
)

// Channel 告警渠道接口
type Channel interface {
	// 配置验证
	Validate() error
	// Send 发送告警消息
	Send(ctx context.Context, title string, content []*errors.ErrorDetail) error
	// Close 关闭渠道
	Close() error
	// Name 渠道名称
	Name() string
}

// 通知内容接口
type Notice interface {
	Email(title string, content []*errors.ErrorDetail) hermes.Email
	Name() string
}

// BaseChannel 基础渠道
type BaseChannel struct {
	Notice
	hermes.Hermes
	Name       string
	Timeout    time.Duration
	RetryTimes int
}

func (e *BaseChannel) Validate() error {

	if e.Notice == nil {
		return fmt.Errorf("通知内容接口不能为空")
	}
	if e.Timeout == 0 {
		e.Timeout = time.Second * 30
	}
	if e.RetryTimes == 0 {
		e.RetryTimes = 3
	}
	return nil
}

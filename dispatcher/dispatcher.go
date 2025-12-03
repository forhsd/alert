package dispatcher

import (
	"context"
	"sync"
	"time"

	"github.com/forhsd/alert/channels"
	"github.com/forhsd/alert/errors"
)

// Dispatcher 分发器
type Dispatcher struct {
	channels map[string]channels.Channel
	interval time.Duration
	mu       sync.RWMutex
}

// NewDispatcher 创建分发器
func NewDispatcher(channels map[string]channels.Channel, interval time.Duration) *Dispatcher {
	return &Dispatcher{
		channels: channels,
		interval: interval,
	}
}

// Dispatch 分发消息到所有渠道
func (d *Dispatcher) Dispatch(errors []*errors.ErrorDetail) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, channel := range d.channels {
		wg.Add(1)
		go func(ch channels.Channel) {
			defer wg.Done()

			var err error

			// 带重试的发送
			for i := range 3 {
				if err = ch.Send(ctx, "系统告警", errors); err == nil {
					break
				}
				if i < 2 {
					time.Sleep(time.Duration(i+1) * time.Second)
				}
			}
			if err == nil {
				for i := range errors {
					errors[i].IsSend = true
				}
			}
		}(channel)
	}

	wg.Wait()
}

// AddChannel 添加渠道
func (d *Dispatcher) AddChannel(name string, channel channels.Channel) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.channels[name] = channel
}

// RemoveChannel 移除渠道
func (d *Dispatcher) RemoveChannel(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.channels, name)
}

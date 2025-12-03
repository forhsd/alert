package storage

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/forhsd/alert/channels"
	"github.com/forhsd/alert/dispatcher"
	"github.com/forhsd/alert/errors"
)

// AlertConfig 告警配置
type AlertConfig struct {
	ReportInterval      time.Duration               `json:"report_interval"`  // 报告间隔，默认10分钟
	DeduplicationWindow time.Duration               `json:"dedup_window"`     // 去重时间窗口
	BufferSize          int                         `json:"buffer_size"`      // 缓冲区大小
	EnabledChannels     []string                    `json:"enabled_channels"` // 启用的渠道
	ChannelConfigs      map[string]channels.Channel `json:"channel_configs"`  // 各渠道配置
	ErrorHandler        func(error)                 `json:"-"`                // 错误处理器
}

// AlertLibrary 告警库主体
type AlertLibrary struct {
	config     *AlertConfig
	channels   map[string]channels.Channel
	storage    *ErrorStorage
	dispatcher *dispatcher.Dispatcher
	errorChan  chan *errors.ErrorDetail
	closeChan  chan struct{}
	shutdownWG sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	isShutdown atomic.Bool
}

// NewAlertLibrary 创建告警库实例
func NewAlertLibrary(config *AlertConfig) (*AlertLibrary, error) {
	if config.ReportInterval == 0 {
		config.ReportInterval = 10 * time.Minute
	}
	if config.DeduplicationWindow == 0 {
		config.DeduplicationWindow = 1 * time.Hour
	}
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())

	lib := &AlertLibrary{
		config:    config,
		channels:  make(map[string]channels.Channel),
		errorChan: make(chan *errors.ErrorDetail, config.BufferSize),
		closeChan: make(chan struct{}),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 初始化存储
	lib.storage = NewErrorStorage(config.DeduplicationWindow)

	// 初始化渠道
	if err := lib.initChannels(); err != nil {
		return nil, err
	}

	// 初始化分发器
	lib.dispatcher = dispatcher.NewDispatcher(lib.channels, config.ReportInterval)

	// 启动工作协程
	lib.startWorkers()

	return lib, nil
}

// initChannels 初始化告警渠道
func (a *AlertLibrary) initChannels() error {
	for _, channelName := range a.config.EnabledChannels {
		config, exists := a.config.ChannelConfigs[channelName]
		if !exists {
			return fmt.Errorf("config for channel %s not found", channelName)
		}

		// var channel channels.Channel
		// var err error

		// switch channelName {
		// case "email":
		// 	channel, err = channels.NewEmailChannel(&config)
		// // case "sms":
		// // 	channel, err = channels.NewSMSChannel(config)
		// // case "wechat":
		// // 	channel, err = channels.NewWeChatChannel(config)
		// // case "dingtalk":
		// // 	channel, err = channels.NewDingTalkChannel(config)
		// default:
		// 	return fmt.Errorf("unknown channel: %s", channelName)
		// }

		// if err != nil {
		// 	return fmt.Errorf("failed to init channel %s: %v", channelName, err)
		// }

		a.channels[channelName] = config
	}

	return nil
}

// startWorkers 启动工作协程
func (a *AlertLibrary) startWorkers() {
	a.shutdownWG.Add(2)

	// 错误处理协程
	go func() {
		defer a.shutdownWG.Done()
		a.processErrors()
	}()

	// 定时报告协程
	go func() {
		defer a.shutdownWG.Done()
		a.scheduledReporting()
	}()
}

// ReportError 报告错误（主接口）
func (a *AlertLibrary) ReportError(err error, metadata ...any) {
	a.ReportErrorWithLevel(err, errors.LevelError, metadata...)
}

// ReportErrorWithLevel 指定级别报告错误
func (a *AlertLibrary) ReportErrorWithLevel(err error, level errors.AlertLevel, metadata ...any) {

	a.shutdownWG.Add(1)
	defer a.shutdownWG.Done()

	a.mu.RLock()
	if a.isShutdown.Load() {
		log.Println("告警己关闭", err)
		a.mu.RUnlock()
		return
	}
	defer a.mu.RUnlock()

	stack := string(debug.Stack())
	errorDetail := &errors.ErrorDetail{
		Message:   err.Error(),
		Stack:     stack,
		Level:     level,
		Count:     1,
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
	}

	if len(metadata) > 0 {
		errorDetail.Metadata = metadata
	}

	// 异步处理
	select {
	case a.errorChan <- errorDetail:
		// 成功发送到通道
	default:
		// 通道满了，直接丢弃或调用错误处理器
		if a.config.ErrorHandler != nil {
			a.config.ErrorHandler(fmt.Errorf("alert buffer full, error dropped: %v", err))
		}
	}
}

// processErrors 处理错误队列
func (a *AlertLibrary) processErrors() {
	for {
		select {
		case <-a.ctx.Done():
			// Context被取消，尝试处理完通道内现有错误再退出
			a.drainErrorChannel()
			return
		case <-a.closeChan:
			// 收到关闭信号，直接退出
			return
		case errDetail := <-a.errorChan:
			a.storage.AddError(errDetail)
		}
	}
}

// scheduledReporting 定时报告
func (a *AlertLibrary) scheduledReporting() {
	ticker := time.NewTicker(a.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.flushErrors()
		}
	}
}

// flushErrors 发送暂存的错误
func (a *AlertLibrary) flushErrors() {
	errors := a.storage.GetErrors()
	if len(errors) == 0 {
		return
	}

	// 分发到各个渠道
	a.dispatcher.Dispatch(errors)

	// 清空已发送的错误
	a.storage.ClearSent()
}

// Shutdown 优雅关闭
func (a *AlertLibrary) Shutdown() error {

	a.mu.Lock()
	if a.isShutdown.Load() {
		a.mu.Unlock()
		return nil
	}
	defer a.mu.Unlock()

	a.isShutdown.CompareAndSwap(false, true)

	// 发送关闭信号
	a.cancel()

	close(a.errorChan)

	// 3. 排空错误通道，确保所有已进入通道的错误被处理
	a.drainErrorChannel()

	// 发送暂存的所有错误
	a.flushErrors()

	// 等待所有协程完成
	close(a.closeChan)
	a.shutdownWG.Wait()

	// 关闭所有渠道
	for name, channel := range a.channels {
		if err := channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel %s: %v", name, err)
		}
	}

	return nil
}

// Flush 手动刷新缓冲区
func (a *AlertLibrary) Flush() {
	a.flushErrors()
}

// drainErrorChannel 安全地排空错误通道
func (a *AlertLibrary) drainErrorChannel() {
	for errDetail := range a.errorChan {
		// 将通道中剩余的错误写入存储
		a.storage.AddError(errDetail)
	}
}

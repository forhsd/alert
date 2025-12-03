package channels

import "time"

// WeChatChannel 微信渠道
type WeChatChannel struct {
	BaseChannel
	config map[string]any
}

func NewWeChatChannel(config any) (*WeChatChannel, error) {
	cfg := config.(map[string]any)
	return &WeChatChannel{
		BaseChannel: BaseChannel{
			Name:       "wechat",
			Timeout:    10 * time.Second,
			RetryTimes: 3,
		},
		config: cfg,
	}, nil
}

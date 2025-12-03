package channels

import "time"

// DingTalkChannel 钉钉渠道
type DingTalkChannel struct {
	BaseChannel
	config map[string]any
}

func NewDingTalkChannel(config any) (*DingTalkChannel, error) {
	cfg := config.(map[string]any)
	return &DingTalkChannel{
		BaseChannel: BaseChannel{
			Name:       "dingtalk",
			Timeout:    10 * time.Second,
			RetryTimes: 3,
		},
		config: cfg,
	}, nil
}

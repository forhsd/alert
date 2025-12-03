package channels

import "time"

// SMSChannel 短信渠道
type SMSChannel struct {
	BaseChannel
	config map[string]any
}

func NewSMSChannel(config any) (*SMSChannel, error) {
	cfg := config.(map[string]any)
	return &SMSChannel{
		BaseChannel: BaseChannel{
			Name:       "sms",
			Timeout:    10 * time.Second,
			RetryTimes: 3,
		},
		config: cfg,
	}, nil
}

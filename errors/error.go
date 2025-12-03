package errors

import (
	"time"
)

// AlertLevel 告警级别
type AlertLevel int

const (
	LevelInfo     AlertLevel = 1
	LevelWarning  AlertLevel = 2
	LevelError    AlertLevel = 3
	LevelCritical AlertLevel = 4
)

func (a AlertLevel) String() string {
	switch a {
	case LevelInfo:
		return "Info"
	case LevelWarning:
		return "Warning"
	case LevelError:
		return "Error"
	case LevelCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	ID        string     `json:"id"`
	Message   string     `json:"message"`
	Stack     string     `json:"stack,omitempty"`
	Level     AlertLevel `json:"level"`
	Count     int        `json:"count"`
	FirstSeen time.Time  `json:"first_seen"`
	LastSeen  time.Time  `json:"last_seen"`
	Metadata  any        `json:"metadata,omitempty"`
	IsSend    bool       `json:"-"`
}

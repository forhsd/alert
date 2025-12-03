package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/forhsd/alert/errors"
)

// ErrorStorage 错误存储
type ErrorStorage struct {
	errors map[string]*errors.ErrorDetail
	mu     sync.RWMutex
	window time.Duration
}

// NewErrorStorage 创建错误存储
func NewErrorStorage(window time.Duration) *ErrorStorage {
	return &ErrorStorage{
		errors: make(map[string]*errors.ErrorDetail),
		window: window,
	}
}

// AddError 添加错误（去重）
func (s *ErrorStorage) AddError(err *errors.ErrorDetail) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 生成错误指纹
	fingerprint := s.generateFingerprint(err)

	if existing, exists := s.errors[fingerprint]; exists {
		// 更新时间窗口内的错误
		if time.Since(existing.LastSeen) < s.window {
			existing.Count++
			existing.LastSeen = time.Now()
			if err.Metadata != nil && existing.Metadata == nil {
				existing.Metadata = err.Metadata
			}
		} else {
			// 超出时间窗口，重新计数
			err.FirstSeen = time.Now()
			err.LastSeen = time.Now()
			err.Count = 1
			s.errors[fingerprint] = err
		}
	} else {
		// 新错误
		err.ID = fingerprint
		s.errors[fingerprint] = err
	}
}

// generateFingerprint 生成错误指纹
func (s *ErrorStorage) generateFingerprint(err *errors.ErrorDetail) string {
	// 使用错误消息和堆栈的哈希作为指纹
	hash := sha256.New()
	hash.Write([]byte(err.Message))
	hash.Write([]byte(err.Stack))
	return hex.EncodeToString(hash.Sum(nil))[:16]
}

// GetErrors 获取所有错误
func (s *ErrorStorage) GetErrors() []*errors.ErrorDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errors := make([]*errors.ErrorDetail, 0, len(s.errors))
	for _, err := range s.errors {
		errors = append(errors, err)
	}
	return errors
}

// ClearSent 清空已发送的错误
func (s *ErrorStorage) ClearSent() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 只清空超过时间窗口的错误
	// now := time.Now()
	for key, err := range s.errors {
		_ = err
		// if now.Sub(err.LastSeen) > s.window {
		// delete(s.errors, key)
		// }
		if err.IsSend {
			delete(s.errors, key)
		}
	}
}

// ClearAll 清空所有错误
func (s *ErrorStorage) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors = make(map[string]*errors.ErrorDetail)
}

// GetStats 获取统计信息
func (s *ErrorStorage) GetStats() (total, unique int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	unique = len(s.errors)
	total = 0
	for _, err := range s.errors {
		total += err.Count
	}
	return total, unique
}

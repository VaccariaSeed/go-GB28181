package go_GB28181

import (
	"context"
	"sync"
	"time"
)

func NewMessageCar(resultSize int, timeout time.Duration) *MessageCar {
	car := &MessageCar{
		timeout:   timeout,
		value:     make([]any, 0),
		ch:        make(chan struct{}),
		valueSize: resultSize,
	}
	return car
}

type MessageCar struct {
	timeout   time.Duration //超时时间
	err       error         //错误
	valueSize int           //结果数量
	value     []any         //结果
	ch        chan struct{}
	isClosed  bool
	lock      sync.Mutex
}

func (m *MessageCar) Down() ([]any, error) {
	select {
	case <-m.ch:
		return m.value, m.err
	case <-time.After(m.timeout):
		m.Close()
		if len(m.value) == 0 {
			return nil, context.DeadlineExceeded
		}
		return m.value, nil
	}
}

// OccError 发生了一个错误
func (m *MessageCar) OccError(err error) {
	m.lock.Lock()
	if m.isClosed {
		m.lock.Unlock()
		return // 已关闭，不再接受新值
	}
	m.lock.Unlock()
	m.err = err
	m.Close()
}

// OccValue 新增一个结果
func (m *MessageCar) OccValue(value any) {
	m.lock.Lock()
	if m.isClosed {
		m.lock.Unlock()
		return
	}
	m.value = append(m.value, value)
	m.lock.Unlock()
	if m.valueSize != 0 {
		if len(m.value) >= m.valueSize {
			m.Close()
		}
	}
}

// Close 截断
func (m *MessageCar) Close() {
	m.lock.Lock()
	if !m.isClosed {
		m.isClosed = true
		close(m.ch)
	}
	m.lock.Unlock()
}

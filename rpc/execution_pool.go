package rpc

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JekaMas/workerpool"
)

type SafePool struct {
	executionPool atomic.Pointer[workerpool.WorkerPool]

	sync.RWMutex
	timeout  time.Duration
	size     int
	fastPath bool // Skip sending task to execution pool
}

//TODO: we call `github.com/ethereum/go-ethereum/rpc.(*Client).newClientConn` as many times as clients calls, that creates multiple execution pools

func NewExecutionPool(initialSize int, timeout time.Duration) *SafePool {
	sp := &SafePool{
		size:    initialSize,
		timeout: timeout,
	}

	if initialSize == 0 {
		sp.fastPath = true

		return sp
	}

	p := workerpool.New(initialSize)
	sp.executionPool.Store(p)

	return sp
}

func (s *SafePool) Submit(ctx context.Context, fn func() error) (<-chan error, bool) {
	if fn == nil {
		return nil, true
	}

	if s.isFastPath() {
		go func() {
			_ = fn()
		}()

		return nil, true
	}

	pool := s.executionPool.Load()
	if pool == nil {
		return nil, false
	}

	return pool.Submit(ctx, fn, s.Timeout()), true
}

func (s *SafePool) ChangeSize(n int) {
	s.Lock()
	if n <= 0 {
		s.fastPath = true
		n = 0
	}

	s.size = n

	var newPool *workerpool.WorkerPool

	if n >= 0 {
		newPool = workerpool.New(n)
	}

	oldPool := s.executionPool.Swap(newPool)

	s.Unlock()

	if oldPool != nil {
		go func() {
			oldPool.StopWait()
		}()
	}
}

func (s *SafePool) ChangeTimeout(n time.Duration) {
	s.Lock()
	defer s.Unlock()

	s.timeout = n
}

func (s *SafePool) Timeout() time.Duration {
	s.RLock()
	defer s.RUnlock()

	return s.timeout
}

func (s *SafePool) isFastPath() bool {
	s.RLock()
	defer s.RUnlock()

	return s.fastPath
}

func (s *SafePool) Size() int {
	s.RLock()
	defer s.RUnlock()

	return s.size
}

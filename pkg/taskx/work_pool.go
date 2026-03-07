package taskx

import (
	"context"
	"hash/fnv"
	"sync"
	"sync/atomic"
)

type ITask interface {
	OnExecute()
	OnDone(err error)
}

type Pool struct {
	nextId    atomic.Uint64
	chanList  []chan ITask
	ctx       context.Context
	cancel    context.CancelFunc
	config    *Config
	wg        sync.WaitGroup
	stopped atomic.Bool
}

func NewPool(opts ...Option) *Pool {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		chanList: make([]chan ITask, cfg.workerNum),
		ctx:      ctx,
		cancel:   cancel,
		config:   cfg,
	}
	for i := range p.chanList {
		p.chanList[i] = make(chan ITask, cfg.chanSize)
	}
	for i := range p.chanList {
		p.wg.Add(1)
		go p.worker(i)
	}
	return p
}

func (p *Pool) Stop() {
	p.stopped.Store(true)
	p.cancel()
	for _, ch := range p.chanList {
		close(ch)
	}
	p.wg.Wait()
}

func (p *Pool) Submit(task ITask) error {
	if p.stopped.Load() {
		return p.reject(task, ErrPoolStopped)
	}
	idx := int(p.nextId.Add(1)-1) % len(p.chanList)
	return p.tryEnqueue(idx, task)
}

func (p *Pool) SubmitWithKey(key string, task ITask) error {
	if p.stopped.Load() {
		return p.reject(task, ErrPoolStopped)
	}
	idx := p.hashIndex(key)
	return p.tryEnqueue(idx, task)
}

func (p *Pool) tryEnqueue(idx int, task ITask) error {
	select {
	case p.chanList[idx] <- task:
		return nil
	default:
		return p.reject(task, ErrPoolFull)
	}
}

func (p *Pool) reject(task ITask, _ error) error {
	return p.config.rejectPolicy.Reject(task, p)
}

func (p *Pool) executeTask(task ITask) {
	var err error
	if p.config.safeRun {
		err = safeRun(task)
	} else {
		task.OnExecute()
	}
	task.OnDone(err)
}

func (p *Pool) discardOldestAndRetry(task ITask) error {
	// 尝试从轮询到的 channel 中取出最老的任务丢弃
	idx := int(p.nextId.Load()-1) % len(p.chanList)
	select {
	case oldTask := <-p.chanList[idx]:
		oldTask.OnDone(ErrPoolFull)
	default:
	}
	// 再次尝试入队
	select {
	case p.chanList[idx] <- task:
		return nil
	default:
		return ErrPoolFull
	}
}

func (p *Pool) worker(workerId int) {
	defer p.wg.Done()
	ch := p.chanList[workerId]
	for task := range ch {
		p.executeTask(task)
	}
}

func (p *Pool) hashIndex(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32()) % len(p.chanList)
}

func safeRun(task ITask) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = ErrPoolFull // fallback, shouldn't happen normally
			}
		}
	}()
	task.OnExecute()
	return nil
}

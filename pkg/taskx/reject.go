package taskx

type RejectPolicy interface {
	Reject(task ITask, pool *Pool) error
}

type AbortPolicy struct{}

func (p *AbortPolicy) Reject(_ ITask, _ *Pool) error {
	return ErrPoolFull
}

type DiscardPolicy struct{}

func (p *DiscardPolicy) Reject(_ ITask, _ *Pool) error {
	return nil
}

type CallerRunsPolicy struct{}

func (p *CallerRunsPolicy) Reject(task ITask, pool *Pool) error {
	pool.executeTask(task)
	return nil
}

type DiscardOldestPolicy struct{}

func (p *DiscardOldestPolicy) Reject(task ITask, pool *Pool) error {
	return pool.discardOldestAndRetry(task)
}

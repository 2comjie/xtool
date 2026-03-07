package taskx

import "runtime"

type Config struct {
	workerNum    int
	chanSize     int
	safeRun      bool
	rejectPolicy RejectPolicy
}

type Option func(cfg *Config)

func WithWorkerNum(workerNum int) Option {
	return func(cfg *Config) {
		if workerNum > 0 {
			cfg.workerNum = workerNum
		}
	}
}

func WithChanSize(chanSize int) Option {
	return func(cfg *Config) {
		if chanSize > 0 {
			cfg.chanSize = chanSize
		}
	}
}

func WithSafeRun(safeRun bool) Option {
	return func(cfg *Config) {
		cfg.safeRun = safeRun
	}
}

// WithRejectPolicy 设置拒绝策略（参照 Java ThreadPoolExecutor 的 RejectedExecutionHandler）。
func WithRejectPolicy(policy RejectPolicy) Option {
	return func(cfg *Config) {
		cfg.rejectPolicy = policy
	}
}

func DefaultConfig() *Config {
	workerNum := max(runtime.NumCPU()/2, 1)
	return &Config{
		workerNum:    workerNum,
		chanSize:     128,
		safeRun:      true,
		rejectPolicy: &AbortPolicy{},
	}
}

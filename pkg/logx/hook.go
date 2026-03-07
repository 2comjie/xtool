package logx

type Hook interface {
	// Levels 返回该钩子关注的日志级别。
	// 只有匹配的级别才会触发 Fire
	Levels() []Level

	// Fire 在日志输出前被调用
	// 返回 error 时该钩子执行失败，但不影响日志输出和其他钩子
	Fire(entry *Entry) error
}

// LevelHooks 按级别索引的钩子集合
type LevelHooks map[Level][]Hook

func (hooks LevelHooks) Add(hook Hook) {
	for _, level := range hook.Levels() {
		hooks[level] = append(hooks[level], hook)
	}
}

func (hooks LevelHooks) Fire(entry *Entry) {
	for _, hook := range hooks[entry.Level] {
		_ = hook.Fire(entry)
	}
}

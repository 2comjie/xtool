package atomicx

import "sync/atomic"

// Version 原子版本号，用于对象所有权追踪
// 每次转移所有权时 Incr，回收时 CompareAndDo 确保只有最后持有者回收
type Version struct {
	val atomic.Int64
}

func (v *Version) Incr() int64 {
	return v.val.Add(1)
}

func (v *Version) Load() int64 {
	return v.val.Load()
}

// CompareAndDo 如果版本号匹配则执行 fn
func (v *Version) CompareAndDo(ver int64, fn func()) bool {
	if v.val.Load() == ver {
		fn()
		return true
	}
	return false
}

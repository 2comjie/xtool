package taskx

type FuncTask struct {
	fn     func()
	doneFn func(error)
}

func NewFuncTask(fn func(), doneFn ...func(error)) *FuncTask {
	t := &FuncTask{fn: fn}
	if len(doneFn) > 0 {
		t.doneFn = doneFn[0]
	}
	return t
}

func (f *FuncTask) OnExecute() {
	f.fn()
}

func (f *FuncTask) OnDone(err error) {
	if f.doneFn != nil {
		f.doneFn(err)
	}
}

package safex

import "github.com/2comjie/xtool/pkg/logx"

func Run(runFunc func()) {
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("safe run err %+v", r)
		}
	}()

	runFunc()
}

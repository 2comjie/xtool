package bytex

import (
	"math/bits"
	"sync"
)

const (
	minSize   = 64
	maxSize   = 1 << 20 // 1MB
	numLevels = 15       // 64B ~ 1MB
)

type BufferPool struct {
	pools [numLevels]sync.Pool
}

var Default = NewBufferPool()

func NewBufferPool() *BufferPool {
	p := &BufferPool{}
	for i := range p.pools {
		size := minSize << i
		p.pools[i] = sync.Pool{
			New: func() any {
				b := make([]byte, size)
				return &b
			},
		}
	}
	return p
}

func (p *BufferPool) Get(size int) []byte {
	if size <= 0 {
		size = minSize
	}
	idx := p.index(size)
	if idx >= numLevels {
		return make([]byte, size)
	}
	bp := p.pools[idx].Get().(*[]byte)
	return (*bp)[:size]
}

func (p *BufferPool) Put(b []byte) {
	c := cap(b)
	if c < minSize || c > maxSize {
		return
	}
	idx := p.index(c)
	if idx >= numLevels {
		return
	}
	b = b[:c]
	p.pools[idx].Put(&b)
}

func (p *BufferPool) index(size int) int {
	if size <= minSize {
		return 0
	}
	// ceil(log2(size / minSize))
	n := (size - 1) / minSize
	return bits.Len(uint(n))
}

package basket_item

import "sync"

type Rules struct {
	maxCount int          // 1
	mx       sync.RWMutex `msgpack:"-"`
}

func (r *Rules) IsMaxCount() bool {
	return r.MaxCount() > 0
}

func (r *Rules) MaxCount() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.maxCount
}

func (r *Rules) SetMaxCount(v int) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.maxCount = v
}

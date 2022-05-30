package component

import (
	"sync"
)

type set map[any]struct{}

func (s set) has(item any) bool {
	_, exists := s[item]
	return exists
}

func (s set) insert(item any) {
	s[item] = struct{}{}
}

func (s set) delete(item any) {
	delete(s, item)
}

type WorkQueue struct {
	queue      []any
	processing set
	cond       *sync.Cond
	mtx        sync.Mutex
}

func (q *WorkQueue) Init() {
	q.queue = make([]any, 0)
	q.cond = sync.NewCond(new(sync.Mutex))
	q.processing = map[any]struct{}{}
}

func (q *WorkQueue) Pop() {
	q.cond.L.Lock()
	q.queue = q.queue[1:]
	q.cond.L.Unlock()
}

func (q *WorkQueue) Push(obj any) {
	q.cond.L.Lock()
	q.queue = append(q.queue, obj)
	q.cond.L.Unlock()
	q.cond.Signal()
}

func (q *WorkQueue) Top() any {
	q.cond.L.Lock()
	obj := q.queue[0]
	q.cond.L.Unlock()
	return obj
}

func (q *WorkQueue) Empty() bool {
	q.cond.L.Lock()
	flag := len(q.queue) == 0
	q.cond.L.Unlock()
	return flag
}

func (q *WorkQueue) Fetch() any {
	q.cond.L.Lock()
	if len(q.queue) == 0 {
		q.cond.Wait()
	}
	q.cond.L.Unlock()

	obj := q.Top()
	q.Pop()
	return obj
}

func (q *WorkQueue) Process(key string) bool {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	if q.processing.has(key) {
		return false
	} else {
		q.processing.insert(key)
		return true
	}
}

func (q *WorkQueue) Done(key string) {
	q.mtx.Lock()
	q.processing.delete(key)
	q.mtx.Unlock()
}

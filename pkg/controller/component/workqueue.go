package component

import (
	"sync"
)

type WorkQueue struct {
	queue []any
	cond  *sync.Cond
}

func (q *WorkQueue) Init() {
	q.queue = make([]any, 0)
	q.cond = sync.NewCond(new(sync.Mutex))
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

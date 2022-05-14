package component

type WorkQueue struct {
	queue  []any
	signal chan any
}

func (q *WorkQueue) Init() {
	q.queue = make([]any, 0)
	q.signal = make(chan any)
}

func (q *WorkQueue) Pop() {
	q.queue = q.queue[1:]
}

func (q *WorkQueue) Push(obj any) {
	q.queue = append(q.queue, obj)
	select {
	case q.signal <- obj:
	default:
	}
}

func (q *WorkQueue) Top() any {
	return q.queue[0]
}

func (q *WorkQueue) Empty() bool {
	return len(q.queue) == 0
}

func (q *WorkQueue) Fetch() any {
	if q.Empty() {
		<-q.signal
	}

	obj := q.Top()
	q.Pop()
	return obj
}

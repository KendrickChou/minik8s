package component

import "sync"

type set map[any]struct{}

type WorkQueue struct {
	// queue defines the order in which we will work on items. Every
	// element of queue should be in the dirty set and not in the
	// processing set.
	queue []any

	dirty set

	// Things that are currently being processed are in the processing set.
	// These things may be simultaneously in the dirty set. When we finish
	// processing something and remove it from this set, we'll check if
	// it's in the dirty set, and if so, add it to the queue.
	processing set

	cond *sync.Cond
}

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

func (s set) len() int {
	return len(s)
}

func (q *WorkQueue) Add(obj any) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.dirty.has(obj) {
		return
	}
	q.dirty.insert(obj)

	if q.processing.has(obj) {
		return
	}
	q.queue = append(q.queue, obj)
	q.cond.Signal()
}

func (q *WorkQueue) Get() (obj any) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for len(q.queue) == 0 {
		q.cond.Wait()
	}

	obj, q.queue = q.queue[0], q.queue[1:]
	q.processing.insert(obj)
	q.dirty.delete(obj)

	return obj
}

func (q *WorkQueue) Done(obj any) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.processing.delete(obj)
	if q.dirty.has(obj) {
		q.queue = append(q.queue, obj)
		q.cond.Signal()
	}
}

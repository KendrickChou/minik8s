package component

import "sync"

type ThreadSafeStore struct {
	lock  sync.RWMutex
	items map[string]any
}

func (store *ThreadSafeStore) Init() {
	store.items = map[string]any{}
}

func (store *ThreadSafeStore) Add(key string, obj any) {
	store.Update(key, obj)
}

func (store *ThreadSafeStore) Update(key string, obj any) {
	store.lock.Lock()
	defer store.lock.Unlock()
	store.items[key] = obj
}

func (store *ThreadSafeStore) Delete(key string) {
	store.lock.Lock()
	defer store.lock.Unlock()
	delete(store.items, key)
}

func (store *ThreadSafeStore) Get(key string) (item any, exists bool) {
	store.lock.RLock()
	defer store.lock.RUnlock()
	item, exists = store.items[key]
	return item, exists
}

func (store *ThreadSafeStore) List() []any {
	store.lock.RLock()
	defer store.lock.RUnlock()
	list := make([]any, 0, len(store.items))
	for _, item := range store.items {
		list = append(list, item)
	}
	return list
}

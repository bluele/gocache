package gocache

import (
	"errors"
	"sync"
	"time"
)

type CacheInterface interface {
	Get(interface{}) (interface{}, error)
	Set(interface{}, interface{})
	SetWithExpiration(interface{}, time.Duration)
	Delete(string) bool
}

type Cache struct {
	option    *Option
	mutex     sync.RWMutex
	items     map[interface{}]*item
	itemsPool chan *item
}

type Option struct {
	MaxPoolSize int
}

type item struct {
	value      interface{}
	expiration *time.Time
}

func New(opt *Option) *Cache {
	if opt == nil {
		opt = DefaultOption()
	}
	return &Cache{
		option:    opt,
		items:     make(map[interface{}]*item),
		itemsPool: make(chan *item, opt.MaxPoolSize),
	}
}

func DefaultOption() *Option {
	return &Option{
		MaxPoolSize: 32,
	}
}

var NotFoundError = errors.New("Not found key")

func (cc *Cache) Get(key interface{}) (interface{}, error) {
	cc.mutex.RLock()
	it, ok := cc.items[key]
	cc.mutex.RUnlock()
	if !ok {
		return nil, NotFoundError
	}
	if it.expiration != nil && it.expiration.Before(time.Now()) {
		cc.mutex.Lock()
		defer cc.mutex.Unlock()
		delete(cc.items, key)
		cc.returnItem(it)
		return nil, NotFoundError
	}
	return it.value, nil
}

func (cc *Cache) Set(key, val interface{}) {
	cc.set(key, val, nil)
}

func (cc *Cache) SetWithExpiration(key, val interface{}, expiration time.Duration) {
	cc.set(key, val, &expiration)
}

func (cc *Cache) GetOrSet(key interface{}, valFunc func() interface{}) interface{} {
	val, err := cc.Get(key)
	if err == nil {
		return val
	}
	val = valFunc()
	cc.set(key, val, nil)
	return val
}

func (cc *Cache) GetOrSetWithExpiration(key interface{}, valFunc func() interface{}, expiration time.Duration) interface{} {
	val, err := cc.Get(key)
	if err == nil {
		return val
	}
	val = valFunc()
	cc.set(key, val, &expiration)
	return val
}

func (cc *Cache) Exists(key interface{}) bool {
	cc.mutex.RLock()
	cc.mutex.RUnlock()
	_, ok := cc.items[key]
	return ok
}

func (cc *Cache) Delete(key interface{}) {
	cc.del(key)
}

func (cc *Cache) del(key interface{}) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	it, ok := cc.items[key]
	if !ok {
		return
	}
	delete(cc.items, key)
	cc.returnItem(it)
}

// returns true if can push old item to itemsPool.
func (cc *Cache) returnItem(it *item) bool {
	it.expiration = nil
	it.value = nil
	select {
	case cc.itemsPool <- it:
		return true
	default:
		return false
	}
}

func (cc *Cache) set(key, val interface{}, expiration *time.Duration) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	var it *item
	select {
	case it = <-cc.itemsPool:
	default:
		it = &item{}
	}

	if expiration == nil {
		it.expiration = nil
	} else {
		exp := time.Now().Add(*expiration)
		it.expiration = &exp
	}
	it.value = val

	cc.items[key] = it
}

func (cc *Cache) Clear() {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	for _, it := range cc.items {
		if !cc.returnItem(it) {
			break
		}
	}
	cc.items = make(map[interface{}]*item)
}

func (cc *Cache) Size() int {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	return len(cc.items)
}

func (cc *Cache) PoolSize() int {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	return len(cc.itemsPool)
}

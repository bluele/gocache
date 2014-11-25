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
	gc()
}

type Cache struct {
	option    *Option
	mutex     sync.RWMutex
	items     map[interface{}]*item
	itemsPool chan *item
}

type Option struct {
	MaxSize     int64
	MaxPoolSize int64
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

func (cc *Cache) returnItem(it *item) {
	it.expiration = nil
	it.value = nil
	cc.itemsPool <- it
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

	cc.items = make(map[interface{}]*item)
}

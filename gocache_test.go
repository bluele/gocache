package gocache_test

import (
	"github.com/bluele/gocache"
	"testing"
	"time"
)

func newCache(opt *gocache.Option) *gocache.Cache {
	return gocache.New(opt)
}

func TestSetGet(t *testing.T) {
	cc := newCache(nil)
	ek := "key"
	ev := "value"
	cc.Set(ek, ev)
	v, err := cc.Get(ek)
	if err != nil {
		t.Errorf("Not found: %v", ek)
	}
	if v != ev {
		t.Errorf("`%v` != `%v`", ev, v)
	}
}

func TestExpiration(t *testing.T) {
	cc := newCache(&gocache.Option{
		MaxPoolSize: 1,
	})
	ek := "key"
	ev := "value"
	cc.SetWithExpiration(ek, ev, time.Second)
	time.Sleep(1 * time.Second)
	_, err := cc.Get(ek)
	if err == nil {
		t.Errorf("Found: %v", ek)
	}
}

func TestDelete(t *testing.T) {
	cc := newCache(nil)
	ek := "key"
	ev := "value"
	cc.Set(ek, ev)
	cc.Delete("key")
	_, err := cc.Get(ek)
	if err == nil {
		t.Errorf("Found: %v", ek)
	}
}

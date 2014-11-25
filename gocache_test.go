package gocache_test

import (
	"fmt"
	"github.com/bluele/gocache"
	"sync"
	"testing"
	"time"
)

func newCache(opt *gocache.Option) *gocache.Cache {
	return gocache.New(opt)
}

func TestSetGet(t *testing.T) {
	cc := newCache(nil)

	var wait sync.WaitGroup
	counter := 10000

	for i := 0; i < counter; i++ {
		s := fmt.Sprintf("%d", i)
		wait.Add(1)
		go func() {
			cc.Set(s, s)
			wait.Done()
		}()
	}
	wait.Wait()

	for i := 0; i < counter; i++ {
		if i%2 == 0 {
			s := fmt.Sprintf("%d", i)
			cc.Delete(s)
		}
	}

	if cc.Size() != counter/2 {
		t.Errorf("Size should returns: %v", counter/2)
	}

	for i := 0; i < counter; i++ {
		if i%2 != 0 {
			s := fmt.Sprintf("%d", i)
			v, err := cc.Get(s)
			if err != nil {
				t.Errorf("Not found key: %v", s)
			}
			if v != s {
				t.Errorf("Expected value: %v, not %v", s, v)
			}
		}
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

func TestGetOrSet(t *testing.T) {
	cc := newCache(nil)
	v := cc.GetOrSet("key", func() interface{} {
		return "value"
	})
	if v != "value" {
		t.Errorf("Expected value: value")
	}
}

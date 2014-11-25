package main

import (
	"github.com/bluele/gocache"
	"log"
	"time"
)

func main() {
	gc := gocache.New(nil)
	gc.SetWithExpiration("key", "value", 2*time.Second)
	log.Println(gc.Get("key"))
	time.Sleep(2 * time.Second)
	log.Println(gc.Get("key"))
}

# gocache
Cache module for golang.

# Example

```go
func main() {
  gc := gocache.New(nil)
  gc.SetWithExpiration("key", "value", 2*time.Second)
  // 2014/11/25 16:50:44 value <nil>
  log.Println(gc.Get("key"))
  time.Sleep(2 * time.Second)
  // 2014/11/25 16:50:46 <nil> Not found key
  log.Println(gc.Get("key"))
}
```

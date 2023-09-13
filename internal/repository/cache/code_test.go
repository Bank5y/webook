package cache

import (
	"sync"
	"testing"
)

func TestLocalCache(t *testing.T) {
	var wg sync.WaitGroup

	biz := "login"
	phone := "18384242"
	code := "100000"
	cache := NewCodeLocalCache()
	cache.Set(nil, biz, phone, code)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ok, _ := cache.Verify(nil, biz, phone, code)
			count := i
			println(count)
			println(ok)
		}(i)
	}
}

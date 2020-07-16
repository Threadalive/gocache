package project

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCount := make(map[string]int, len(db))
	gocache := NewGroup("score", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Printf("slow db search")
		if v, ok := db[key]; ok {
			if _, ok := loadCount[key]; !ok {
				loadCount[key] = 0
			}
			loadCount[key]++
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	for k, v := range db {
		if view, err := gocache.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		if _, err := gocache.Get(k); err != nil || loadCount[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := gocache.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}

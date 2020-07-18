package project

import (
	"fmt"
	"log"
	"sync"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name: name,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
		getter: getter,
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	if v, ok := groups[name]; ok {
		g := v
		return g
	} else {
		return nil
	}
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key can not be empty")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Printf("cache hit")
		return v, nil
	}
	//缓存未命中，加载数据库数据
	return g.load(key)
}

//加载数据
func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

//通过回调函数获取本地数据
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)

	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.addCache(key, value)

	return value, nil
}

//缓存未命中时，添加该缓存
func (g *Group) addCache(key string, value ByteView) {
	g.mainCache.set(key, value)
}

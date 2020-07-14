package lru

import "container/list"

type Cache struct {
	//缓存允许使用的最大内存,int64相当于long
	maxBytes int64
	//当前已使用的内存
	nBytes int64
	//双向链表
	ll *list.List
	//map映射，指向链表节点
	cache map[string]*list.Element
	//移除缓存
	onEvicted func(key string, value Value)
}

type node struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

//构造一个缓存实例
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (cache *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := cache.cache[key]; ok {
		//将节点移到队首
		cache.ll.MoveToFront(element)
		kv := element.Value.(*node)
		return kv.value, ok
	}
	return
}

func (cache *Cache) RemoveOldest() {
	//取到队尾节点
	tail := cache.ll.Back()
	if tail != nil {
		//删除节点
		cache.ll.Remove(tail)
		kv := tail.Value.(*node)
		//删除map映射中的缓存
		delete(cache.cache, kv.key)
		//计算剩余内存
		cache.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		//若回调函数非空，则执行回调函数
		if cache.onEvicted != nil {
			cache.onEvicted(kv.key, kv.value)
		}
	}
}

func (cache *Cache) Set(key string, value Value) {
	if element, ok := cache.cache[key]; ok {
		cache.ll.MoveToFront(element)
		kv := element.Value.(*node)
		//加上新value内存，减去原value内存
		cache.nBytes += (int64(value.Len()) - int64(kv.value.Len()))
		kv.value = value
	} else {
		//若缓存中还没有该节点，插入一个
		element := cache.ll.PushFront(&node{key: key, value: value})
		//添加映射
		cache.cache[key] = element
		cache.nBytes += int64(len(key)) + int64(value.Len())
	}
	for cache.maxBytes != 0 && cache.nBytes > cache.maxBytes {
		cache.RemoveOldest()
	}
}

func (cache *Cache) Len() int {
	return cache.ll.Len()
}

package main

import (
	"fmt"
	"github.com/Threadalive/gocache/project"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

//
//var m sync.Mutex
//
//var set = make(map[int]bool)
//
//func printOnce(num int) {
//	m.Lock()
//	defer m.Unlock()
//	if _, exist := set[num]; !exist {
//		fmt.Println(num)
//	}
//	set[num] = true
//}
func main() {
	//for i := 0; i < 10; i++ {
	//	go printOnce(100)
	//}
	//time.Sleep(time.Second)

	project.NewGroup("score", 2<<10, project.GetterFunc(func(key string) ([]byte, error) {
		log.Printf("slow db search")
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	addr := "localhost:9999"
	peers := project.NewHttpPool(addr)

	log.Println("gocache is running at", addr)
	//在9999端口启动服务器
	log.Fatal(http.ListenAndServe(addr, peers))
}

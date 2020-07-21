package main

import (
	"flag"
	"fmt"
	"github.com/Threadalive/gocache/project"
	"github.com/Threadalive/gwf"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

//创建新分组
func createGroup(name string) *project.Group {
	return project.NewGroup(name, 2<<10, project.GetterFunc(func(key string) ([]byte, error) {
		log.Printf("slow db search")
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

//注册与启动服务器节点
func startCacheServer(addr string, addrs []string, group *project.Group) {
	peers := project.NewHttpPool(addr)
	//在HttpPool上注册节点地址
	peers.Set(addrs...)
	//将节点注入group
	group.RegisterPeers(peers)
	log.Println("gocache is running at", addr)
	//传入地址格式如下 "http://localhost:9999"
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

//前台交互模块
func startApiServer(apiAddr string, group *project.Group) {
	//http.HandleFunc("api", func(w http.ResponseWriter, r *http.Request) {
	//
	//})
	r := gwf.New()
	r.GET("/api", func(c *gwf.Context) {
		key := c.Query("key")
		view, err := group.Get(key)
		if err != nil {
			c.Fail(http.StatusInternalServerError, err.Error())
			return
		}
		//若获取数据成功
		c.SetHeader("Content-Type", "application/octet-stream")
		c.Json(http.StatusOK, view.ByteSlice())

	})
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], r))
}

//
//var m sync.Mutex
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
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Gocache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup("score")
	if api {
		go startApiServer(apiAddr, group)
	}
	//注册与启动缓存服务器
	startCacheServer(addrMap[port], addrs, group)

	//for i := 0; i < 10; i++ {
	//	go printOnce(100)
	//}
	//time.Sleep(time.Second)

	//project.NewGroup("score", 2<<10, project.GetterFunc(func(key string) ([]byte, error) {
	//	log.Printf("slow db search")
	//	if v, ok := db[key]; ok {
	//		return []byte(v), nil
	//	}
	//	return nil, fmt.Errorf("%s not exist", key)
	//}))

	//addr := "localhost:9999"
	//peers := project.NewHttpPool(addr)
	//
	//log.Println("gocache is running at", addr)
	////在9999端口启动服务器
	//log.Fatal(http.ListenAndServe(addr, peers))
}

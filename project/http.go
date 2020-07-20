package project

import (
	"fmt"
	"github.com/Threadalive/gocache/project/consistent_hash"
	"github.com/Threadalive/gwf"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_gocache/"
	defaultReplicas = 50
)

//承载http通信的结构
type HttpPool struct {
	self     string
	basePath string
	mu       sync.Mutex
	//服务器节点的哈希环
	peers *consistent_hash.Map
	//记录远程节点与其对应的httpGetter
	httpGetter map[string]*HttpGetter
}

//http客户端类
type HttpGetter struct {
	baseURL string
}

//新建一个HttpPool
func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

//注册服务器节点
func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	//根据默认的副本数和默认哈希算法新建哈希结构
	p.peers = consistent_hash.New(defaultReplicas, nil)
	//添加服务器节点
	p.peers.AddNodes(peers...)
	p.httpGetter = make(map[string]*HttpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetter[peer] = &HttpGetter{baseURL: peer + p.basePath}
	}
}

//根据key选取所需节点
func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetter[peer], true
	}
	return nil, false
}

//添加节点的服务端操作
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := gwf.NewContext(w, r)
	if !strings.HasPrefix(context.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + context.Path)
	}
	p.Log("%s %s", context.Method, context.Path)

	// 请求示例：<basepath>/<groupname>/<key>
	parts := strings.SplitN(context.Path[len(p.basePath):], "/", 2)

	if len(parts) != 2 {
		context.Fail(http.StatusBadRequest, "bad request")
		//http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	//若分组不存在
	if group == nil {
		context.Fail(http.StatusNotFound, "no such group "+groupName)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		context.Fail(http.StatusInternalServerError, err.Error())
	}
	context.SetHeader("Content-Type", "application/octet-stream")
	context.String(http.StatusOK, string(view.ByteSlice()))
}

func (h *HttpGetter) Get(group string, key string) ([]byte, error) {
	acceptUrl := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))

	resp, err := http.Get(acceptUrl)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}

	bytes, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

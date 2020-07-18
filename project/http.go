package project

import (
	"fmt"
	"github.com/Threadalive/gwf"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_gocache/"

//承载http通信的结构
type HttpPool struct {
	self     string
	basePath string
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

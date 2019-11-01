package common

import (
	"net/http"
)

// 声明一个函数类型
type FilterHandle func(rw http.ResponseWriter, r *http.Request) error

// 拦截器结构体
type Filter struct {
	// 用来存储要拦截的url
	filterMap map[string]FilterHandle
}

func NewFilter() *Filter {
	return &Filter{filterMap: make(map[string]FilterHandle)}
}

// 注册拦截器
func (f *Filter) RegisterFilterUri(uri string, handler FilterHandle) {
	f.filterMap[uri] = handler
}

// 根据uri获取对应handle
func (f *Filter) GetFilterHandle(uri string) FilterHandle {
	return f.filterMap[uri]
}

// 定义一个新的函数类型
type WebHandle func(rw http.ResponseWriter, r *http.Request)

// handle方法，拦截器实例通过调用handle，传入一个函数，返回一个新函数。
// 确切的说，传入的函数被改装了，增加了拦截器实例内的路径和方法执行，如果请求的路径是一致的，则直接使用拦截器内的方法处理
// 如果路径不一致，则使用传入函数的逻辑处理
func (f *Filter) Handle(webHandle WebHandle) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		for path, handle := range f.filterMap {
			if path == r.RequestURI {
				err := handle(rw, r)
				if err != nil {
					rw.Write([]byte(err.Error()))
					return
				}
				break
			}
		}
		// 执行正常注册的函数
		webHandle(rw, r)
	}
}

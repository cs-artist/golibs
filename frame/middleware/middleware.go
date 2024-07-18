// 中间件实现
/*
	r = NewRouter()
	r.Use(logger)
	r.Use(timeout)
	r.Use(ratelimit)
	r.Add("/", helloHandler)

	执行顺序：logger -> timeout -> ratelimit -> helloHandler -> ratelimit -> timeout -> logger
*/
package middleware

import "net/http"

type middleware func(http.Handler) http.Handler

type Router struct {
	chain []middleware
	mux   map[string]http.Handler
}

func NewRouter() *Router {
	return &Router{mux: make(map[string]http.Handler)}
}

func (r *Router) Use(m middleware) {
	r.chain = append(r.chain, m)
}

func (r *Router) Add(route string, h http.Handler) {
	var mergedHandler = h
	for i := len(r.chain) - 1; i >= 0; i-- {
		mergedHandler = r.chain[i](mergedHandler)
	}
	r.mux[route] = mergedHandler
}

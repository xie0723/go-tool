package functions

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/astaxie/beego/logs"
)

type handle struct {
	host string
	port string
}

func (h *handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remote, err := url.Parse("http://" + h.host + ":" + h.port)
	if err != nil {
		logs.Error("Parse url failed,", err.Error())
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}

// 本地端口映射
func LocalPortMap(src int, dest int, ctx context.Context) (err error) {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", src)}
	srv.Handler = &handle{host: "127.0.0.1", port: fmt.Sprintf("%d", dest)}
	go srv.ListenAndServe()
	<-ctx.Done()
	srv.Close()
	srv.Shutdown(context.TODO())
	return
}

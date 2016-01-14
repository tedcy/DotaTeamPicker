package main

import (
	"net/http"
	"net"
	"strings"
	"time"
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"fmt"
)

type Overview struct {
	Version string
	Compile string
}

type apiServer struct {
	Version string
	Compile string
}

func (s *apiServer) overview() (int, string) {
	b, _ := json.Marshal(&Overview{
		Version: s.Version,
		Compile: s.Compile,
	})
	return 200, string(b)
}

func newApiServer(o *Overview) http.Handler{
	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(render.Renderer())
	m.Use(func(w http.ResponseWriter, req *http.Request, c martini.Context) {
		path := req.URL.Path
		if strings.HasPrefix(path, "/overview") {
			var remoteAddr = req.RemoteAddr
			var headerAddr string
			for _, key := range []string{"X-Real-IP", "X-Forwarded-For"} {
				if val := req.Header.Get(key); val != "" {
					headerAddr = val
					break
				}
            }
			fmt.Printf("API call %s from %s [%s]\n", path, remoteAddr, headerAddr)
        }
		c.Next()
    })
	api := &apiServer{Version: o.Version,Compile: o.Compile}
	m.Use(func(c martini.Context, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
    })
	r := martini.NewRouter()
	r.Get("/", func(r render.Render) {
		r.Redirect("/overview")
    })
	r.Get("/overview", api.overview)
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return m
}

func serve() error{
	l, err := net.Listen("tcp", "172.17.140.76:8080")
	if err != nil {
		return err
    }
	eh := make(chan error, 1)
	go func(l net.Listener) {
		h := http.NewServeMux()
		h.Handle("/", newApiServer(&Overview{Version: "1.00", Compile: "go"}))
		hs := &http.Server{Handler: h}
		eh <- hs.Serve(l)
	}(l)
	err = <-eh
	return err
}

func main() {
	err := serve()
	fmt.Printf("%v\n",err)
	for {
		time.Sleep(time.Second)
    }
}

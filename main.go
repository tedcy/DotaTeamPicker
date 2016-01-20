package main

import (
	"fmt"
	"github.com/tedcy/DotaTeamPicker/picker"
	"net"
	"net/http"
	"time"
)

func serve() error {
	l, err := net.Listen("tcp", picker.ConfigData.Addr)
	if err != nil {
		return err
	}
	eh := make(chan error, 1)
	go func(l net.Listener) {
		h := http.NewServeMux()
		h.Handle("/", picker.NewApiServer())
		hs := &http.Server{Handler: h}
		eh <- hs.Serve(l)
	}(l)
	err = <-eh
	return err
}

func main() {
	picker.InitHeroIdMap()
	picker.LoadConfig()
	err := serve()
	fmt.Printf("%v\n", err)
	for {
		time.Sleep(time.Second)
	}
}

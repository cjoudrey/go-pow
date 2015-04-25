package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type HttpServer struct {
	Port         int
	Applications []*Application
}

func (s *HttpServer) Listen() error {
	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%v", s.Port),
		Handler: s,
	}
	err := server.ListenAndServe()
	return err
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s %s %s", r.RemoteAddr, r.Method, r.Host, r.URL)

	var requestApplication *Application
	for _, application := range s.Applications {
		if application.Host == r.Host {
			requestApplication = application
			break
		}
	}

	if requestApplication == nil {
		s.handleApplicationNotFound(w, r)
		return
	}

	filePath := fmt.Sprintf("%s/public/%s", requestApplication.Dir, r.URL.Path[1:])
	if fileInfo, err := os.Stat(filePath); err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, filePath)
		return
	}

	// @TODO Proxy request to application
}

func (s *HttpServer) handleApplicationNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("Application <b>%s</b> does not exist.", r.Host)))
}

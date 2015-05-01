package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type HttpServer struct {
	Configuration *Configuration
	Applications  map[string]*Application
}

func NewHttpServer(configuration *Configuration) *HttpServer {
	server := HttpServer{
		Configuration: configuration,
		Applications:  make(map[string]*Application),
	}
	return &server
}

func (s *HttpServer) Listen() error {
	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%v", s.Configuration.Port),
		Handler: s,
	}
	err := server.ListenAndServe()
	return err
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] %s %s %s", r.RemoteAddr, r.Method, r.Host, r.URL)

	hosts, err := s.Configuration.Hosts()
	if err != nil {
		s.handleInternalServerError(err, w, r)
		return
	}

	hostExtractor := regexp.MustCompile("(.*)\\.(" + strings.Join(s.Configuration.Domains, "|") + "):[0-9]+")
	hostExtractorMatches := hostExtractor.FindStringSubmatch(r.Host)

	var host string
	if len(hostExtractorMatches) < 2 {
		host = ""
	} else {
		host = hostExtractorMatches[1]
	}

	// @TODO cleanup s.Applications[...] if app no longer exists

	if _, ok := hosts[host]; !ok {
		s.handleApplicationNotFound(w, r)
		return
	}

	// @TODO Apps that do not have a `config.ru` should fail gracefully

	if _, ok := s.Applications[hosts[host].Root]; !ok {
		s.Applications[hosts[host].Root] = NewApplication(hosts[host].Root)
	}

	filePath := fmt.Sprintf("%s/public/%s", hosts[host].Root, r.URL.Path[1:])
	if fileInfo, err := os.Stat(filePath); err == nil && !fileInfo.IsDir() {
		http.ServeFile(w, r, filePath)
		return
	}

	s.Applications[hosts[host].Root].HandleRequest(w, r) // @TODO handle err
}

func (s *HttpServer) handleApplicationNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("Application <b>%s</b> does not exist.", r.Host)))
}

func (s *HttpServer) handleInternalServerError(err error, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("Internal server error: %s", err)))
}

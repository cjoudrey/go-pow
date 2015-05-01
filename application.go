package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Application struct {
	Root      string
	state     string
	port      int
	stdout    io.Writer
	stderr    io.Writer
	cmd       *exec.Cmd
	mutex     *sync.Mutex
	transport *http.Transport
}

func NewApplication(root string) *Application {
	application := Application{
		Root:      root,
		port:      9999,      // @TODO auto-detect
		stdout:    os.Stdout, // @TODO temporary
		stderr:    os.Stdout, // @TODO temporary
		mutex:     &sync.Mutex{},
		transport: &http.Transport{},
	}

	return &application
}

func (a *Application) Launch() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.state == "ready" || a.state == "launching" {
		return nil
	}

	a.state = "launching"

	a.cmd = exec.Command("rbenv", "exec", "bundle", "exec", "rackup", "-p", strconv.Itoa(a.port))
	a.cmd.Dir = a.Root

	a.cmd.Stdout = a.stdout
	a.cmd.Stderr = a.stderr

	if err := a.cmd.Start(); err != nil {
		a.state = ""
		log.Printf("failed to start %s", err)
		return err
	}

	if err := a.checkServerIsResponding(); err != nil {
		a.state = ""
		// @TODO kill a.cmd.Process
		log.Printf("can't connect to %s", err)
		return err
	}

	a.state = "ready"

	go func() {
		// @TODO need to let the user know the app crashed
		if err := a.cmd.Wait(); err != nil {
			a.state = ""

			if exiterr, ok := err.(*exec.ExitError); ok {
				log.Printf("exiterr %s", exiterr)
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					log.Printf("Exit Status: %d", status.ExitStatus())
				}
			} else {
				log.Printf("wait err %s", err)
			}
		}
	}()

	return nil
}

func (a *Application) Terminate() error {
	if a.state != "ready" {
		return nil
	}

	a.state = "terminating"

	if err := a.cmd.Process.Signal(os.Kill); err != nil {
		return err
	}

	a.state = ""

	return nil
}

func (a *Application) Restart() error {
	if err := a.Terminate(); err != nil {
		return err
	}

	if err := a.Launch(); err != nil {
		return err
	}

	return nil
}

func (a *Application) checkServerIsResponding() error {
	// @TODO Timeout
	for {
		_, err := net.Dial("tcp", fmt.Sprintf(":%d", a.port))

		if err == nil {
			break
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}

	return nil
}

func (a *Application) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	if err := a.Launch(); err != nil {
		return err
	}

	r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Host = fmt.Sprintf("%s:%d", "127.0.0.1", a.port)

	resp, err := a.transport.RoundTrip(r)
	if err != nil {
		// @TODO this should cause a nice error page to show up
		return err
	}

	// @TODO better error handling
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
	resp.Body.Close()

	return nil
}

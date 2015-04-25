package main

import (
	"os"
	"os/exec"
	"syscall"
	"strconv"
	"log"
	"io"
)

type Application struct {
	Dir  string
	Host string
	//	LaunchCmd string // @TODO
	State string
	Port int
	Stdout io.Writer
	Stderr io.Writer
	cmd *exec.Cmd
}

func (a *Application) Launch() error {
	a.State = "launching"

	a.cmd = exec.Command("rbenv", "exec", "rackup", "-p", strconv.Itoa(a.Port))
	a.cmd.Dir = a.Dir

	if a.Stdout != nil {
		a.cmd.Stdout = a.Stdout
	}

	if a.Stderr != nil {
		a.cmd.Stderr = a.Stderr
	}

	if err := a.cmd.Start(); err != nil {
		a.State = ""
		log.Printf("failed to start %s", err)
		return err
	}

	a.State = "ready"

	if err := a.cmd.Wait(); err != nil {
		a.State = ""

		if exiterr, ok := err.(*exec.ExitError); ok {
			log.Printf("exiterr %s", exiterr)
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit Status: %d", status.ExitStatus())
			}
		} else {
			log.Printf("wait err %s", err)
			return err
		}
	}

	return nil
}

func (a *Application) Terminate() error {
	if a.State != "ready" {
		return nil
	}

	a.State = "terminating"

	err := a.cmd.Process.Signal(os.Kill)
	return err
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

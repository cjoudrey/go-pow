package main

import (
	"os/exec"
	"strconv"
)

type Application struct {
	Dir  string
	Host string
	//	LaunchCmd string // @TODO
	//	State string // @TODO
	Port int
}

func (a *Application) Launch() error {
	cmd := exec.Command("rbenv", "exec", "rackup", "-p", strconv.Itoa(a.Port))
	cmd.Dir = a.Dir

	err := cmd.Run() // @TODO Might want to use Start() and wait() in a goroutine
	return err
}

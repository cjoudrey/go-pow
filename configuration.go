package main

import (
	"io/ioutil"
	_ "log"
)

type Configuration struct {
	HostsRoot string
	Port      int
	DnsPort   int
	Domains   []string
}

type Host struct {
	Root string
	Url  string
}

func (c *Configuration) Hosts() (map[string]Host, error) {
	hosts := make(map[string]Host)

	files, err := ioutil.ReadDir(c.HostsRoot)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			hosts[file.Name()] = Host{Root: c.HostsRoot + "/" + file.Name()}
		}
	}

	return hosts, nil
}

package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

type Configuration struct {
	HostRoot string
	Port     int
	DnsPort  int
	Domains  []string
}

type Host struct {
	Root string
	Url  string
}

func NewConfiguration() *Configuration {
	configuration := Configuration{}

	if envHostRoot := os.Getenv("POW_HOST_ROOT"); envHostRoot != "" {
		configuration.HostRoot = envHostRoot
	} else {
		// @TODO Use ~/Library/Application Support/Pow/Hosts
	}

	if envPort := os.Getenv("POW_PORT"); envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil {
			configuration.Port = port
		} else {
			log.Printf("Invalid POW_PORT: %s", err)
		}
	}

	if configuration.Port == 0 {
		configuration.Port = 8080
	}

	if envDnsPort := os.Getenv("POW_DNS_PORT"); envDnsPort != "" {
		if dnsPort, err := strconv.Atoi(envDnsPort); err == nil {
			configuration.DnsPort = dnsPort
		} else {
			log.Printf("Invalid POW_DNS_PORT: %s", err)
		}
	}

	if configuration.DnsPort == 0 {
		configuration.DnsPort = 25083
	}

	if envDomains := os.Getenv("POW_DOMAINS"); envDomains != "" {
		configuration.Domains = strings.Split(envDomains, ",")
	} else {
		configuration.Domains = []string{"dev"}
	}

	return &configuration
}

func (c *Configuration) Hosts() (map[string]Host, error) {
	hosts := make(map[string]Host)

	files, err := ioutil.ReadDir(c.HostRoot)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			hosts[file.Name()] = Host{Root: c.HostRoot + "/" + file.Name()}
		}
	}

	return hosts, nil
}

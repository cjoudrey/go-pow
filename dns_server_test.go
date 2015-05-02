package main

import (
	"os/exec"
	"regexp"
	"strconv"
	"testing"
)

var port = 8000

func SetupServer() {
	configuration := Configuration{
		DnsPort: port,
		Domains: []string{"powdev", "powtest"},
	}

	server := NewDnsServer(&configuration)

	go server.Listen()
}

func Resolve(domain string) (error, string, string) {
	out, err := exec.Command("dig", "-p", strconv.Itoa(port), "@127.0.0.1", domain, "+noall", "+answer", "+comments").Output()
	if err != nil {
		return err, "", ""
	}

	output := string(out)

	statusRe := regexp.MustCompile("status: (.*?),")
	answerRe := regexp.MustCompile("IN\tA\t([\\d.]+)")

	status := statusRe.FindStringSubmatch(output)[1]

	answer := answerRe.FindStringSubmatch(output)
	if len(answer) == 0 {
		return nil, status, ""
	}

	return nil, status, answer[1]
}

var resolveTests = []struct {
	domain string
	status string
	answer string
}{
	{"hello.powtest", "NOERROR", "127.0.0.1"},
	{"hello.powdev", "NOERROR", "127.0.0.1"},
	{"a.b.c.powtest", "NOERROR", "127.0.0.1"},
	{"powtest.", "NOERROR", "127.0.0.1"},
	{"powdev.", "NOERROR", "127.0.0.1"},
	{"foo.", "NXDOMAIN", ""},
}

func TestRespondsToDomains(t *testing.T) {
	SetupServer()

	for _, tt := range resolveTests {
		err, status, answer := Resolve(tt.domain)

		if err != nil {
			t.Errorf("Failed to resolve %s got error %s", tt.domain, err)
		} else if status != tt.status {
			t.Errorf("Failed to resolve %s wanted status %s got %s", tt.domain, tt.status, status)
		} else if answer != tt.answer {
			t.Errorf("Failed to resolve %s wanted answer %s got %s", tt.domain, tt.answer, answer)
		}
	}
}

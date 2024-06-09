package probe

import (
	"fmt"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/guarandoo/neko/pkg/core"
)

type sshProbe struct {
	host            string
	port            int
	hostKeyCallback ssh.HostKeyCallback
}

func (p *sshProbe) Probe() (*core.Result, error) {
	ips, err := net.LookupIP(p.host)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup domain: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            "",
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: p.hostKeyCallback,
	}

	tests := []core.Test{}
	for _, ip := range ips {
		test := core.Test{Target: ip.String(), Status: core.StatusUp}

		client, err := ssh.Dial("tcp", fmt.Sprintf("[%v]:%v", ip, p.port), config)
		if err != nil {
			if !strings.Contains(err.Error(), "ssh: unable to authenticate") {
				test.Status = core.StatusDown
				test.Error = err
				tests = append(tests, test)
				continue
			}
		}
		if client != nil {
			client.Close()
		}

		tests = append(tests, test)
	}

	return &core.Result{Tests: tests}, nil
}

type SshProbeOptions struct {
	Host    string
	Port    int
	HostKey *string
}

func NewSshProbe(options SshProbeOptions) (Probe, error) {
	var hostKeyCallback ssh.HostKeyCallback

	if options.HostKey == nil {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		publicKey, err := ssh.ParsePublicKey([]byte(*options.HostKey))
		if err != nil {
			return nil, fmt.Errorf("unable to create ssh probe: %w", err)
		}
		hostKeyCallback = ssh.FixedHostKey(publicKey)
	}

	return &sshProbe{
		host:            options.Host,
		port:            options.Port,
		hostKeyCallback: hostKeyCallback,
	}, nil
}

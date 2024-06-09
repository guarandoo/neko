package probe

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"

	"github.com/guarandoo/neko/pkg/core"
)

type sshProbe struct {
	host string
	port int
}

func (p *sshProbe) Probe() (*core.Result, error) {
	ips, err := net.LookupIP(p.host)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup domain: %w", err)
	}

	tests := []core.Test{}
	for _, ip := range ips {
		test := core.Test{Target: ip.String(), Status: core.StatusUp}

		var hostKey ssh.PublicKey

		config := &ssh.ClientConfig{
			User: "username",
			Auth: []ssh.AuthMethod{
				ssh.Password("pass"),
			},
			HostKeyCallback: ssh.FixedHostKey(hostKey),
		}
		client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", p.host, p.port), config)
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err
			tests = append(tests, test)
			continue
		}
		defer client.Close()

		tests = append(tests, test)
	}

	return &core.Result{Tests: tests}, nil
}

type SshProbeOptions struct {
	Host string
	Port int
}

func NewSshProbe(options SshProbeOptions) (Probe, error) {
	return &sshProbe{host: options.Host, port: options.Port}, nil
}

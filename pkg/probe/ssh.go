package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/guarandoo/neko/pkg/core"
)

const SshProbeType string = "ssh"

var onceInitSshProbe sync.Once

func initSshProbe() {

}

type sshProbe struct {
	host            string
	port            int
	hostKeyCallback ssh.HostKeyCallback
	user            string
	authMethods     []ssh.AuthMethod
}

func connectAndAuthenticate(ctx context.Context, host string, config *ssh.ClientConfig) (err error) {
	dialer := net.Dialer{}
	con, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return err
	}

	sshCon, _, _, err := ssh.NewClientConn(con, host, config)
	if err != nil {
		return err
	}
	defer func() {
		err = sshCon.Close()
	}()

	return
}

func (p *sshProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	ips, err := net.LookupIP(p.host)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup domain: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            p.user,
		Auth:            p.authMethods,
		HostKeyCallback: p.hostKeyCallback,
	}

	tests := []core.Test{}
	for _, ip := range ips {
		test := core.Test{Target: ip.String(), Status: core.StatusUp}

		host := fmt.Sprintf("[%v]:%v", ip, p.port)

		if err := connectAndAuthenticate(ctx, host, config); err != nil {
			test.Status = core.StatusDown
			test.Error = err
			tests = append(tests, test)
			continue
		}

		tests = append(tests, test)
	}

	return &core.Result{Tests: tests}, nil
}

const (
	SshPasswordAuth = "password"
	SshPubKeyAuth   = "pubkey"
)

type SshProbePasswordAuthMethodOptions struct {
	Password string
}

type SshProbeKeyAuthMethodOptions struct {
	PrivateKey []byte
}

type SshProbeAuthOptions struct {
	User   string
	Method any
}

type SshProbeOptions struct {
	ProbeOptions
	Host           string
	Port           int
	HostKey        string
	Authentication SshProbeAuthOptions
}

func NewSshProbe(options SshProbeOptions) (Probe, error) {
	onceInitSshProbe.Do(initSshProbe)

	var hostKeyCallback ssh.HostKeyCallback

	if len(options.HostKey) > 0 {
		publicKey, err := ssh.ParsePublicKey([]byte(options.HostKey))
		if err != nil {
			return nil, fmt.Errorf("unable to create ssh probe: %w", err)
		}
		hostKeyCallback = ssh.FixedHostKey(publicKey)
	} else {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	var authMethod ssh.AuthMethod
	if authOpts, ok := options.Authentication.Method.(SshProbePasswordAuthMethodOptions); ok {
		authMethod = ssh.Password(authOpts.Password)
	} else if authOpts, ok := options.Authentication.Method.(SshProbeKeyAuthMethodOptions); ok {
		signer, err := ssh.ParsePrivateKey(authOpts.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %w", err)
		}

		authMethod = ssh.PublicKeys(signer)
	} else {
		return nil, errors.New("unknown ssh auth method")
	}

	return &sshProbe{
		host:            options.Host,
		port:            options.Port,
		hostKeyCallback: hostKeyCallback,
		user:            options.Authentication.User,
		authMethods:     []ssh.AuthMethod{authMethod},
	}, nil
}

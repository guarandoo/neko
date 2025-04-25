package probe

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	sshTest "github.com/gliderlabs/ssh"
	"github.com/guarandoo/neko/pkg/core"
	"golang.org/x/crypto/ssh"
)

func generateKeyPair() (key *rsa.PrivateKey, err error) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key: %w", err)
	}

	return pk, nil
}

func TestSsh(t *testing.T) {
	var wg sync.WaitGroup

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("unable to initiate ssh listener: %v", err)
	}
	defer listener.Close()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("unable to get listener address")
	}

	user := "ssh_test"
	privateKey, err := generateKeyPair()

	server := sshTest.Server{
		Addr: fmt.Sprintf("%v:%v", tcpAddr.IP, tcpAddr.Port),
		Handler: func(s sshTest.Session) {
			io.WriteString(s, "Test!")
		},
		PublicKeyHandler: func(ctx sshTest.Context, pubkey sshTest.PublicKey) bool {
			expectedPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
			if err != nil {
				t.Fatalf("unable to extract public key from private key: %v", err)
			}

			return ctx.User() == user &&
				expectedPublicKey.Type() == pubkey.Type() &&
				string(expectedPublicKey.Marshal()) == string(pubkey.Marshal())
		},
	}

	go func() {
		defer wg.Done()
		wg.Add(1)
		err = server.Serve(listener)
		if !errors.Is(err, sshTest.ErrServerClosed) {
			t.Fatalf("unable to start ssh server: %v", err)
		}
	}()

	privDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDer,
	})

	probe, err := NewSshProbe(SshProbeOptions{
		Host: tcpAddr.IP.String(),
		Port: tcpAddr.Port,
		Authentication: SshProbeAuthOptions{
			User:   user,
			Method: SshProbeKeyAuthMethodOptions{PrivateKey: privPem},
		},
	})
	if err != nil {
		t.Fatalf("unable to create ssh probe: %v", err)
	}

	ctx, cancel := getContextWithTimeout(context.Background(), time.Second*30)
	defer cancel()

	res, err := probe.Probe(ctx)
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("unable to shutdown ssh server: %v", err)
	}
	wg.Wait()

	if err != nil {
		t.Fatalf("probe returned an error: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Fatalf("returned unexpected number of tests")
	}

	test := res.Tests[0]
	if test.Status != core.StatusUp {
		t.Fatalf("probe did not report target as up")
	}
}

package probe

import (
	"context"

	"github.com/guarandoo/neko/pkg/core"

	"github.com/cloudsoda/go-smb2"
)

const SmbProbeType string = "smb"

type smbProbe struct {
	host     string
	user     string
	password string
	share    string
}

func (p *smbProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	test := core.Test{
		Target: p.host,
		Status: core.StatusUp,
		Error:  nil,
		Extras: make(map[string]any),
	}

	dialer := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     p.user,
			Password: p.password,
		},
	}

	session, err := dialer.Dial(ctx, p.host)
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
		return &core.Result{Tests: []core.Test{test}}, nil
	}
	defer func() { session.Logoff() }()

	share, err := session.Mount(p.share)
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
		return &core.Result{Tests: []core.Test{test}}, nil
	}
	defer func() { share.Umount() }()

	info, err := share.Statfs(".")
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
		return &core.Result{Tests: []core.Test{test}}, nil
	}

	blockSize := info.BlockSize()
	totalBlocks := info.TotalBlockCount()
	freeBlocks := info.FreeBlockCount()
	availableBlocks := info.AvailableBlockCount()
	freeSpaceBytes := availableBlocks * blockSize

	test.Extras["block_size"] = blockSize
	test.Extras["total_blocks"] = totalBlocks
	test.Extras["free_blocks"] = freeBlocks
	test.Extras["available_blocks"] = availableBlocks
	test.Extras["free_space_bytes"] = freeSpaceBytes

	return &core.Result{Tests: []core.Test{test}}, nil
}

type SmbProbeOptions struct {
	ProbeOptions
	Host     string
	User     string
	Password string
	Share    string
}

func NewSmbProbe(options SmbProbeOptions) (Probe, error) {
	return &smbProbe{
		host:     options.Host,
		user:     options.User,
		password: options.Password,
		share:    options.Share,
	}, nil
}

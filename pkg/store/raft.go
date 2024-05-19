package store

import (
	"io"

	"github.com/hashicorp/raft"
)

type MyFsm struct {
	a int
}

func (m *MyFsm) Apply(*raft.Log) interface{} {
	return nil
}

func (m *MyFsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (m *MyFsm) Restore(snapshot io.ReadCloser) error {
	return nil
}

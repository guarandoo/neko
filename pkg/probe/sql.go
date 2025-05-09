package probe

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/guarandoo/neko/pkg/core"
)

const SqlProbeType string = "sql"

var onceInitSqlProbe sync.Once

func initSqlProbe() {

}

type sqlProbe struct {
	driver string
	dsn    string
	query  string
}

func (p *sqlProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	tests := []core.Test{}
	test := core.Test{
		Target: "",
		Status: core.StatusUp,
		Error:  nil,
		Extras: make(map[string]any),
	}

	con, err := sql.Open(p.driver, p.dsn)
	if err != nil {
		return &core.Result{Tests: []core.Test{}}, nil
	}
	defer con.Close()

	res := con.QueryRowContext(ctx, p.query)
	var val *int
	if err := res.Scan(&val); err != nil {
		return nil, fmt.Errorf("unable to perform query: %w", err)
	}

	if val == nil || *val == 0 {
		test.Status = core.StatusDown
	}

	tests = append(tests, test)

	return &core.Result{Tests: tests}, nil
}

type SqlProbeOptions struct {
	ProbeOptions
	Driver string
	DSN    string
	Query  string
}

func NewSqlProbe(options SqlProbeOptions) (Probe, error) {
	onceInitSqlProbe.Do(initSqlProbe)

	return &sqlProbe{
		driver: options.Driver,
		dsn:    options.DSN,
		query:  options.Query,
	}, nil
}

package probe

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/guarandoo/neko/pkg/core"

	"github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5"
)

const SqlProbeType string = "sql"

var onceInitSqlProbe sync.Once

func getHostFromDsn(driver string, dsn string) (string, error) {
	switch driver {
	case "mysql":
		c, err := mysql.ParseDSN(dsn)
		if err != nil {
			return "", fmt.Errorf("unable to parse DSN: %w", err)
		}
		return c.Addr, nil

	default:
		if strings.Contains(dsn, "://") {
			u, err := url.Parse(dsn)
			if err == nil {
				return u.Host, nil
			}
		}

		return "", errors.New("unable to get host from DSN")
	}
}

func initSqlProbe() {

}

type sqlProbe struct {
	host   string
	driver string
	dsn    string
	query  string
}

func (p *sqlProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {

	tests := []core.Test{}
	test := core.Test{
		Target: p.host,
		Status: core.StatusUp,
		Error:  nil,
		Extras: make(map[string]any),
	}

	con, err := sql.Open(p.driver, p.dsn)
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
		return &core.Result{Tests: []core.Test{test}}, nil
	}
	defer func() { _ = con.Close() }()

	res := con.QueryRowContext(ctx, p.query)
	var val *int
	if err := res.Scan(&val); err != nil {
		test.Status = core.StatusDown
		test.Error = err
		return &core.Result{Tests: []core.Test{test}}, nil
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

	host, err := getHostFromDsn(options.Driver, options.DSN)
	if err != nil {
		return nil, err
	}

	return &sqlProbe{
		host:   host,
		driver: options.Driver,
		dsn:    options.DSN,
		query:  options.Query,
	}, nil
}

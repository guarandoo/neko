package probe

import (
	"database/sql"

	"github.com/guarandoo/neko/pkg/core"
)

type sqlProbe struct {
	db *sql.DB
}

type msSqlProbe struct {
	sqlProbe
}

func NewMsSqlProbe() (Probe, error) {
	return &msSqlProbe{}, nil
}

type postgresProbe struct {
	sqlProbe
}

func NewPostgresProbe() (Probe, error) {
	return &postgresProbe{}, nil
}

func (p *sqlProbe) Probe() (*core.Result, error) {
	con, err := sql.Open("mysql", "")
	if err != nil {
		return &core.Result{Tests: []core.Test{}}, nil
	}

	defer con.Close()

	return &core.Result{Tests: []core.Test{}}, nil
}

package probe

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/guarandoo/neko/pkg/core"
)

func TestSqlProbe(t *testing.T) {
	driver := "sqlmock"
	dsn := "test"

	db, mock, err := sqlmock.NewWithDSN(dsn)
	if err != nil {
		t.Fatalf("unable to create sql mock: %v", err)
	}
	defer db.Close()

	rowsMock := sqlmock.NewRows([]string{"value"})
	rowsMock.AddRow(1)

	query := "SELECT 1"
	mock.ExpectQuery(query).WillReturnRows(rowsMock)

	probe, err := NewSqlProbe(SqlProbeOptions{
		ProbeOptions: ProbeOptions{},
		Driver:       driver,
		DSN:          dsn,
		Query:        query,
	})
	if err != nil {
		t.Fatalf("unable to create sql probe: %v", err)
	}

	ctx, cancel := getContextWithTimeout(context.Background(), time.Second*30)
	defer cancel()

	res, err := probe.Probe(ctx, "", "")
	if err != nil {
		t.Fatalf("unable to probe: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Fatalf("probe returned unexpected test count: %v", len(res.Tests))
	}

	test := res.Tests[0]
	if test.Status != core.StatusUp {
		t.Fatalf("probe returned unexpected status, expecting: %v found: %v", core.StatusUp, test.Status)
	}
}

func TestSqlProbeFail(t *testing.T) {
	driver := "sqlmock"
	dsn := "test"

	db, mock, err := sqlmock.NewWithDSN(dsn)
	if err != nil {
		t.Fatalf("unable to create sql mock: %v", err)
	}
	defer db.Close()

	rowsMock := sqlmock.NewRows([]string{"value"})
	rowsMock.AddRow(0)

	query := "SELECT 0"
	mock.ExpectQuery(query).WillReturnRows(rowsMock)

	probe, err := NewSqlProbe(SqlProbeOptions{
		ProbeOptions: ProbeOptions{},
		Driver:       driver,
		DSN:          dsn,
		Query:        query,
	})
	if err != nil {
		t.Fatalf("unable to create sql probe: %v", err)
	}

	ctx, cancel := getContextWithTimeout(context.Background(), time.Second*30)
	defer cancel()

	res, err := probe.Probe(ctx, "", "")
	if err != nil {
		t.Fatalf("unable to probe: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Fatalf("probe returned unexpected test count: %v", len(res.Tests))
	}

	test := res.Tests[0]
	if test.Status != core.StatusDown {
		t.Fatalf("probe returned unexpected status, expecting: %v found: %v", core.StatusDown, test.Status)
	}
}

package store

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	// enable the pq driver
	_ "github.com/lib/pq"
	// enable the sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

// SQL abstracts access to the database.
type SQL struct {
	DB     *sqlx.DB
	logger logrus.FieldLogger
}

// Config contains database configuration
type Config struct {
	Scheme          string
	URL             string
	IdleConns       int           `mapstructure:"idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

var (
	// errDBSchemeRequired the type of the database
	errDBSchemeRequired = errors.New("failed database scheme is required")
)

func New(cfg Config, logger logrus.FieldLogger) (*SQL, error) {
	if cfg.Scheme == "" {
		return nil, errDBSchemeRequired
	}

	url, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dsn as an url")
	}

	var db *sqlx.DB
	switch strings.ToLower(cfg.Scheme) {
	case "sqlite", "sqlite3":
		db, err = sqlx.Connect("sqlite3", fmt.Sprintf("%s?%s", url.Host, url.RawQuery))
		if err != nil {
			return nil, errors.Wrap(err, "failed to connect to sqlite database")
		}
		// Override the default mapper to use the field names "as-is"
		db.MapperFunc(func(s string) string { return s })

	case "postgres", "postgresql":
		db, err = sqlx.Connect("postgres", url.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to connect to postgres database")
		}
	default:
		return nil, errors.Errorf("unsupported dsn scheme %s", cfg.Scheme)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.IdleConns)
	db.SetConnMaxLifetime(cfg.MaxConnLifetime)

	return &SQL{
		DB:     db,
		logger: logger,
	}, nil

}

// tableExists checks if a table exists in the database
func (sqlStore *SQL) tableExists(tableName string) (bool, error) {
	var tableExists bool

	switch sqlStore.DB.DriverName() {
	case "sqlite3":
		err := sqlStore.Get(sqlStore.DB, &tableExists,
			"SELECT COUNT(*) == 1 FROM sqlite_master WHERE type='table' AND name='System'",
		)
		if err != nil {
			return false, errors.Wrapf(err, "failed to check if %s table exists", tableName)
		}

	case "postgres":
		err := sqlStore.Get(sqlStore.DB, &tableExists,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 'system')",
		)
		if err != nil {
			return false, errors.Wrapf(err, "failed to check if %s table exists", tableName)
		}

	default:
		return false, errors.Errorf("unsupported driver %s", sqlStore.DB.DriverName())
	}

	return tableExists, nil
}

// queryer is an interface describing a resource that can query.
//
// It exactly matches sqlx.Queryer, existing simply to constrain sqlx usage to this file.
type queryer interface {
	sqlx.Queryer
}

// get queries for a single row, writing the result into dest.
//
// Use this to simplify querying for a single row or column. Dest may be a pointer to a simple
// type, or a struct with fields to be populated from the returned columns.
func (sqlStore *SQL) Get(q sqlx.Queryer, dest interface{}, query string, args ...interface{}) error {
	query = sqlStore.DB.Rebind(query)

	return sqlx.Get(q, dest, query, args...)
}

// builder is an interface describing a resource that can construct SQL and arguments.
//
// It exists to allow consuming any squirrel.*Builder type.
type builder interface {
	ToSql() (string, []interface{}, error)
}

// get queries for a single row, building the sql, and writing the result into dest.
//
// Use this to simplify querying for a single row or column. Dest may be a pointer to a simple
// type, or a struct with fields to be populated from the returned columns.
func (sqlStore *SQL) GetBuilder(q sqlx.Queryer, dest interface{}, b builder) error {
	sql, args, err := b.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build sql")
	}

	sql = sqlStore.DB.Rebind(sql)

	err = sqlx.Get(q, dest, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// selectBuilder queries for one or more rows, building the sql, and writing the result into dest.
//
// Use this to simplify querying for multiple rows (and possibly columns). Dest may be a slice of
// a simple, or a slice of a struct with fields to be populated from the returned columns.
func (sqlStore *SQL) SelectBuilder(q sqlx.Queryer, dest interface{}, b builder) error {
	sql, args, err := b.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build sql")
	}

	sql = sqlStore.DB.Rebind(sql)

	err = sqlx.Select(q, dest, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

// execer is an interface describing a resource that can execute write queries.
//
// It allows the use of *sqlx.Db and *sqlx.Tx.
type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	DriverName() string
}

// exec executes the given query using positional arguments, automatically rebinding for the db.
func (sqlStore *SQL) Exec(e execer, sql string, args ...interface{}) (sql.Result, error) {
	sql = sqlStore.DB.Rebind(sql)
	return e.Exec(sql, args...)
}

// exec executes the given query, building the necessary sql.
func (sqlStore *SQL) ExecBuilder(e execer, b builder) (sql.Result, error) {
	sql, args, err := b.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build sql")
	}

	return sqlStore.Exec(e, sql, args...)
}

package sql

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	_ "github.com/lib/pq"
)

var (
	queryInsert = "INSERT INTO %s (%s) VALUES %s"
	queryUpdate = "UPDATE %s SET %s WHERE %s"
)

type DB struct {
	*sqlx.DB
}

//NewDB won't attmept connection to DB unless needed. Callers can call Ping() to test connection.
func NewDB(driver, url string) (*DB, error) {
	db, err := sqlx.Open(driver, url)
	if err != nil {
		return nil, err
	}

	db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return &DB{db}, err
}

type Tx struct {
	*sqlx.Tx
}

func (d *DB) Begin() *Tx {
	tx := d.MustBegin()
	return &Tx{tx}
}

func (d *DB) BatchInsertStatement(tableName string, records []interface{}, fi *FieldInfo) (string, []interface{}) {
	var params []interface{}
	var stmt string

	for _, p := range records {
		iter, err := newStructIterator(p, fi)
		if err != nil {
			return "", nil
		}

		for {
			sf := iter.next()
			if sf == nil {
				break
			}
			params = append(params, sf.value)
		}
	}

	{
		if d.DB.DriverName() == "postgres" {
			stmt = fmt.Sprintf(queryInsert, tableName, fi.DBTags, fi.QuestionBindVar(len(records)))
			stmt = d.DB.Rebind(stmt) // needed for QBindVars as Postgres support EnumBindVars
		} else {
			stmt = fmt.Sprintf(queryInsert, tableName, fi.DBTags, fi.DollarBindVar(len(records)))
		}
	}

	return stmt, params
}

// PartialUpdateStmt create
func (d *DB) PartialUpdateStmt(b interface{}, tableName string, condition string, fi *FieldInfo) (string, error) {
	iter, err := newStructIterator(b, fi)
	if err != nil {
		return "", err
	}

	var stmts []string
	for {
		sf := iter.next()
		if sf == nil {
			break
		}

		if isZero(sf.value) {
			continue
		}
		stmts = append(stmts, fmt.Sprintf("%s = '%v'", sf.dbtag, adaptValue(sf.value)))
	}

	return fmt.Sprintf(queryUpdate, tableName, strings.Join(stmts, ","), condition), nil
}

func adaptValue(v interface{}) interface{} {
	switch reflect.TypeOf(v).Name() {
	case "Time":
		return v.(time.Time).Round(time.Millisecond).Format(time.RFC3339)
	default:
		return v
	}
}

func (d *DB) BatchInsert(tableName string, records []interface{}, fi *FieldInfo) error {
	stmt, params := d.BatchInsertStatement(tableName, records, fi)
	return d.ExecTxn(stmt, params...)
}

func (d *DB) In(query string, args ...interface{}) (string, []interface{}, error) {
	return sqlx.In(query, args...)
}

func (d *DB) ExecTxn(stmt string, args ...interface{}) error {
	tx := d.Begin()
	var err error

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()
	_, err = tx.Exec(stmt, args...)
	return err
}

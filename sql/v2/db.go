package v2

import (
	"errors"
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
	queryUpsert = "%s ON CONFLICT (%s) DO UPDATE SET %s"
	driverName  = "postgres"
)

// DB wrapper struct.
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

// Begin a transaction.
func (d *DB) Begin() *Tx {
	return &Tx{d.DB.MustBegin()}
}

// Transaction executes a method in a transaction.
func (d *DB) Transaction(fn func(tx *Tx) (interface{}, error)) error {
	tx := d.Begin()
	var err error

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()
	_, err = fn(tx)
	return err
}

// In query wrapper.
func In(query string, args ...interface{}) (string, []interface{}, error) {
	return sqlx.In(query, args...)
}

// Rebind generates statement for rebinding query.
func Rebind(driverName string, stmt string) string {
	return sqlx.Rebind(sqlx.BindType(driverName), stmt)
}

// BatchInsertStatement prepares a insert statement.
func BatchInsertStatement(table string, records []interface{}, fi *FieldInfo) (string, []interface{}) {
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
		if fi.DriverName == "postgres" {
			stmt = fmt.Sprintf(queryInsert, table, fi.DBTags, fi.QuestionBindVar(len(records)))
			stmt = Rebind(fi.DriverName, stmt) // needed for QBindVars as Postgres support EnumBindVars
		} else {
			stmt = fmt.Sprintf(queryInsert, table, fi.DBTags, fi.DollarBindVar(len(records)))
		}
	}

	return stmt, params
}

// BatchUpsertStmt prepares upsert statement.
func BatchUpsertStmt(table string, records []interface{}, conflictKey string, fi *FieldInfo) (string, []interface{}, error) {
	if len(records) == 0 {
		return "", nil, errors.New("no reords to upsert")
	}

	insertStmt, params := BatchInsertStatement(table, records, fi)

	iter, err := newStructIterator(records[0], fi)
	if err != nil {
		return "", nil, err
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
		stmts = append(stmts, fmt.Sprintf("%s = excluded.%s", sf.dbtag, sf.dbtag))
	}

	return fmt.Sprintf(queryUpsert, insertStmt, conflictKey, strings.Join(stmts, ",")), params, nil
}

// PartialUpdateStmt create
func PartialUpdateStmt(b interface{}, table string, condition string, fi *FieldInfo) (string, error) {
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

	return fmt.Sprintf(queryUpdate, table, strings.Join(stmts, ","), condition), nil
}

func adaptValue(v interface{}) interface{} {
	switch reflect.TypeOf(v).Name() {
	case "Time":
		return v.(time.Time).Round(time.Millisecond).Format(time.RFC3339)
	default:
		return v
	}
}

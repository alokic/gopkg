package sql

import (
	"strings"

	"github.com/alokic/gopkg/sql"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/lib/pq"

	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	sqlxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/jmoiron/sqlx"
)

// NewDB return sql.DB
func NewDB(driver, url string, serviceName string) (*sql.DB, error) {
	// Register informs the sqlxtrace package of the driver that we will be using in our program.
	// It uses a default service name, in the below case "postgres.db". To use a custom service
	// name use RegisterWithServiceName.
	sqltrace.Register(driver, &pq.Driver{}, sqltrace.WithServiceName(serviceName))
	db, err := sqlxtrace.Open(driver, url)
	if err != nil {
		return nil, err
	}

	db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return &sql.DB{DB: db}, nil
}

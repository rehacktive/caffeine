package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	pg_dbName             = "caffeine"
	pg_insertQuery        = "INSERT INTO %v (id, data) VALUES($1, $2) ON CONFLICT (id) DO UPDATE SET data = $2"
	pg_tablesQuery        = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
	pg_getQuery           = "SELECT data FROM %v WHERE id = $1"
	pg_getAllQuery        = "SELECT id, data FROM %v ORDER BY id"
	pg_deleteQuery        = "DELETE FROM %v WHERE id = $1"
	pg_dropNamespaceQuery = "DROP TABLE %v"
)

type PGDatabase struct {
	Host string
	User string
	Pass string

	db *sql.DB
}

func (p *PGDatabase) Init() {
	connInfo := fmt.Sprintf("user=%v password=%v host=%v sslmode=disable", p.User, p.Pass, p.Host)
	db, err := sql.Open("postgres", connInfo)

	if err != nil {
		log.Fatalf("error connecting to postgres: %v", err)
	}

	_, err = db.Exec("create database " + pg_dbName)
	if err != nil {
		log.Println(err)
	}
	p.db = db
}

func (p PGDatabase) Upsert(namespace string, key string, value []byte) *DbError {
	err := p.ensureNamespace(namespace)

	if err != nil {
		return &DbError{
			ErrorCode: NAMESPACE_NOT_FOUND,
			Message:   fmt.Sprintf("namespace %v does not exist", namespace),
		}
	}
	_, dbErr := p.db.Exec(fmt.Sprintf(pg_insertQuery, namespace), key, string(value))
	if dbErr != nil {
		return &DbError{
			ErrorCode: INTERNAL_ERROR,
			Message:   fmt.Sprintf("error on Upsert: %v", dbErr),
		}
	}
	return nil
}

func (p PGDatabase) Get(namespace string, key string) ([]byte, *DbError) {
	rows, dbErr := p.db.Query(fmt.Sprintf(pg_getQuery, namespace), key)
	if dbErr != nil {
		return nil, &DbError{
			ErrorCode: INTERNAL_ERROR,
			Message:   fmt.Sprintf("error on Get: %v", dbErr),
		}
	}
	defer rows.Close()
	if rows.Next() {
		var data string
		scanErr := rows.Scan(&data)
		if scanErr != nil {
			return nil, &DbError{
				ErrorCode: INTERNAL_ERROR,
				Message:   fmt.Sprintf("scan %v", scanErr),
			}
		}
		return []byte(data), nil
	}
	return nil, &DbError{
		ErrorCode: ID_NOT_FOUND,
		Message:   fmt.Sprintf("value not found in namespace %v for key %v", namespace, key),
	}
}

func (p PGDatabase) GetAll(namespace string) (map[string][]byte, *DbError) {
	sqlStatement := fmt.Sprintf(pg_getAllQuery, namespace)
	rows, dbErr := p.db.Query(sqlStatement)
	if dbErr != nil {
		return nil, &DbError{
			ErrorCode: INTERNAL_ERROR,
			Message:   fmt.Sprintf("error on Get: %v", dbErr),
		}
	}
	defer rows.Close()

	ret := make(map[string][]byte)

	for rows.Next() {
		var id, data string
		scanErr := rows.Scan(&id, &data)
		if scanErr != nil {
			return nil, &DbError{
				ErrorCode: INTERNAL_ERROR,
				Message:   fmt.Sprintf("scan %v", scanErr),
			}
		}
		ret[id] = []byte(data)
	}
	return ret, nil
}

func (p PGDatabase) Delete(namespace string, key string) *DbError {
	_, err := p.db.Exec(fmt.Sprintf(pg_deleteQuery, namespace), key)
	if err != nil {
		message := fmt.Sprintf("error on Delete: %v", err)
		return &DbError{
			ErrorCode: INTERNAL_ERROR,
			Message:   message,
		}
	}
	return nil
}

func (p PGDatabase) DeleteAll(namespace string) *DbError {
	sqlStatement := fmt.Sprintf(pg_dropNamespaceQuery, namespace)
	_, err := p.db.Exec(sqlStatement)
	if err != nil {
		message := fmt.Sprintf("error on DeleteAll: %v", err)
		return &DbError{
			ErrorCode: INTERNAL_ERROR,
			Message:   message,
		}
	}
	return nil
}

func (p PGDatabase) GetNamespaces() []string {
	rows, err := p.db.Query(pg_tablesQuery)
	if err != nil {
		log.Printf("error on GetNamespaces: %v\n", err)
	}
	defer rows.Close()

	ret := make([]string, 0)
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			log.Printf("error on Scan: %v\n", err)
		}
		ret = append(ret, tableName)
	}
	return ret
}

func (p PGDatabase) ensureNamespace(namespace string) (err error) {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v ( id text PRIMARY KEY, data json NOT NULL)", namespace)
	_, err = p.db.Exec(query)

	if err != nil {
		log.Printf("error creating table: %v\n", err)
	}

	return err
}

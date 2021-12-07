package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqlite_dbName             = "caffeine"
	sqlite_insertQuery        = "INSERT INTO %v (id, data) VALUES($1, $2) ON CONFLICT (id) DO UPDATE SET data = $2"
	sqlite_tablesQuery        = "SELECT  `name` FROM sqlite_master WHERE `type`='table'  ORDER BY name"
	sqlite_getQuery           = "SELECT data FROM %v WHERE id = $1"
	sqlite_getAllQuery        = "SELECT id, data FROM %v ORDER BY id"
	sqlite_deleteQuery        = "DELETE FROM %v WHERE id = $1"
	sqlite_dropNamespaceQuery = "DROP TABLE %v"
)

type SQLiteDatabase struct {
	DirPath string

	db *sql.DB
}

func (p *SQLiteDatabase) Init() {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%v/%v", p.DirPath, sqlite_dbName))
	if err != nil {
		log.Fatal("error connecting to postgres: ", err)
	}
	p.db = db
}

func (p SQLiteDatabase) Upsert(namespace string, key string, value []byte) *DbError {
	err := p.ensureNamespace(namespace)

	if err != nil {
		return &DbError{
			ErrorCode: NAMESPACE_NOT_FOUND,
			Message:   fmt.Sprintf("namespace %v does not exist", namespace),
		}
	}
	_, dbErr := p.db.Exec(fmt.Sprintf(sqlite_insertQuery, namespace), key, string(value))
	if dbErr != nil {
		return &DbError{
			ErrorCode: INTERNAL_ERROR,
			Message:   fmt.Sprintf("error on Upsert: %v", dbErr),
		}
	}
	return nil
}

func (p SQLiteDatabase) Get(namespace string, key string) ([]byte, *DbError) {
	rows, dbErr := p.db.Query(fmt.Sprintf(sqlite_getQuery, namespace), key)
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

func (p SQLiteDatabase) GetAll(namespace string) (map[string][]byte, *DbError) {
	sqlStatement := fmt.Sprintf(sqlite_getAllQuery, namespace)
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

func (p SQLiteDatabase) Delete(namespace string, key string) *DbError {
	_, err := p.db.Exec(fmt.Sprintf(sqlite_deleteQuery, namespace), key)
	if err != nil {
		message := fmt.Sprintf("error on Delete: %v", err)
		return &DbError{
			ErrorCode: INTERNAL_ERROR,
			Message:   message,
		}
	}
	return nil
}

func (p SQLiteDatabase) DeleteAll(namespace string) *DbError {
	sqlStatement := fmt.Sprintf(sqlite_dropNamespaceQuery, namespace)
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

func (p SQLiteDatabase) GetNamespaces() []string {
	ret := []string{}
	rows, err := p.db.Query(sqlite_tablesQuery)
	if err != nil {
		log.Println("error on GetNamespaces: ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			log.Println("Scan: ", err)
		}
		ret = append(ret, tableName)
	}
	return ret
}

func (p SQLiteDatabase) ensureNamespace(namespace string) (err error) {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v ( id string PRIMARY KEY, data string NOT NULL)", namespace)
	_, err = p.db.Exec(query)

	if err != nil {
		log.Println("error creating table: ", err)
	}

	return err
}

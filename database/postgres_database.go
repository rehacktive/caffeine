package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	dbName             = "caffeine"
	dbPort             = 5432
	insertQuery        = "INSERT INTO %v (id, data) VALUES($1, $2) ON CONFLICT (id) DO UPDATE SET data = $2"
	tablesQuery        = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
	getQuery           = "SELECT data FROM %v WHERE id = $1"
	getAllQuery        = "SELECT id, data FROM %v"
	deleteQuery        = "DELETE FROM %v WHERE id = $1"
	dropNamespaceQuery = "DROP TABLE %v"
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
		log.Fatal("error connecting to postgres: ", err)
	}

	_, err = db.Exec("create database " + dbName)
	if err != nil {
		log.Println(err)
	}
	p.db = db
}

func (p PGDatabase) Upsert(namespace string, key string, value []byte) (err error) {
	err = p.ensureNamespace(namespace)
	if err != nil {
		return
	}
	_, err = p.db.Exec(fmt.Sprintf(insertQuery, namespace), key, string(value))
	if err != nil {
		log.Println("error on Upsert: ", err)
	}
	return
}

func (p PGDatabase) Get(namespace string, key string) ([]byte, error) {
	rows, err := p.db.Query(fmt.Sprintf(getQuery, namespace), key)
	if err != nil {
		log.Println("error on Get: ", err)
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var data string
		err = rows.Scan(&data)
		if err != nil {
			log.Println("Scan: ", err)
		}
		return []byte(data), nil
	}
	return nil, fmt.Errorf("value not found in namespace %v for key %v", namespace, key)
}

func (p PGDatabase) GetAll(namespace string) (map[string][]byte, error) {
	sqlStatement := fmt.Sprintf(getAllQuery, namespace)
	rows, err := p.db.Query(sqlStatement)
	if err != nil {
		log.Println("error on GetAll: ", err)
		return nil, err
	}
	defer rows.Close()

	ret := make(map[string][]byte)

	for rows.Next() {
		var id, data string
		err = rows.Scan(&id, &data)
		if err != nil {
			log.Println("Scan: ", err)
		}
		ret[id] = []byte(data)
	}
	return ret, nil
}

func (p PGDatabase) Delete(namespace string, key string) error {
	_, err := p.db.Exec(fmt.Sprintf(deleteQuery, namespace), key)
	if err != nil {
		log.Println("error on Delete: ", err)
		return err
	}
	return nil
}

func (p PGDatabase) DeleteAll(namespace string) error {
	sqlStatement := fmt.Sprintf(dropNamespaceQuery, namespace)
	_, err := p.db.Exec(sqlStatement)
	if err != nil {
		log.Println("error on DeleteAll: ", err)
	}
	return nil
}

func (p PGDatabase) GetNamespaces() []string {
	ret := []string{}
	rows, err := p.db.Query(tablesQuery)
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

func (p PGDatabase) ensureNamespace(namespace string) (err error) {
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v ( id text PRIMARY KEY, data json NOT NULL)", namespace)
	_, err = p.db.Exec(query)

	if err != nil {
		log.Println("error creating table: ", err)
	}

	return err
}

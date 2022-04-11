package data

import (
	_ "database/sql" // Database.
	"fmt"
	config "github.com/harisbeha/media-transcoder/internal/config"
	"log"
	_ "github.com/lib/pq" // Postgres driver.
	// _ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
)

var (
	connectionString = ""
	conn             *sqlx.DB
)

// ConnectDB Connects to postgres database
func ConnectDB() (*sqlx.DB, error) {
	var err error
	if connectionString == "" {
		fmt.Println("connection not set. setting now.")
		var (
			host     = config.Get().DatabaseHost
			port     = config.Get().DatabasePort
			user     = config.Get().DatabaseUser
			password = config.Get().DatabasePassword
			dbname   = config.Get().DatabaseName
		)
		connectionString = fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}

	if conn, err = sqlx.Connect("postgres", connectionString); err != nil {
		log.Panic(err)
	}
	return conn, err
}
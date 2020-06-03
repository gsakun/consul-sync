package db

import (
	"database/sql"
	// Register some standard stuff
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

//DB Initializes global variables
var DB *sql.DB

// Init init db connection
func Init(dbaddress string, maxconn, maxidle int) {
	var err error
	DB, err = sql.Open("mysql", dbaddress)
	if err != nil {
		log.Fatalf("open db fail:", err)
	}

	DB.SetMaxIdleConns(maxidle)
	DB.SetMaxOpenConns(maxconn)
	err = DB.Ping()
	if err != nil {
		log.Fatalf("ping db fail:", err)
	}
}

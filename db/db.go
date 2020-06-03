package db

import (
	"database/sql"

	log "github.com/sirupsen/logrus"
)

// Init init db connection
func Init(dbaddress string, maxconn, maxidle int) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", dbaddress)
	if err != nil {
		log.Errorf("open db fail:", err)
		return nil, err
	}

	db.SetMaxIdleConns(maxidle)
	db.SetMaxOpenConns(maxconn)
	err = db.Ping()
	if err != nil {
		log.Errorf("ping db fail:", err)
		return nil, err
	}
	return db, err
}

package main

import (
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	loglevel := os.Getenv("LOG_LEVEL")
	var logLevel log.Level
	log.Infof("loglevel env is %s", loglevel)
	if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
		logLevel = log.DebugLevel
		log.Infof("log level is %s", loglevel)
		log.SetReportCaller(true)
	} else {
		log.SetLevel(log.InfoLevel)
		logLevel = log.InfoLevel
		log.Infoln("log level is normal")
	}
	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   "logs/alarmtransfer.log",
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Level:      logLevel,
		Formatter: &log.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		},
	})
	log.SetOutput(colorable.NewColorableStdout())
	if err != nil {
		log.Fatalf("Failed to initialize file rotate hook: %v", err)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})
	log.AddHook(rotateFileHook)
}

func main() {
	var (
		dbaddress = kingpin.Flag(
			"database",
			"The database address for get machine info.",
		).Default("").String()
		consuladdress = kingpin.Flag(
			"consul",
			"The consul address for registry machine info.",
		).Default("").String()
		maxconn = kingpin.Flag(
			"maxconn",
			"Database maxconn.",
		).Default("100").Int()
		maxidle = kingpin.Flag(
			"maxidle",
			"Database maxidle.",
		).Default(string(*maxconn)).Int()
		interval = kingpin.Flag("interval", "Sync Interval.").Default("60").Int()
	)

	kingpin.Version("consul-sync v1.0")
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	go func() {
		var db *sql.DB
		db, err := DBInit(*dbaddress, *maxidle, *maxconn)
		if err != nil {
			log.Errorf("Init db connection failed %v", err)
			time.Sleep(60 * time.Second)
		} else {

			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		os.Exit(0)
	}()

	select {}

}

func DBInit(database string, maxidle, maxconn int) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", database)
	if err != nil {
		log.Fatalln("open db fail:", err)
	}

	db.SetMaxIdleConns(maxidle)
	db.SetMaxOpenConns(maxconn)

	err = db.Ping()
	if err != nil {
		log.Fatalf("ping db fail:", err)
		return err
	}
	return nil
}

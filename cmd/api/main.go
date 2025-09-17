package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq" // alias of this import is blank intentionally to stop go compiler from complaining
	"github.com/solomonsitotaw23/greenlight/internal/data"
)

const version = "1.0.0"

// configuration setting struct
type config struct {
	port int
	env  string //development, staging, production
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleCons  int
		maxIdleTime  time.Duration
	}

	limiter struct {
		rps    float64 //request per second
		burst  int     //bucket size
		enable bool    //enable disable rate limiter
	}
}

// dependencies for http handlers
type application struct {
	config config
	logger *slog.Logger
	models data.Models
}

func main() {
	var cfg config

	//read value of port and env from command line
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development | staging | production)")

	// read the DSN value from db-dsn command-line flag
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	// Read connection pool settings from command-line flags
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Postgresql max open connections")
	flag.IntVar(&cfg.db.maxIdleCons, "db-max-idle-conns", 25, "Postgresql max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgresql max connection idle time")

	// read config for the rate limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enable, "limiter-enable", true, "Enable rate limiter")

	flag.Parse()
	// initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// call openDB() helper function to create a connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// this function will return a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	// set maximum of inuse + idle connections
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	//  set maximum number of idle connections
	db.SetMaxIdleConns(cfg.db.maxIdleCons)
	// set maximum idle timeout for connection in the pool
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

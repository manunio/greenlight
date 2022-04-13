package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Import the pq driver so that it can register itself with the database/sql
	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used.
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "development|staging|production")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgresql://localhost:5431/greenlight", "PostgreSQL DSN")

	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Call openDB() helper function to create the connection pool,
	// passing in the config struct. If this return an error, we log it and exit the
	// application immediately.
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Defer a call db.Close() so that the connection pool is closed before the
	// main() function exists.
	defer func() {
		err = db.Close()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	logger.Println("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
	}

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	logger.Printf("starting %s server on %d", cfg.env, cfg.port)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// openDB returns a sql.DB connection pool.
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config
	// struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PingContext() establishes a new connection to the database, passing in the
	// context we create above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return an
	// error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}

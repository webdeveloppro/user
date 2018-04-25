package main

import (
	"log"
	"os"

	"github.com/jackc/pgx"
)

func main() {

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USERNAME")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		log.Print("Empty host string, setup DB_HOST env")
		host = "localhost"
	}

	if user == "" {
		log.Fatal("Empty user string, setup DB_USER env")
		return
	}

	if dbname == "" {
		log.Fatal("Empty dbname string, setup DB_DBNAME env")
		return
	}

	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     host,
			User:     user,
			Password: dbpassword,
			Database: dbname,
		},
		MaxConnections: 100,
	}

	pg, err := pgx.NewConnPool(connPoolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool %v", err)
	}
	defer pg.Close()

	storage := NewPostgres(pg)

	app, _ := NewApp(storage)
	app.Run(os.Getenv("PORT"))
}

package main

import (
	"log"
	"os"

	"github.com/jackc/pgx"
	u "github.com/webdeveloppro/user/pkg/user"
)

func main() {

	dbhost := os.Getenv("DB_HOST")
	dbuser := os.Getenv("DB_USERNAME")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if dbhost == "" {
		log.Print("Empty host string, setup DB_HOST env")
		dbhost = "localhost"
	}

	if dbuser == "" {
		log.Fatal("Empty user string, setup DB_USER env")
		return
	}

	if dbname == "" {
		log.Fatal("Empty dbname string, setup DB_DBNAME env")
		return
	}

	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     dbhost,
			User:     dbuser,
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

	storage := u.NewPostgres(pg)

	app, _ := u.NewApp(storage)
	app.Run(os.Getenv("HOST") + ":" + os.Getenv("PORT"))
}

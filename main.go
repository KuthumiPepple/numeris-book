package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kuthumipepple/numeris-book/api"
	"github.com/kuthumipepple/numeris-book/db"
)

const (
	dbSource          = "postgresql://root:secret@localhost:5432/bank?sslmode=disable"
	httpServerAddress = "0.0.0.0:8080"
)

func main() {
	connPool, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("cannot connect to db")
	}

	store := db.NewStore(connPool)
	apiServer := api.NewServer(store)
	err = apiServer.Start(httpServerAddress)
	if err != nil {
		log.Fatal("cannot start server")
	}
}

package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/luan-k/fiber-cms/api"
	db "github.com/luan-k/fiber-cms/db/sqlc"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5432/fiber_cms?sslmode=disable"
	apiPort  = ":8080"
)

func main() {

	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	server := api.NewServer(db.NewStore(conn))

	log.Println("Starting Fiber CMS API on port 8080...")
	err = server.Start(apiPort)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

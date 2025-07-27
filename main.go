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

	log.Println("🚀 Starting Fiber CMS...")

	log.Println("📊 Connecting to database...")
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("❌ Cannot connect to db:", err)
	}
	log.Println("✅ Database connected successfully")

	log.Println("🔧 Setting up server...")
	server := api.NewServer(db.NewStore(conn))

	log.Println("🌐 Starting Fiber CMS API on port 8080...")
	err = server.Start(apiPort)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

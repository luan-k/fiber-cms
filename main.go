package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/luan-k/fiber-cms/api"
	db "github.com/luan-k/fiber-cms/db/sqlc"
	"github.com/luan-k/fiber-cms/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("❌ Cannot load config:", err)
	}

	log.Println("🚀 Starting Fiber CMS...")

	log.Println("📊 Connecting to database...")
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("❌ Cannot connect to db:", err)
	}
	log.Println("✅ Database connected successfully")

	log.Println("🔧 Setting up server...")
	server := api.NewServer(db.NewStore(conn))

	log.Println("🌐 Starting Fiber CMS API on port", config.APIPort)
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

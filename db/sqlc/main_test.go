package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/go-live-cms/go-live-cms/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testStore Store

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(conn)
	testStore = NewStore(conn)

	os.Exit(m.Run())
}

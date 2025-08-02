package api

import (
	"testing"
	"time"

	db "github.com/go-live-cms/go-live-cms/db/sqlc"
	"github.com/go-live-cms/go-live-cms/util"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:    "12345678901234567890123456789012",
		AccessTokenDuration:  time.Minute * 15,
		RefreshTokenDuration: time.Hour * 24,
	}

	server, err := NewServer(config, store)
	if err != nil {
		t.Fatal("Failed to create test server:", err)
	}

	return server
}

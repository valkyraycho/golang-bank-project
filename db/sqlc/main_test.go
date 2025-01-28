package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkyraycho/bank_project/utils"
)

var testStore Store

func TestMain(m *testing.M) {
	cfg, err := utils.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot log environment variable: ", err)
	}

	connPool, err := pgxpool.New(context.Background(), cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}

	testStore = NewStore(connPool)
	os.Exit(m.Run())
}

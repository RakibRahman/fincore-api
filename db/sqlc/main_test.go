package sqlc

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const testDBSource = "postgresql://root:secret@localhost:5433/fincore_db?sslmode=disable"

var testQueries *Queries

func TestMain(m *testing.M) {
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, testDBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	log.Printf("database connection established successfully to port: %d", 5433)

	testQueries = New(conn)
	os.Exit(m.Run())
}

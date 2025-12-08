package sqlc

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testDBSource = "postgresql://root:secret@localhost:5433/fincore_db?sslmode=disable"

var testQueries *Queries
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, testDBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	log.Printf("database connection established successfully to port: %d", 5433)

	testDB = conn
	testQueries = New(conn)
	os.Exit(m.Run())
}

// Helper function to create a test transaction that will be rolled back
func createTestTx(t *testing.T) (pgx.Tx, *Queries) {
	ctx := context.Background()
	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatal("failed to begin transaction:", err)
	}

	// Setup cleanup to rollback transaction
	t.Cleanup(func() {
		tx.Rollback(ctx)
	})

	// Return transaction-aware queries
	return tx, testQueries.WithTx(tx)
}

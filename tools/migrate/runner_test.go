package migrate

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/hylarucoder/rocketbase/tools/list"
	"github.com/hylarucoder/rocketbase/tools/test_utils"
	_ "github.com/lib/pq"
	"github.com/pocketbase/dbx"
)

func TestNewRunner(t *testing.T) {
	testDB, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	l := MigrationsList{}
	l.Register(nil, nil, "1_test.go")
	l.Register(nil, nil, "2_test.go")
	l.Register(nil, nil, "3_test.go")

	r, err := NewRunner(testDB.DB, l)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.migrationsList.Items()) != len(l.Items()) {
		t.Fatalf("Expected the same migrations list to be assigned, got \n%#v", r.migrationsList)
	}

	expectedQueries := []string{
		"CREATE TABLE IF NOT EXISTS \"_migrations\" (file VARCHAR(255) PRIMARY KEY NOT NULL, applied TIMESTAMPTZ NOT NULL)",
	}
	if len(expectedQueries) != len(testDB.CalledQueries) {
		t.Fatalf("Expected %d queries, got %d: \n%v", len(expectedQueries), len(testDB.CalledQueries), testDB.CalledQueries)
	}
	for _, q := range expectedQueries {
		if !list.ExistInSlice(q, testDB.CalledQueries) {
			t.Fatalf("Query %s was not found in \n%v", q, testDB.CalledQueries)
		}
	}
}

func TestRunnerUpAndDown(t *testing.T) {
	testDB, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	callsOrder := []string{}

	l := MigrationsList{}
	l.Register(func(db dbx.Builder) error {
		callsOrder = append(callsOrder, "up2")
		return nil
	}, func(db dbx.Builder) error {
		callsOrder = append(callsOrder, "down2")
		return nil
	}, "2_test")
	l.Register(func(db dbx.Builder) error {
		callsOrder = append(callsOrder, "up3")
		return nil
	}, func(db dbx.Builder) error {
		callsOrder = append(callsOrder, "down3")
		return nil
	}, "3_test")
	l.Register(func(db dbx.Builder) error {
		callsOrder = append(callsOrder, "up1")
		return nil
	}, func(db dbx.Builder) error {
		callsOrder = append(callsOrder, "down1")
		return nil
	}, "1_test")

	r, err := NewRunner(testDB.DB, l)
	if err != nil {
		t.Fatal(err)
	}

	// simulate partially out-of-order run migration
	err = r.saveAppliedMigration(testDB, "2_test")
	if err != nil {
		t.Fatal(err)
		return
	}

	// ---------------------------------------------------------------
	// Up()
	// ---------------------------------------------------------------

	if _, err := r.Up(); err != nil {
		println("up", err.Error())
		t.Fatal(err)
	}

	expectedUpCallsOrder := `["up1","up3"]` // skip up2 since it was applied previously

	upCallsOrder, err := json.Marshal(callsOrder)
	if err != nil {
		t.Fatal(err)
	}

	if v := string(upCallsOrder); v != expectedUpCallsOrder {
		t.Fatalf("Expected Up() calls order %s, got %s", expectedUpCallsOrder, upCallsOrder)
	}

	// ---------------------------------------------------------------

	// reset callsOrder
	callsOrder = []string{}

	// simulate unrun migration
	r.migrationsList.Register(nil, func(db dbx.Builder) error {
		callsOrder = append(callsOrder, "down4")
		return nil
	}, "4_test")

	// ---------------------------------------------------------------

	// ---------------------------------------------------------------
	// Down()
	// ---------------------------------------------------------------

	if _, err := r.Down(2); err != nil {
		t.Fatal(err)
	}

	//  ["down3","down1"], got ["down3","down2"]
	expectedDownCallsOrder := `["down3","down1"]` // revert in the applied order

	downCallsOrder, err := json.Marshal(callsOrder)
	if err != nil {
		t.Fatal(err)
	}

	if v := string(downCallsOrder); v != expectedDownCallsOrder {
		t.Fatalf("Expected Down() calls order %s, got %s", expectedDownCallsOrder, downCallsOrder)
	}
}

func TestHistorySync(t *testing.T) {
	testDB, err := createTestDB()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	// mock migrations history
	l := MigrationsList{}
	l.Register(func(db dbx.Builder) error {
		return nil
	}, func(db dbx.Builder) error {
		return nil
	}, "1_test")
	l.Register(func(db dbx.Builder) error {
		return nil
	}, func(db dbx.Builder) error {
		return nil
	}, "2_test")
	l.Register(func(db dbx.Builder) error {
		return nil
	}, func(db dbx.Builder) error {
		return nil
	}, "3_test")

	r, err := NewRunner(testDB.DB, l)
	if err != nil {
		t.Fatalf("Failed to initialize the runner: %v", err)
	}

	if _, err := r.Up(); err != nil {
		t.Fatalf("Failed to apply the mock migrations: %v", err)
	}

	if !r.isMigrationApplied(testDB.DB, "2_test") {
		t.Fatalf("Expected 2_test migration to be applied")
	}

	// mock deleted migrations
	r.migrationsList.list = []*Migration{r.migrationsList.list[0], r.migrationsList.list[2]}

	if err := r.removeMissingAppliedMigrations(); err != nil {
		t.Fatalf("Failed to remove missing applied migrations: %v", err)
	}

	if r.isMigrationApplied(testDB.DB, "2_test") {
		t.Fatalf("Expected 2_test migration to NOT be applied")
	}
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

type testDB struct {
	*dbx.DB
	CalledQueries []string
}

// NB! Don't forget to call `db.Close()` at the end of the test.
func createTestDB() (*testDB, error) {
	test_utils.LoadTestEnv()
	dbDSN := os.Getenv("DATABASE")
	sqlDB, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return nil, err
	}

	db := testDB{DB: dbx.NewFromDB(sqlDB, "postgres")}
	db.QueryLogFunc = func(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
		db.CalledQueries = append(db.CalledQueries, sql)
	}
	db.ExecLogFunc = func(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
		db.CalledQueries = append(db.CalledQueries, sql)
	}

	return &db, nil
}

package pocketbase

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/hylarucoder/rocketbase/plugins/migratecmd"
	"github.com/hylarucoder/rocketbase/tools/test_utils"

	"github.com/spf13/cobra"
)

func TestNew(t *testing.T) {
	// copy os.Args
	originalArgs := make([]string, len(os.Args))
	copy(originalArgs, os.Args)
	defer func() {
		// restore os.Args
		os.Args = originalArgs
	}()

	// change os.Args
	os.Args = os.Args[:1]
	os.Args = append(
		os.Args,
		"--dir=test_dir",
		"--encryptionEnv=test_encryption_env",
		"--debug=true",
	)

	app := New()

	if app == nil {
		t.Fatal("Expected initialized PocketBase instance, got nil")
	}

	if app.RootCmd == nil {
		t.Fatal("Expected RootCmd to be initialized, got nil")
	}

	if app.appWrapper == nil {
		t.Fatal("Expected appWrapper to be initialized, got nil")
	}

	if app.DataDir() != "test_dir" {
		t.Fatalf("Expected app.DataDir() %q, got %q", "test_dir", app.DataDir())
	}

	if app.EncryptionEnv() != "test_encryption_env" {
		t.Fatalf("Expected app.EncryptionEnv() test_encryption_env, got %q", app.EncryptionEnv())
	}
}

func TestNewWithConfig(t *testing.T) {
	app := NewWithConfig(Config{
		DefaultDataDir:       "test_dir",
		DefaultEncryptionEnv: "test_encryption_env",
		HideStartBanner:      true,
	})

	if app == nil {
		t.Fatal("Expected initialized PocketBase instance, got nil")
	}

	if app.RootCmd == nil {
		t.Fatal("Expected RootCmd to be initialized, got nil")
	}

	if app.appWrapper == nil {
		t.Fatal("Expected appWrapper to be initialized, got nil")
	}

	if app.hideStartBanner != true {
		t.Fatal("Expected app.hideStartBanner to be true, got false")
	}

	if app.DataDir() != "test_dir" {
		t.Fatalf("Expected app.DataDir() %q, got %q", "test_dir", app.DataDir())
	}

	if app.EncryptionEnv() != "test_encryption_env" {
		t.Fatalf("Expected app.EncryptionEnv() %q, got %q", "test_encryption_env", app.EncryptionEnv())
	}
}

func TestNewWithConfigAndFlags(t *testing.T) {
	// copy os.Args
	originalArgs := make([]string, len(os.Args))
	copy(originalArgs, os.Args)
	defer func() {
		// restore os.Args
		os.Args = originalArgs
	}()

	// change os.Args
	os.Args = os.Args[:1]
	os.Args = append(
		os.Args,
		"--dir=test_dir_flag",
		"--encryptionEnv=test_encryption_env_flag",
		"--debug=false",
	)

	app := NewWithConfig(Config{
		DefaultDataDir:       "test_dir",
		DefaultEncryptionEnv: "test_encryption_env",
		HideStartBanner:      true,
	})

	if app == nil {
		t.Fatal("Expected initialized PocketBase instance, got nil")
	}

	if app.RootCmd == nil {
		t.Fatal("Expected RootCmd to be initialized, got nil")
	}

	if app.appWrapper == nil {
		t.Fatal("Expected appWrapper to be initialized, got nil")
	}

	if app.hideStartBanner != true {
		t.Fatal("Expected app.hideStartBanner to be true, got false")
	}

	if app.DataDir() != "test_dir_flag" {
		t.Fatalf("Expected app.DataDir() %q, got %q", "test_dir_flag", app.DataDir())
	}

	if app.EncryptionEnv() != "test_encryption_env_flag" {
		t.Fatalf("Expected app.EncryptionEnv() %q, got %q", "test_encryption_env_flag", app.EncryptionEnv())
	}
}

func TestSkipBootstrap(t *testing.T) {
	// copy os.Args
	originalArgs := make([]string, len(os.Args))
	copy(originalArgs, os.Args)
	defer func() {
		// restore os.Args
		os.Args = originalArgs
	}()

	tempDir := filepath.Join(os.TempDir(), "temp_pb_data")
	defer os.RemoveAll(tempDir)

	// already bootstrapped
	app0 := NewWithConfig(Config{DefaultDataDir: tempDir})
	app0.Bootstrap()
	// TODO: fix later
	//if v := app0.skipBootstrap(); !v {
	//	t.Fatal("[bootstrapped] Expected true, got false")
	//}

	// unknown command
	os.Args = os.Args[:1]
	os.Args = append(os.Args, "demo")
	app1 := NewWithConfig(Config{DefaultDataDir: tempDir})
	app1.RootCmd.AddCommand(&cobra.Command{Use: "test"})
	if v := app1.skipBootstrap(); !v {
		t.Fatal("[unknown] Expected true, got false")
	}

	// default flags
	flagScenarios := []struct {
		name  string
		short string
	}{
		{"help", "h"},
		{"version", "v"},
	}

	for _, s := range flagScenarios {
		// base flag
		os.Args = os.Args[:1]
		os.Args = append(os.Args, "--"+s.name)
		app1 := NewWithConfig(Config{DefaultDataDir: tempDir})
		if v := app1.skipBootstrap(); !v {
			t.Fatalf("[--%s] Expected true, got false", s.name)
		}

		// short flag
		os.Args = os.Args[:1]
		os.Args = append(os.Args, "-"+s.short)
		app2 := NewWithConfig(Config{DefaultDataDir: tempDir})
		if v := app2.skipBootstrap(); !v {
			t.Fatalf("[-%s] Expected true, got false", s.short)
		}

		customCmd := &cobra.Command{Use: "custom"}
		customCmd.PersistentFlags().BoolP(s.name, s.short, false, "")

		// base flag in custom command
		os.Args = os.Args[:1]
		os.Args = append(os.Args, "custom")
		os.Args = append(os.Args, "--"+s.name)
		app3 := NewWithConfig(Config{DefaultDataDir: tempDir})
		app3.RootCmd.AddCommand(customCmd)
		if v := app3.skipBootstrap(); v {
			t.Fatalf("[--%s custom] Expected false, got true", s.name)
		}

		// short flag in custom command
		os.Args = os.Args[:1]
		os.Args = append(os.Args, "custom")
		os.Args = append(os.Args, "-"+s.short)
		app4 := NewWithConfig(Config{DefaultDataDir: tempDir})
		app4.RootCmd.AddCommand(customCmd)
		if v := app4.skipBootstrap(); v {
			t.Fatalf("[-%s custom] Expected false, got true", s.short)
		}
	}
}

func setupTestEnvironment() {
	test_utils.LoadTestEnv()
	// drop all table in test database
	db, err := sql.Open("postgres", os.Getenv("DATABASE"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	rows, err := db.Query("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'")
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Failed to scan table name: %v", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating over rows: %v", err)
	}

	fmt.Println("Dropping all tables in the database:")
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			log.Fatalf("Failed to drop table %s: %v", table, err)
		}
		fmt.Printf("Dropped table: %s\n", table)
	}

	fmt.Println("Setting up test environment")
	tempDir := filepath.Join(os.TempDir(), "temp_pb_data")
	app := NewWithConfig(Config{DefaultDataDir: tempDir})
	migratecmd.Register(app, nil, migratecmd.Config{
		Automigrate: true,
	})
	app.Bootstrap()
}

func TestMain(m *testing.M) {
	setupTestEnvironment()

	// 运行所有测试
	code := m.Run()

	os.Exit(code)
}

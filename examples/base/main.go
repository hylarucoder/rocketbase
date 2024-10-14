package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	pocketbase "github.com/hylarucoder/rocketbase"
	"github.com/hylarucoder/rocketbase/core"
	"github.com/hylarucoder/rocketbase/plugins/migratecmd"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	app := pocketbase.New()

	// ---------------------------------------------------------------
	// Optional plugin flags:
	// ---------------------------------------------------------------
	var automigrate bool
	app.RootCmd.PersistentFlags().BoolVar(
		&automigrate,
		"automigrate",
		false,
		"enable/disable auto migrations",
	)

	var queryTimeout int
	app.RootCmd.PersistentFlags().IntVar(
		&queryTimeout,
		"queryTimeout",
		30,
		"the default SELECT queries timeout in seconds",
	)

	app.RootCmd.ParseFlags(os.Args[1:])

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	// migrate command (with js templates)
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  automigrate,
		// Dir:          migrationsDir,
	})

	// // GitHub selfupdate
	// ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{})

	app.OnAfterBootstrap().PreAdd(func(e *core.BootstrapEvent) error {
		app.Dao().ModelQueryTimeout = time.Duration(queryTimeout) * time.Second
		return nil
	})

	// app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
	// 	// serves static files from the provided public dir (if exists)
	// 	e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(publicDir), indexFallback))
	// 	return nil
	// })

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

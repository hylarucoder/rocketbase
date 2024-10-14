//TODO: go:build pq

package core

import (
	_ "github.com/lib/pq"
	"github.com/pocketbase/dbx"
)

func ConnectDB(dsn string) (*dbx.DB, error) {
	return dbx.MustOpen("postgres", dsn)
}

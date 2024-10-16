package models

import "github.com/hylarucoder/rocketbase/tools/types"

type TableInfoRow struct {
	// the `db:"pk"` tag has special semantic so we cannot rename
	// the original field without specifying a custom mapper
	PK int

	Index        int           `db:"cid"`
	Name         string        `db:"column_name"`
	Type         string        `db:"type"`
	NotNull      bool          `db:"notnull"`
	DefaultValue types.JsonRaw `db:"column_default"`
}

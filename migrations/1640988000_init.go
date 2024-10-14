// Package migrations contains the system PocketBase DB migrations.
package migrations

import (
	"path/filepath"
	"runtime"

	"github.com/hylarucoder/rocketbase/daos"
	"github.com/hylarucoder/rocketbase/models"
	"github.com/hylarucoder/rocketbase/models/schema"
	"github.com/hylarucoder/rocketbase/models/settings"
	"github.com/hylarucoder/rocketbase/tools/migrate"
	"github.com/hylarucoder/rocketbase/tools/types"
	"github.com/pocketbase/dbx"
)

var AppMigrations migrate.MigrationsList

// Register is a short alias for `AppMigrations.Register()`
// that is usually used in external/user defined migrations.
func Register(
	up func(db dbx.Builder) error,
	down func(db dbx.Builder) error,
	optFilename ...string,
) {
	var optFiles []string
	if len(optFilename) > 0 {
		optFiles = optFilename
	} else {
		_, path, _, _ := runtime.Caller(1)
		optFiles = append(optFiles, filepath.Base(path))
	}
	AppMigrations.Register(up, down, optFiles...)
}

func init() {
	AppMigrations.Register(func(db dbx.Builder) error {
		_, tablesErr := db.NewQuery(`
			create or replace function json_valid(p_json text) returns boolean as $$
			begin
				return (p_json::json is not null);
				exception when others then return false;  
			end;
			$$ language plpgsql immutable;

			CREATE OR REPLACE FUNCTION json_extract(json_data json, key text)
			RETURNS text AS $$
			BEGIN
				RETURN json_data ->> key;
			END;
			$$ LANGUAGE plpgsql;

			CREATE SEQUENCE IF NOT EXISTS global_id_seq;
			CREATE OR REPLACE FUNCTION generate_snowflake(OUT result text) AS $$
			DECLARE
				our_epoch bigint := 1314220021721;
				seq_id bigint;
				now_millis bigint;
				shard_id int := 5;
				resultint bigint;
			BEGIN
				SELECT nextval('global_id_seq')::bigint % 1024 INTO seq_id;
				SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
				resultint := (now_millis - our_epoch) << 23;
				resultint := resultint | (shard_id <<10);
				resultint := resultint | (seq_id);
				-- convert result from bigint to text
				result := resultint::text;
			END;
				$$ LANGUAGE PLPGSQL;

			CREATE TABLE {{_admins}} (
				[[id]]        		VARCHAR(32) PRIMARY KEY DEFAULT generate_snowflake() NOT NULL,
				[[avatar]]          INTEGER DEFAULT 0 NOT NULL,
				[[email]]           TEXT UNIQUE NOT NULL,
				[[tokenKey]]        TEXT UNIQUE NOT NULL,
				[[passwordHash]]    TEXT NOT NULL,
				[[lastResetSentAt]] TEXT DEFAULT '' NOT NULL,
				[[created]]         TIMESTAMPTZ DEFAULT NOW() NOT NULL,
				[[updated]]         TIMESTAMPTZ DEFAULT NOW() NOT NULL
			);

			CREATE TABLE {{_collections}} (
				[[id]]		   VARCHAR(32) PRIMARY KEY DEFAULT generate_snowflake() NOT NULL,
				[[system]]     BOOLEAN DEFAULT FALSE NOT NULL,
				[[type]]       TEXT DEFAULT 'base' NOT NULL,
				[[name]]       TEXT UNIQUE NOT NULL,
				[[schema]]     JSON DEFAULT '[]' NOT NULL,
				[[indexes]]    JSON DEFAULT '[]' NOT NULL,
				[[listRule]]   TEXT DEFAULT NULL,
				[[viewRule]]   TEXT DEFAULT NULL,
				[[createRule]] TEXT DEFAULT NULL,
				[[updateRule]] TEXT DEFAULT NULL,
				[[deleteRule]] TEXT DEFAULT NULL,
				[[options]]    JSON DEFAULT '{}' NOT NULL,
				[[created]]    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
				[[updated]]    TIMESTAMPTZ DEFAULT NOW() NOT NULL
			);

			CREATE TABLE {{_params}} (
				[[id]]      VARCHAR(32) PRIMARY KEY DEFAULT generate_snowflake() NOT NULL,
				[[key]]     TEXT UNIQUE NOT NULL,
				[[value]]   JSON DEFAULT NULL,
				[[created]] TIMESTAMPTZ DEFAULT NOW() NOT NULL,
				[[updated]] TIMESTAMPTZ DEFAULT NOW() NOT NULL
			);

			CREATE TABLE {{_externalAuths}} (
				[[id]]           VARCHAR(32) PRIMARY KEY DEFAULT generate_snowflake() NOT NULL,
				[[collectionId]] TEXT NOT NULL,
				[[recordId]]     TEXT NOT NULL,
				[[provider]]     TEXT NOT NULL,
				[[providerId]]   TEXT NOT NULL,
				[[created]]      TIMESTAMPTZ DEFAULT NOW() NOT NULL,
				[[updated]]      TIMESTAMPTZ DEFAULT NOW() NOT NULL,
				---
				FOREIGN KEY ([[collectionId]]) REFERENCES {{_collections}} ([[id]]) ON UPDATE CASCADE ON DELETE CASCADE
			);

			CREATE UNIQUE INDEX _externalAuths_record_provider_idx on {{_externalAuths}} ([[collectionId]], [[recordId]], [[provider]]);
			CREATE UNIQUE INDEX _externalAuths_collection_provider_idx on {{_externalAuths}} ([[collectionId]], [[provider]], [[providerId]]);
		`).Execute()
		if tablesErr != nil {
			return tablesErr
		}

		dao := daos.New(db)

		// inserts default settings
		// -----------------------------------------------------------
		defaultSettings := settings.New()
		if err := dao.SaveSettings(defaultSettings); err != nil {
			return err
		}

		// inserts the system users collection
		// -----------------------------------------------------------
		usersCollection := &models.Collection{}
		usersCollection.MarkAsNew()
		usersCollection.Id = "_pb_users_auth_"
		usersCollection.Name = "users"
		usersCollection.Type = models.CollectionTypeAuth
		usersCollection.ListRule = types.Pointer("id = @request.auth.id")
		usersCollection.ViewRule = types.Pointer("id = @request.auth.id")
		usersCollection.CreateRule = types.Pointer("")
		usersCollection.UpdateRule = types.Pointer("id = @request.auth.id")
		usersCollection.DeleteRule = types.Pointer("id = @request.auth.id")

		// set auth options
		usersCollection.SetOptions(models.CollectionAuthOptions{
			ManageRule:        nil,
			AllowOAuth2Auth:   true,
			AllowUsernameAuth: true,
			AllowEmailAuth:    true,
			MinPasswordLength: 8,
			RequireEmail:      false,
		})

		// set optional default fields
		usersCollection.Schema = schema.NewSchema(
			&schema.SchemaField{
				Id:      "users_name",
				Type:    schema.FieldTypeText,
				Name:    "name",
				Options: &schema.TextOptions{},
			},
			&schema.SchemaField{
				Id:   "users_avatar",
				Type: schema.FieldTypeFile,
				Name: "avatar",
				Options: &schema.FileOptions{
					MaxSelect: 1,
					MaxSize:   5242880,
					MimeTypes: []string{
						"image/jpeg",
						"image/png",
						"image/svg+xml",
						"image/gif",
						"image/webp",
					},
				},
			},
		)

		return dao.SaveCollection(usersCollection)
	}, func(db dbx.Builder) error {
		tables := []string{
			"users",
			"_externalAuths",
			"_params",
			"_collections",
			"_admins",
		}

		for _, name := range tables {
			if _, err := db.DropTable(name).Execute(); err != nil {
				return err
			}
		}

		return nil
	})
}

package logs

import (
	"github.com/hylarucoder/rocketbase/tools/migrate"
	"github.com/pocketbase/dbx"
)

var LogsMigrations migrate.MigrationsList

func init() {
	LogsMigrations.Register(func(db dbx.Builder) error {
		_, err := db.NewQuery(`
			CREATE OR REPLACE FUNCTION json_extract(json_data json, key text)
			RETURNS text AS $$
			BEGIN
				RETURN json_data ->> key;
			END;
			$$ LANGUAGE plpgsql;

			-- Create a new IMMUTABLE function that wraps date_trunc
			CREATE OR REPLACE FUNCTION immutable_date_trunc(text, timestamp with time zone) RETURNS timestamp with time zone AS $$
			BEGIN
				RETURN date_trunc($1, $2);
			END;
			$$ LANGUAGE plpgsql IMMUTABLE;


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


			CREATE TABLE {{_requests}} (
				[[id]]        VARCHAR(32) PRIMARY KEY DEFAULT generate_snowflake() NOT NULL,
				[[url]]       TEXT DEFAULT '' NOT NULL,
				[[method]]    TEXT DEFAULT 'get' NOT NULL,
				[[status]]    INTEGER DEFAULT 200 NOT NULL,
				[[auth]]      TEXT DEFAULT 'guest' NOT NULL,
				[[ip]]        TEXT DEFAULT '127.0.0.1' NOT NULL,
				[[referer]]   TEXT DEFAULT '' NOT NULL,
				[[userAgent]] TEXT DEFAULT '' NOT NULL,
				[[meta]]      JSON DEFAULT '{}' NOT NULL,
				[[created]]   TIMESTAMPTZ DEFAULT NOW() NOT NULL,
				[[updated]]   TIMESTAMPTZ DEFAULT NOW() NOT NULL
			);

			CREATE INDEX _request_status_idx on {{_requests}} ([[status]]);
			CREATE INDEX _request_auth_idx on {{_requests}} ([[auth]]);
			CREATE INDEX _request_ip_idx on {{_requests}} ([[ip]]);
			CREATE INDEX _request_created_hour_idx on {{_requests}} (immutable_date_trunc('hour', [[created]]));
		`).Execute()

		return err
	}, func(db dbx.Builder) error {
		_, err := db.DropTable("_requests").Execute()
		return err
	})
}

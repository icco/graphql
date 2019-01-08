package graphql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GuiaBolso/darwin"
	"github.com/opencensus-integrations/ocsql"

	// Needed to talk to postgres
	_ "github.com/lib/pq"
)

var (
	db         *sql.DB
	driver     = "postgres"
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Creating table posts",
			Script: `
      CREATE TABLE posts (
        id serial primary key,
        title text,
        content text,
        date timestamp with time zone,
        tags text[],
        draft boolean,
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
		{
			Version:     2,
			Description: "Creating table stats",
			Script: `
      CREATE TABLE stats (
        id serial primary key,
        key text,
        value text,
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
		{
			Version:     3,
			Description: "Creating table users",
			Script: `
      CREATE TABLE users(
        id serial primary key,
        role text,
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
		{
			Version:     4,
			Description: "Cleanup users",
			Script: `
      DROP TABLE IF EXISTS users;
      DROP TABLE IF EXISTS auth_identities;
      CREATE TABLE users(
        id text primary key,
        role text,
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
		{
			Version:     5,
			Description: "Add links table",
			Script: `
      CREATE EXTENSION pgcrypto;
      CREATE TABLE links(
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        title text,
        uri text,
        created timestamp with time zone,
        description text,
        screenshot text,
        tags text[],
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
		{
			Version:     6,
			Description: "Make things unique",
			Script: `
      ALTER TABLE links ADD CONSTRAINT links_uri_key UNIQUE (uri);
      ALTER TABLE stats ADD CONSTRAINT stats_key_key UNIQUE (key);
      `,
		},
		{
			Version:     7,
			Description: "Add API keys",
			Script: `
      ALTER TABLE users ADD COLUMN apikey UUID DEFAULT gen_random_uuid();
      `,
		},
		{
			Version:     8,
			Description: "Add tweets table",
			Script: `
      CREATE TABLE tweets(
        id text PRIMARY KEY NOT NULL,
        text text,
        hashtags text[],
        symbols text[],
        user_mentions text[],
        urls text[],
        user text,
        favorites bigint,
        retweets bigint,
        posted timestamp with time zone,
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
	}
)

// InitDB creates a package global db connection from a database string.
func InitDB(dataSourceName string) (*sql.DB, error) {
	var err error

	// Connect to Database
	wrappedDriver, err := ocsql.Register(driver, ocsql.WithAllTraceOptions())
	if err != nil {
		return nil, fmt.Errorf("Failed to register the ocsql driver: %v", err)
	}

	db, _ = sql.Open(wrappedDriver, dataSourceName)
	if err = db.PingContext(context.Background()); err != nil {
		return nil, err
	}

	// Migrate
	driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
	d := darwin.New(driver, migrations, nil)
	err = d.Migrate()
	if err != nil {
		return nil, err
	}

	return db, err
}

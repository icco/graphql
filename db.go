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
	dbDriver   = "postgres"
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
        screen_name text,
        favorites bigint,
        retweets bigint,
        posted timestamp with time zone,
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
		{
			Version:     9,
			Description: "Add books table",
			Script: `
      CREATE TABLE books(
        id text PRIMARY KEY NOT NULL,
        title text,
        goodreads_id text,
        created_at timestamp with time zone,
        modified_at timestamp with time zone
      );
      `,
		},
		{
			Version:     10,
			Description: "Add logs table",
			Script: `
      CREATE EXTENSION postgis;
      CREATE EXTENSION postgis_topology;
      CREATE TABLE logs(
        id text PRIMARY KEY NOT NULL,
        code TEXT,
        datetime TIMESTAMP WITH TIME ZONE,
        description TEXT,
        location GEOGRAPHY(POINT),
        project TEXT,
        user_id TEXT,
        created_at TIMESTAMP WITH TIME ZONE,
        modified_at TIMESTAMP WITH TIME ZONE
      );
      `,
		},
		{
			Version:     11,
			Description: "Add photos table",
			Script: `
      CREATE TABLE photos (
        id TEXT PRIMARY KEY NOT NULL,
        year INT,
        user_id TEXT,
        content_type TEXT,
        created_at TIMESTAMP WITH TIME ZONE,
        modified_at TIMESTAMP WITH TIME ZONE
      );
      `,
		},
		{
			Version:     12,
			Description: "Add pages table",
			Script: `
      CREATE TABLE pages (
        id TEXT PRIMARY KEY NOT NULL,
        slug TEXT,
        title TEXT,
        content TEXT,
        category TEXT,
        tags TEXT[],
        user_id TEXT,
        created_at TIMESTAMP WITH TIME ZONE,
        modified_at TIMESTAMP WITH TIME ZONE
      );
      `,
		},
		{
			Version:     13,
			Description: "Add trgm",
			Script: `
      CREATE EXTENSION pg_trgm;
      SELECT set_limit(0.6);
      `,
		},
		{
			Version:     14,
			Description: "Add trgm index",
			Script: `
      CREATE INDEX content_gin_idx ON posts USING GIN(content gin_trgm_ops);
      `,
		},
		{
			Version:     15,
			Description: "Add second trgm index",
			Script: `
      CREATE INDEX title_gin_idx ON posts USING GIN(title gin_trgm_ops);
      `,
		},
		{
			Version:     16,
			Description: "Add comments table",
			Script: `
      CREATE TABLE comments (
        id TEXT PRIMARY KEY NOT NULL,
        post_id BIGINT,
        user_id TEXT NOT NULL,
        content TEXT,
        created_at TIMESTAMP WITH TIME ZONE,
        modified_at TIMESTAMP WITH TIME ZONE
      );
      `,
		},
		{
			Version:     17,
			Description: "Add cache table",
			Script: `
      CREATE TABLE cache (
        key TEXT PRIMARY KEY NOT NULL,
        value TEXT NOT NULL,
        modified_at TIMESTAMP WITH TIME ZONE
      );
      `,
		},
		{
			Version:     18,
			Description: "Add name to user",
			Script: `
      ALTER TABLE users ADD COLUMN name Text DEFAULT 'anonymous';
      `,
		},
		{
			Version:     19,
			Description: "add serial to tweets",
			Script:      `ALTER TABLE tweets ADD COLUMN IF NOT EXISTS internal_id BIGSERIAL NOT NULL`,
		},
		{
			Version:     20,
			Description: "alter stats value",
			Script:      `ALTER TABLE stats ALTER COLUMN value TYPE float USING (value::float)`,
		},
		{
			Version:     21,
			Description: "new stats table",
			Script: `
      DROP TABLE stats;
      CREATE TABLE stats (
        id SERIAL PRIMARY KEY,
        key TEXT,
        value FLOAT,
        inserted_at TIMESTAMP WITH TIME ZONE
      );
      CREATE INDEX stats_key_inserted_at_idx ON stats (key, inserted_at DESC);
      `,
		},
		{
			Version:     22,
			Description: "add serial index to tweets",
			Script:      `CREATE INDEX ON tweets (internal_id);`,
		},
		{
			Version:     23,
			Description: "new page table",
			Script: `
      DROP TABLE logs;
      DROP TABLE pages;
      CREATE TABLE pages (
        id SERIAL PRIMARY KEY,
        slug TEXT NOT NULL,
        content TEXT,
        user_id TEXT NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE,
        modified_at TIMESTAMP WITH TIME ZONE
      );
      `,
		},
		{
			Version:     24,
			Description: "unique constraint for pages",
			Script:      `ALTER TABLE pages ADD UNIQUE (slug, user_id);`,
		},
	}
)

// InitDB creates a package global db connection from a database string.
func InitDB(dataSourceName string) (*sql.DB, error) {
	var err error

	// Connect to Database
	wrappedDriver, err := ocsql.Register(dbDriver, ocsql.WithAllTraceOptions())
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

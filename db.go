package graphql

import (
	"context"
	"database/sql"
	"log"

	"github.com/GuiaBolso/darwin"
	"github.com/basvanbeek/ocsql"

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
	}
)

// InitDB creates a package global db connection from a database string.
func InitDB(dataSourceName string) *sql.DB {
	var err error

	// Connect to Database
	wrappedDriver, err := ocsql.Register(driver, ocsql.WithAllTraceOptions())
	if err != nil {
		log.Fatalf("Failed to register the ocsql driver: %v", err)
	}

	db, _ = sql.Open(wrappedDriver, dataSourceName)
	if err = db.PingContext(context.Background()); err != nil {
		log.Panic(err)
	}

	// Migrate
	driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
	d := darwin.New(driver, migrations, nil)
	err = d.Migrate()
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Connected to %+v", dataSourceName)
	return db
}

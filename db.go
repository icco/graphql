package writing

import (
	"database/sql"
	"log"

	"github.com/GuiaBolso/darwin"
	_ "github.com/lib/pq"
)

var (
	db         *sql.DB
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
	}
)

func InitDB(dataSourceName string) {
	var err error

	// Connect to Database
	db, err = sql.Open("postgres", dataSourceName)
	if err = db.Ping(); err != nil {
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
}

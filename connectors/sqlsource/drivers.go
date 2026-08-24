package sqlsource

// Register the database/sql drivers this connector dials. Keeping the blank
// imports in the connector package makes it self-sufficient: importing
// sqlsource brings the drivers it needs.
import (
	_ "github.com/go-sql-driver/mysql" // "mysql" — MySQL and Dolt (MySQL wire)
	_ "github.com/lib/pq"              // "postgres"
	_ "github.com/mattn/go-sqlite3"    // "sqlite3"
)

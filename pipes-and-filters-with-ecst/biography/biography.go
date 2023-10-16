package biography

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

// Biography struct represents the data structure of the Biography table
type Biography struct {
	Name    string
	Content string
}

// GetBiography retrieves the biography from the database based on the given name
func GetBiography(db *sql.DB, name string) (string, error) {
	query := "SELECT Biography FROM Biography WHERE Name = ?"
	row := db.QueryRow(query, name)

	var content string
	err := row.Scan(&content)
	if err != nil {
		return "", err
	}

	return content, nil
}

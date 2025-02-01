package database

import (
	"database/sql"
	"log"
	"testing"
)

// Simple tests to see if database works

type Country struct {
	COUNTRY_ISO2_CODEID string `json:"countryISO2"`
	COUNTRY_NAME        string `json:"countryName"`
}

func TestExistingQuery(t *testing.T) {
	db, err := Connect()
	if err != nil {
		log.Println(err)
	}
	query := `SELECT * FROM countries WHERE country_name = "POLAND"`
	want := Country{"PL", "POLAND"}
	rows := db.QueryRow(query)
	var got Country
	if err := rows.Scan(&got.COUNTRY_ISO2_CODEID, &got.COUNTRY_NAME); err != nil {
		t.Fatalf(`Query("SELECT * FROM countries WHERE country_name = "POLAND"") failed: %v`, err)
	}
	if got.COUNTRY_ISO2_CODEID != want.COUNTRY_ISO2_CODEID || got.COUNTRY_NAME != want.COUNTRY_NAME {
		t.Fatalf(`Query("SELECT * FROM countries WHERE country_name = "POLAND"") returned: %v %v, expected: %v, %v`,
			got.COUNTRY_ISO2_CODEID, got.COUNTRY_NAME, want.COUNTRY_ISO2_CODEID, want.COUNTRY_NAME)
	}
}

func TestNotExistingQuery(t *testing.T) {
	db, err := Connect()
	if err != nil {
		log.Println(err)
	}
	query := `SELECT * FROM countries WHERE country_name = "ABC"`
	rows := db.QueryRow(query)
	var got Country
	if err := rows.Scan(&got.COUNTRY_ISO2_CODEID, &got.COUNTRY_NAME); err != sql.ErrNoRows {
		t.Fatalf(`Query("SELECT * FROM countries WHERE country_name = "ABC"") failed: %v, expected sql.ErrNoRows`, err)
	}
}

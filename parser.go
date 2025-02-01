package parser

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type Record struct {
	COUNTRY_ISO2_CODEID string `json:"countryISO2"`
	SWIFT_CODE          string `json:"swiftCode"`
	CODE_TYPE           string `json:"codeType"`
	NAME                string `json:"bankName"`
	ADDRESS             string `json:"address"`
	TOWN_NAME           string `json:"townName"`
	COUNTRY_NAME        string `json:"countryName"`
	TIME_ZONE           string `json:"timeZone"`
	IS_HEADQUARTER      bool   `json:"isHeadquarter"`
}

type CountryRecord struct {
	COUNTRY_ISO2_CODEID string `json:"countryISO2"`
	COUNTRY_NAME        string `json:"countryName"`
}

// Both parse and insert data into database
func Parse(db *sql.DB) error {
	// Open and parse CSV file
	parsedRecords, err := ParseCSV("../../internal/data/SWIFT_CODES.csv")
	if err != nil {
		fmt.Printf("Failed to parseCSV: %v\n", err)
		return err
	}

	// Insert countries first to avoid foreign key errors
	err = InsertCountries(db, parsedRecords)
	if err != nil {
		fmt.Printf("Failed to insert countries to database: %v\n", err)
		return err
	}

	// Insert branches to database
	err = InsertBranches(db, parsedRecords)
	if err != nil {
		fmt.Printf("Failed to insert branches to database: %v\n", err)
		return err
	}
	return nil
}

func ParseCSV(filePath string) ([]Record, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	var parsedRecords []Record
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		parsedRecords = append(parsedRecords, Record{
			COUNTRY_ISO2_CODEID: record[0],
			SWIFT_CODE:          record[1],
			CODE_TYPE:           record[2],
			NAME:                record[3],
			ADDRESS:             record[4],
			TOWN_NAME:           record[5],
			COUNTRY_NAME:        record[6],
			TIME_ZONE:           record[7],
			IS_HEADQUARTER:      strings.HasSuffix(record[1], "XXX"),
		})
	}
	return parsedRecords, nil
}

func InsertCountries(db *sql.DB, records []Record) error {
	countrySet := map[string]CountryRecord{}

	for _, record := range records {
		if _, exists := countrySet[record.COUNTRY_NAME]; !exists {
			countrySet[record.COUNTRY_NAME] = CountryRecord{
				COUNTRY_ISO2_CODEID: record.COUNTRY_ISO2_CODEID,
				COUNTRY_NAME:        record.COUNTRY_NAME,
			}
		}
	}

	for _, country := range countrySet {
		if err := insertCountry(db, country); err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
				log.Print("Duplicate entry, ignoring error.")
			} else {
				log.Fatalf("Database error: %v", err)
			}
		}
	}
	return nil
}

func InsertBranches(db *sql.DB, records []Record) error {
	for _, record := range records {
		if err := insertRecord(db, record); err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
				log.Print("Duplicate entry, ignoring error.")
			} else {
				log.Fatalf("Database error: %v", err)
			}
		}
	}
	return nil
}

// Helper function to insert a record into the database
func insertRecord(db *sql.DB, record Record) error {
	query := `INSERT INTO branches (swift_code, name, town_name, address, time_zone, country_iso2, is_headquarter) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, record.SWIFT_CODE, record.NAME, record.TOWN_NAME, record.ADDRESS, record.TIME_ZONE, record.COUNTRY_ISO2_CODEID, strings.HasSuffix(record.SWIFT_CODE, "XXX"))
	return err
}

// Helper function to insert a country into the database
func insertCountry(db *sql.DB, record CountryRecord) error {
	query := `INSERT INTO countries (country_iso2, country_name) VALUES (?, ?)`
	_, err := db.Exec(query, record.COUNTRY_ISO2_CODEID, record.COUNTRY_NAME)
	return err
}

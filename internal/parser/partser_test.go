package parser

import (
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestParseCSV(t *testing.T) {
	csvContent := `COUNTRY_ISO2_CODEID,SWIFT_CODE,CODE_TYPE,NAME,ADDRESS,TOWN_NAME,COUNTRY_NAME,TIME_ZONE
					PL,ABCABCABCAB,BANK,ABC BANK,Main Street,Warsaw,POLAND,Europe/Warsaw
					US,DEFDEFDEFDEFXXX,BANK,DEF BANK,Wall Street,New York,USA,America/New_York`

	tempFile, err := os.CreateTemp("", "test_swift_codes_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	tempFile.WriteString(csvContent)
	tempFile.Close()

	records, err := ParseCSV(tempFile.Name())
	assert.NoError(t, err, "ParseCSV should not return an error")
	assert.Len(t, records, 2, "Should parse two records")
}

func TestInsertCountries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	records := []Record{
		{COUNTRY_ISO2_CODEID: "PL", COUNTRY_NAME: "POLAND"},
		{COUNTRY_ISO2_CODEID: "US", COUNTRY_NAME: "USA"},
	}

	mock.ExpectExec("^INSERT INTO countries.*").WithArgs("PL", "POLAND").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("^INSERT INTO countries.*").WithArgs("US", "USA").WillReturnResult(sqlmock.NewResult(1, 1))

	err = InsertCountries(db, records)
	assert.NoError(t, err, "InsertCountries should not return an error")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}
func TestInsertBranches(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	records := []Record{
		{
			COUNTRY_ISO2_CODEID: "PL",
			SWIFT_CODE:          "ABCABCABCAB",
			NAME:                "ABC BANK",
			TOWN_NAME:           "Warsaw",
			ADDRESS:             "Main Street",
			TIME_ZONE:           "Europe/Warsaw",
			IS_HEADQUARTER:      false,
		},
		{
			COUNTRY_ISO2_CODEID: "US",
			SWIFT_CODE:          "DEFDEFDEFDEFXXX",
			NAME:                "DEF BANK",
			TOWN_NAME:           "New York",
			ADDRESS:             "Wall Street",
			TIME_ZONE:           "America/New_York",
			IS_HEADQUARTER:      true,
		},
	}

	mock.ExpectExec("^INSERT INTO branches.*").WithArgs("ABCABCABCAB", "ABC BANK", "Warsaw", "Main Street", "Europe/Warsaw", "PL", false).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("^INSERT INTO branches.*").WithArgs("DEFDEFDEFDEFXXX", "DEF BANK", "New York", "Wall Street", "America/New_York", "US", true).WillReturnResult(sqlmock.NewResult(1, 1))

	err = InsertBranches(db, records)
	assert.NoError(t, err, "InsertBranches should not return an error")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met: %v", err)
	}
}

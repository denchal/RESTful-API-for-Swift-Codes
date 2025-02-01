package main

import (
	"Michal_Gomulczak_Assessment/SWIFT-API/internal/database"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"Michal_Gomulczak_Assessment/SWIFT-API/internal/parser"

	"github.com/gin-gonic/gin"
)

type Branch struct {
	ADDRESS             string `json:"address"`
	NAME                string `json:"bankName"`
	COUNTRY_ISO2_CODEID string `json:"countryISO2"`
	COUNTRY_NAME        string `json:"countryName"`
	IS_HEADQUARTER      bool   `json:"isHeadquarter"`
	SWIFT_CODE          string `json:"swiftCode"`
}

type CountryBranch struct {
	ADDRESS             string `json:"address"`
	NAME                string `json:"bankName"`
	COUNTRY_ISO2_CODEID string `json:"countryISO2"`
	IS_HEADQUARTER      bool   `json:"isHeadquarter"`
	SWIFT_CODE          string `json:"swiftCode"`
}

type Headquarter struct {
	ADDRESS            string   `json:"address"`
	NAME               string   `json:"bankName"`
	OUNTRY_ISO2_CODEID string   `json:"countryISO2"`
	COUNTRY_NAME       string   `json:"countryName"`
	IS_HEADQUARTER     bool     `json:"isHeadquarter"`
	SWIFT_CODE         string   `json:"swiftCode"`
	BRANCHES           []Branch `json:"branches"`
}

type Country struct {
	COUNTRY_ISO2_CODEID string          `json:"countryISO2"`
	COUNTRY_NAME        string          `json:"countryName"`
	SWIFT_CODES         []CountryBranch `json:"swiftCodes"`
}

type MessageResponse struct {
	MESSAGE string `json:"message"`
}

var db *sql.DB

func main() {
	gin.SetMode(gin.ReleaseMode)

	var err error
	db, err = database.Connect()
	if err != nil {
		log.Println(err)
		return
	}
	err = parser.Parse(db)
	if err != nil {
		log.Println(err) // It's normal to get errors here since some data might be already parsed
	}

	router := gin.Default()
	router.GET("/v1/swift-codes/:swift-code", getBranchBySwift)
	router.GET("/v1/swift-codes/country/:countryISO2code", getBranchesByCountry)
	router.POST("/v1/swift-codes/", postBranch)
	router.DELETE("/v1/swift-codes/:swift-code", deleteBranch)
	router.Run("0.0.0.0:8080")
}

func deleteBranch(c *gin.Context) {
	swift := c.Param("swift-code")

	query := `DELETE FROM branches WHERE swift_code = ?`
	res, err := db.Exec(query, swift)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to delete swift " + swift + " from database"})
		log.Println(err)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to delete swift " + swift + " from database"})
		log.Println(err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Succesfully deleted branch from database!"})
}

func postBranch(c *gin.Context) {
	var branch Branch

	if err := c.BindJSON(&branch); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to bind JSON, is data correct?"})
		log.Println(err)
		return
	}

	if branch.ADDRESS == "" || branch.COUNTRY_ISO2_CODEID == "" || branch.COUNTRY_NAME == "" || branch.NAME == "" || branch.SWIFT_CODE == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to bind JSON, is data complete?"})
		return
	}

	// Run a query to check if bank is in country which is not in database
	query := `SELECT * FROM countries WHERE country_iso2 = ?`
	_, err := db.Query(query, branch.COUNTRY_ISO2_CODEID)
	if err == sql.ErrNoRows {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to insert branch: Wrong country code"})
		log.Println(err)
		return
	}

	// Insert new branch to database
	query = `INSERT INTO branches (address, name, country_iso2, is_headquarter, swift_code) VALUES (?, ?, ?, ?, ?)`
	_, err = db.Exec(query, branch.ADDRESS, branch.NAME, branch.COUNTRY_ISO2_CODEID, branch.IS_HEADQUARTER, branch.SWIFT_CODE)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to insert branch: Already exists or wrong data " + branch.SWIFT_CODE})
		log.Println(err)
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Succesfully added branch to database!"})
	}
}

func getBranchesByCountry(c *gin.Context) {
	country_code := c.Param("countryISO2code")

	// Run a query to get country name
	rows := db.QueryRow(`
	SELECT country_iso2, country_name
	FROM countries 
	WHERE country_iso2 = ?`, country_code)
	var country Country

	if err := rows.Scan(&country.COUNTRY_ISO2_CODEID, &country.COUNTRY_NAME); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to query country name from code " + country_code})
		log.Println(err)
		return
	}

	var countryBranches []CountryBranch

	// Query country swift codes
	countryRows, err := db.Query(`
		SELECT address, name, branches.country_iso2, is_headquarter, swift_code
		FROM branches 
		INNER JOIN countries ON branches.country_iso2 = countries.country_iso2 
		WHERE branches.country_iso2 = ?`, country_code)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to query country swift codes: " + country_code})
		log.Println(err)
		return
	}

	for countryRows.Next() {
		var countryBranch CountryBranch
		if err := countryRows.Scan(&countryBranch.ADDRESS, &countryBranch.NAME, &countryBranch.COUNTRY_ISO2_CODEID,
			&countryBranch.IS_HEADQUARTER, &countryBranch.SWIFT_CODE); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to extract data from query"})
			log.Println(err)
			return
		}
		countryBranches = append(countryBranches, countryBranch)
	}

	country.SWIFT_CODES = countryBranches

	c.IndentedJSON(http.StatusOK, country)
}

func getBranchBySwift(c *gin.Context) {
	swift := c.Param("swift-code")

	// Query to get "base" branch, either headquarter or branch
	rows := db.QueryRow(`
	SELECT address, name, branches.country_iso2, country_name, is_headquarter 
	FROM branches 
	INNER JOIN countries ON branches.country_iso2 = countries.country_iso2 
	WHERE branches.swift_code = ?`, swift)

	var branch Branch
	if err := rows.Scan(&branch.ADDRESS, &branch.NAME, &branch.COUNTRY_ISO2_CODEID,
		&branch.COUNTRY_NAME, &branch.IS_HEADQUARTER); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to extract data from query"})
		log.Println(err)
		return
	}
	branch.SWIFT_CODE = swift
	if !branch.IS_HEADQUARTER {
		c.IndentedJSON(http.StatusOK, branch)
		return
	}

	// If base branch is a headquarter
	if branch.IS_HEADQUARTER {
		swiftPrefix, _ := strings.CutSuffix(swift, "XXX")
		var branches []Branch

		// Querry all branches under a headquarter
		branchRows, err := db.Query(`
		SELECT address, name, branches.country_iso2, country_name, swift_code, is_headquarter 
		FROM branches 
		INNER JOIN countries ON branches.country_iso2 = countries.country_iso2 
		WHERE swift_code LIKE ? AND swift_code NOT LIKE "%XXX"`, swiftPrefix+"%")
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to query branches under a headquarter " + branch.SWIFT_CODE})
			log.Println(err)
			return
		}

		for branchRows.Next() {
			var hqBranch Branch
			if err := branchRows.Scan(&hqBranch.ADDRESS, &hqBranch.NAME, &hqBranch.COUNTRY_ISO2_CODEID,
				&hqBranch.COUNTRY_NAME, &hqBranch.SWIFT_CODE, &hqBranch.IS_HEADQUARTER); err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Failed to extract data from query"})
				log.Println(err)
				return
			}
			branches = append(branches, hqBranch)
		}

		headquarter := Headquarter{branch.ADDRESS, branch.NAME, branch.COUNTRY_ISO2_CODEID,
			branch.COUNTRY_NAME, branch.IS_HEADQUARTER, branch.SWIFT_CODE, branches}

		c.IndentedJSON(http.StatusOK, headquarter)
	}
}

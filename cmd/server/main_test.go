package main

import (
	"Michal_Gomulczak_Assessment/SWIFT-API/internal/database"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetBranchBySwift(t *testing.T) {
	var err error
	db, err = database.Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/v1/swift-codes/:swift-code", getBranchBySwift)

	// Test 1: Valid Swift Code for a Branch
	t.Run("Valid Branch", func(t *testing.T) {
		// Assuming "AIZKLV22CLN" exists as a branch in the database
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/AIZKLV22CLN", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusOK)
		}

		var branch Branch
		err := json.Unmarshal(resp.Body.Bytes(), &branch)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		if branch.SWIFT_CODE != "AIZKLV22CLN" {
			t.Errorf("Unexpected SWIFT_CODE: got %v, want %v", branch.SWIFT_CODE, "AIZKLV22CLN")
		}
	})

	// Test 2: Valid Swift Code for a Headquarter
	t.Run("Valid Headquarter", func(t *testing.T) {
		// Assuming "AIZKLV22XXX" exists and has associated branches
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/AIZKLV22XXX", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusOK)
		}

		var headquarter Headquarter
		err := json.Unmarshal(resp.Body.Bytes(), &headquarter)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		if !headquarter.IS_HEADQUARTER {
			t.Errorf("Expected IS_HEADQUARTER to be true, got false")
		}

		if len(headquarter.BRANCHES) == 0 {
			t.Errorf("Expected branches under headquarter, but got none")
		}
	})

	// Test 3: Invalid Swift Code
	t.Run("Invalid Swift Code", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/INVALIDSWIF", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusBadRequest)
		}

		var errorResponse map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Could not decode error response: %v", err)
		}

		if errorResponse["message"] != "Failed to extract data from query" {
			t.Errorf("Unexpected error message: got %v, want %v", errorResponse["message"], "Failed to extract data from query")
		}
	})

	// Test 4: Valid Headquarter without Branches
	t.Run("Valid Headquarter Without Branches", func(t *testing.T) {
		// Assuming "AAAJBG21XXX" exists as a headquarter without associated branches
		req, _ := http.NewRequest(http.MethodGet, "/v1/swift-codes/AAAJBG21XXX", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusOK)
		}

		var headquarter Headquarter
		err := json.Unmarshal(resp.Body.Bytes(), &headquarter)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		if len(headquarter.BRANCHES) != 0 {
			t.Errorf("Expected no branches under headquarter, but got %v", len(headquarter.BRANCHES))
		}
	})
}

func TestGetBranchesByCountry(t *testing.T) {
	var err error
	db, err = database.Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/v1/countries/:countryISO2code", getBranchesByCountry)

	// Test 1: Valid Country Code with Branches
	t.Run("Valid Country with Branches", func(t *testing.T) {
		// Assuming "PL" exists and has associated branches
		req, _ := http.NewRequest(http.MethodGet, "/v1/countries/PL", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusOK)
		}

		var country Country
		err := json.Unmarshal(resp.Body.Bytes(), &country)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		if country.COUNTRY_ISO2_CODEID != "PL" {
			t.Errorf("Unexpected COUNTRY_ISO2_CODEID: got %v, want %v", country.COUNTRY_ISO2_CODEID, "PL")
		}

		if len(country.SWIFT_CODES) == 0 {
			t.Errorf("Expected branches for country, but got none")
		}
	})

	// Test 2: Invalid Country Code
	t.Run("Invalid Country Code", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/countries/QQ", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusBadRequest)
		}

		var errorResponse map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Could not decode error response: %v", err)
		}

		if errorResponse["message"] != "Failed to query country name from code QQ" {
			t.Errorf("Unexpected error message: got %v, want %v", errorResponse["message"], "Failed to query country name from code QQ")
		}
	})

	// Test 3: Database Query Error
	t.Run("Database Query Error", func(t *testing.T) {
		// Simulate database connection failure or query error
		db.Close() // Close the database connection to simulate an error
		req, _ := http.NewRequest(http.MethodGet, "/v1/countries/PL", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusBadRequest)
		}

		var errorResponse map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Could not decode error response: %v", err)
		}

		if _, exists := errorResponse["message"]; !exists {
			t.Errorf("Error message missing in response")
		}
	})
}

func TestPostBranch(t *testing.T) {
	var err error
	db, err = database.Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/v1/swift-codes/", postBranch)

	// Test 1: Correct input data
	t.Run("Valid Request", func(t *testing.T) {
		requestBody, _ := json.Marshal(Branch{
			ADDRESS:             "abc",
			NAME:                "ABC BANK",
			COUNTRY_ISO2_CODEID: "PL",
			COUNTRY_NAME:        "POLAND",
			IS_HEADQUARTER:      false,
			SWIFT_CODE:          "ABCABCABCAB",
		})
		req, _ := http.NewRequest(http.MethodPost, "/v1/swift-codes/", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		// Check response code
		if resp.Code != http.StatusOK {
			t.Errorf("Unexpected status code: got %v want %v", resp.Code, http.StatusOK)
		}

		// Check response value, should be expectedMessage
		var response MessageResponse
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}
		expectedMessage := "Succesfully added branch to database!"
		if response.MESSAGE != expectedMessage {
			t.Errorf("Unexpected response message: got %v want %v", response.MESSAGE, expectedMessage)
		}
	})

	// Test 2: Incomplete Data
	t.Run("Missing Fields", func(t *testing.T) {
		requestBody, _ := json.Marshal(map[string]string{
			"address":     "Example Bank",
			"countryISO2": "PL",
		})
		req, _ := http.NewRequest(http.MethodPost, "/v1/swift-codes/", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Check status, should be 400
		if resp.Code != http.StatusBadRequest {
			t.Errorf("Unexpected status code: got %v want %v", resp.Code, http.StatusBadRequest)
		}

		// Check if response is missing error "message" field
		var errorResponse map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Could not decode error response: %v", err)
		}
		if _, exists := errorResponse["message"]; !exists {
			t.Errorf("Error field missing in response")
		}
	})
}

func TestDeleteBranch(t *testing.T) {
	var err error
	db, err = database.Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/v1/swift-codes/:swift-code", deleteBranch)

	// Test 1: Valid Swift Code
	t.Run("Valid Swift Code", func(t *testing.T) {
		// Assuming "ABCABCABCAB" exists in the database (was created by post before)
		req, _ := http.NewRequest(http.MethodDelete, "/v1/swift-codes/ABCABCABCAB", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusOK)
		}

		var response map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Could not decode response: %v", err)
		}

		if response["message"] != "Succesfully deleted branch from database!" {
			t.Errorf("Unexpected response message: got %v, want %v", response["message"], "Succesfully deleted branch from database!")
		}
	})

	// Test 2: Invalid Swift Code
	t.Run("Invalid Swift Code", func(t *testing.T) {
		// Assuming "INVALIDCODE" does not exist in the database
		req, _ := http.NewRequest(http.MethodDelete, "/v1/swift-codes/INVALIDCODE", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusBadRequest)
		}

		var errorResponse map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Could not decode error response: %v", err)
		}

		if errorResponse["message"] != "Failed to delete swift INVALIDCODE from database" {
			t.Errorf("Unexpected error message: got %v, want %v", errorResponse["message"], "Failed to delete swift INVALIDCODE from database")
		}
	})

	// Test 3: Database Error
	t.Run("Database Error", func(t *testing.T) {
		// Simulate database error by closing the connection
		db.Close()
		req, _ := http.NewRequest(http.MethodDelete, "/v1/swift-codes/ABCABCABCAB", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Unexpected status code: got %v, want %v", resp.Code, http.StatusBadRequest)
		}

		var errorResponse map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Could not decode error response: %v", err)
		}

		if _, exists := errorResponse["message"]; !exists {
			t.Errorf("Error message missing in response")
		}
	})
}

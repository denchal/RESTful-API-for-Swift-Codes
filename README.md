# SWIFT Code Management API

## Overview
This project is a RESTful API built using Go and the Gin framework for managing SWIFT codes of bank branches. The API allows users to retrieve, add, and delete SWIFT codes associated with different bank branches and headquarters in various countries.

## Features
- Retrieve bank details using a SWIFT code
- Get all bank branches within a country
- Add new bank branches
- Delete existing bank branches

## Technologies Used
- **Go (Golang)** - Main programming language
- **Gin** - HTTP web framework
- **MySQL** - Database for storing bank and country information

## API Endpoints

### Retrieve Branch by SWIFT Code
**GET** `/v1/swift-codes/:swift-code`

#### Response (Branch Example)
```json
{
  "address": "123 Bank St, NY, USA",
  "bankName": "Bank of America",
  "countryISO2": "US",
  "countryName": "United States",
  "isHeadquarter": false,
  "swiftCode": "BOFAUS3N123"
}
```
#### Response (Headquarter Example)
```json
{
  "address": "123 Bank St, NY, USA",
  "bankName": "Bank of America",
  "countryISO2": "US",
  "countryName": "United States",
  "isHeadquarter": true,
  "swiftCode": "BOFAUS3NXXX",
  "branches": [
    {
      "address": "456 Branch St, CA, USA",
      "bankName": "Bank of America",
      "countryISO2": "US",
      "countryName": "United States",
      "isHeadquarter": false,
      "swiftCode": "BOFAUS3N123"
    }
  ]
}
```

### Retrieve Branches by Country
**GET** `/v1/swift-codes/country/:countryISO2code`

#### Response
```json
{
  "countryISO2": "US",
  "countryName": "United States",
  "swiftCodes": [
    {
      "address": "123 Bank St, NY, USA",
      "bankName": "Bank of America",
      "countryISO2": "US",
      "isHeadquarter": true,
      "swiftCode": "BOFAUS3NXXX"
    }
  ]
}
```

### Add a New Branch
**POST** `/v1/swift-codes/`

#### Request Body
```json
{
  "address": "789 New Branch St, TX, USA",
  "bankName": "Bank of America",
  "countryISO2": "US",
  "countryName": "United States",
  "isHeadquarter": false,
  "swiftCode": "BOFAUS3N789"
}
```

#### Response
```json
{
  "message": "Successfully added branch to database!"
}
```

### Delete a Branch
**DELETE** `/v1/swift-codes/:swift-code`

#### Response
```json
{
  "message": "Successfully deleted branch from database!"
}
```

## Installation & Setup
1. **Clone the repository:**
   ```sh
   git clone https://github.com/denchal/RESTful-API-for-Swift-Codes.git
   cd RESTful-API-for-Swift-Codes
   ```
2. **Set up the database:**
   - Ensure MySQL is installed and running.
   - Create a database and required tables. (You can use ```initdb/init.sql```)
   - Configure database credentials, and other environmental variables, for example (powershell code):
     ```ps
     $env:DB_HOST = "0.0.0.0"
     $env:DB_PORT = "3306" 
     $env:DB_USER = "root"
     $env:DB_PASSWORD = "1234"
     $env:DB_NAME = "swift_db" 
     ```
3. **Install dependencies:**
   ```sh
   go mod tidy
   ```
4. **Run the API:**
   ```sh
   go run main.go
   ```
## Alternative Setup using Docker:
**Run using Docker Compose:**
   - Ensure Docker and Docker Compose are installed.
   - Edit ```Dockerfile``` and ```docker-compose.yaml``` to match database needs.
   - Start the application using:
     ```sh
     docker-compose up --build
     ```
     
**Warning: First run takes longer due to parsing of the entire CSV file, as I did not include all data inside the init.sql file!** <br>
**App is hardcoded to run on ```localhost:8080```!**

## Database Schema
### `branches` Table
| Column           | Type      | Description |
|----------------|----------|-------------|
| `swift_code`    | VARCHAR  | Unique SWIFT code |
| `name`          | VARCHAR  | Bank name |
| `town_name`     | VARCHAR  | Town name |
| `address`       | VARCHAR  | Branch address |
| `time_zone`     | VARCHAR  | Bank time zone |
| `country_iso2`  | VARCHAR  | Country ISO2 code |
| `is_headquarter`| BOOLEAN  | True if headquarter |

### `countries` Table
| Column           | Type      | Description |
|----------------|----------|-------------|
| `country_iso2`  | VARCHAR  | Country ISO2 code |
| `country_name`  | VARCHAR  | Country name |

## Error Handling
All API responses return appropriate HTTP status codes:
- `200 OK` - Success
- `400 Bad Request` - Invalid input or database constraints
- `404 Not Found` - Resource not found

## Testing
To ensure everything is running correctly, head to cloned repository and run:
```sh
go test ./...
```

## License
This project is licensed under the MIT License.


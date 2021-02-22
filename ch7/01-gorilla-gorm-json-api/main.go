package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

// Shipment GORM model (table)
type Shipment struct {
	gorm.Model
	Packages []Package
	Data     string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB" json:"-"`
}

// Package GORM model (table)
type Package struct {
	gorm.Model
	Data string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
}

// TableName method returns singular name (GORM creates tables with plural names)
func (Shipment) TableName() string {
	return "Shipment"
}

// TableName method returns singular name (GORM creates tables with plural names)
func (Package) TableName() string {
	return "Package"
}

// PackageResponse is the response to be send back for Package
type PackageResponse struct {
	Package Package `json:"Package"`
}

// DBClient stores the database session imformation. Needs to be initialized once
type DBClient struct {
	db *gorm.DB
}

// PostPackage saves a package
func (driver *DBClient) PostPackage(w http.ResponseWriter, r *http.Request) {
	var Package = Package{}
	postBody, _ := ioutil.ReadAll(r.Body)
	Package.Data = string(postBody)
	driver.db.Save(&Package)
	responseMap := map[string]interface{}{"id": Package.ID}
	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(responseMap)
	w.Write(response)
}

// GetPackage fetches a package
func (driver *DBClient) GetPackage(w http.ResponseWriter, r *http.Request) {
	var Package = Package{}
	vars := mux.Vars(r)
	driver.db.First(&Package, vars["id"])
	var PackageData interface{}
	json.Unmarshal([]byte(Package.Data), &PackageData)
	var response = PackageResponse{Package: Package}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	respJSON, _ := json.Marshal(response)
	w.Write(respJSON)
}

// GetPackagesbyWeight fetches all packages with given weight
func (driver *DBClient) GetPackagesbyWeight(w http.ResponseWriter, r *http.Request) {
	var packages []Package
	weight := r.FormValue("weight")
	var query = "SELECT * FROM \"Package\" WHERE data->>'weight'=?"
	driver.db.Raw(query, weight).Scan(&packages)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	respJSON, _ := json.Marshal(packages)
	w.Write(respJSON)
}

// InitDB initializes database
func InitDB() (*gorm.DB, error) {
	var connectionString = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, host, dbname)
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Shipment{}, &Package{})
	return db, nil
}

// entrypoint
func main() {
	db, err := InitDB()
	if err != nil {
		panic(err)
	}
	dbclient := &DBClient{db: db}
	if err != nil {
		panic(err)
	}
	defer db.Close()
	r := mux.NewRouter()
	r.HandleFunc("/v1/package/{id:[a-zA-Z0-9]*}", dbclient.GetPackage).Methods("GET")
	r.HandleFunc("/v1/package", dbclient.PostPackage).Methods("POST")
	r.HandleFunc("/v1/package", dbclient.GetPackagesbyWeight).Methods("GET")
	srv := &http.Server{
		Handler:      r,
		Addr:         "localhost:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

/*
postgres-# \dt+
                               List of relations
 Schema |   Name   | Type  |  Owner   | Persistence |    Size    | Description
--------+----------+-------+----------+-------------+------------+-------------
 public | Package  | table | postgres | permanent   | 8192 bytes |
 public | Shipment | table | postgres | permanent   | 8192 bytes |
(2 rows)

postgres-# \d+ "Shipment"
                                                            Table "public.Shipment"
   Column   |           Type           | Collation | Nullable |                Default                 | Storage  | Stats target | Description
------------+--------------------------+-----------+----------+----------------------------------------+----------+--------------+-------------
 id         | integer                  |           | not null | nextval('"Shipment_id_seq"'::regclass) | plain    |              |
 created_at | timestamp with time zone |           |          |                                        | plain    |              |
 updated_at | timestamp with time zone |           |          |                                        | plain    |              |
 deleted_at | timestamp with time zone |           |          |                                        | plain    |              |
 data       | jsonb                    |           | not null | '{}'::jsonb                            | extended |              |
Indexes:
    "Shipment_pkey" PRIMARY KEY, btree (id)
    "idx_shipment_deleted_at" btree (deleted_at)
Access method: heap

postgres-# \d+ "Package"
                                                            Table "public.Package"
   Column   |           Type           | Collation | Nullable |                Default                | Storage  | Stats target | Description
------------+--------------------------+-----------+----------+---------------------------------------+----------+--------------+-------------
 id         | integer                  |           | not null | nextval('"Package_id_seq"'::regclass) | plain    |              |
 created_at | timestamp with time zone |           |          |                                       | plain    |              |
 updated_at | timestamp with time zone |           |          |                                       | plain    |              |
 deleted_at | timestamp with time zone |           |          |                                       | plain    |              |
 data       | jsonb                    |           | not null | '{}'::jsonb                           | extended |              |
Indexes:
    "Package_pkey" PRIMARY KEY, btree (id)
    "idx_package_deleted_at" btree (deleted_at)
Access method: heap

$ curl -i -w'\n' localhost:8080/v1/package -H 'cache-control: no-cache' -H 'content-type: application/json'   -d '{ "dimensions": { "width": 21, "height": 12 },
        "weight": 10,
        "is_damaged": false,
        "status": "In transit" }'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Mon, 22 Feb 2021 15:56:14 GMT
Content-Length: 8

{"id":1}

$ curl -s localhost:8080/v1/package/1 | jq .
{
  "Package": {
    "ID": 1,
    "CreatedAt": "2021-02-22T15:56:14.66136Z",
    "UpdatedAt": "2021-02-22T15:56:14.66136Z",
    "DeletedAt": null,
    "Data": "{\"status\": \"In transit\", \"weight\": 10, \"dimensions\": {\"width\": 21, \"height\": 12}, \"is_damaged\": false}"
  }
}

$ curl -s localhost:8080/v1/package?weight=10 | jq
[
  {
    "ID": 1,
    "CreatedAt": "2021-02-22T15:56:14.66136Z",
    "UpdatedAt": "2021-02-22T15:56:14.66136Z",
    "DeletedAt": null,
    "Data": "{\"status\": \"In transit\", \"weight\": 10, \"dimensions\": {\"width\": 21, \"height\": 12}, \"is_damaged\": false}"
  }
]

postgres=# SELECT * FROM "Package";
 id |          created_at          |          updated_at          | deleted_at |                                                  data
----+------------------------------+------------------------------+------------+--------------------------------------------------------------------------------------------------------
  1 | 2021-02-22 15:56:14.66136+00 | 2021-02-22 15:56:14.66136+00 |            | {"status": "In transit", "weight": 10, "dimensions": {"width": 21, "height": 12}, "is_damaged": false}
(1 row)
*/

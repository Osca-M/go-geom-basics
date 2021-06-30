package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	_ "github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/twpayne/go-geom/encoding/geojson"
	"go-geom-basics/facilities"
	"log"
	"net/http"
)

const (
	host     = "localhost"
	port     = 6432
	user     = "postgres"
	password = "postgres"
	dbname   = "go-geom"
)

var psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

// allFacilities return a geojson of all medical facilities
//goland:noinspection SqlNoDataSourceInspection
func allFacilities(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {

		}
	}(db)
	rows, err := db.Query(`
		SELECT ST_AsEWKB(geom), country, city, cap_beds, emergency, ref_date, house_number, pub_date, street, tel,
		       ref_id, facility_type, list_specs, email, hospital_name, cc, public_private, comments, postcode, url,
		       site_name, geo_qual
		FROM public.medical_facilities;
`)
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)
	err = json.NewEncoder(w).Encode(convertRowsToGeoJSON(rows))
	if err != nil {
		log.Fatal(err)
	}
}

// convertRowsToGeoJSON Generic function that converts rows from sql queries to a GeoJSON FeatureCollection
func convertRowsToGeoJSON(r *sql.Rows) map[string]interface{} {
	defer func(r *sql.Rows) {
		err := r.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(r)
	var medFacilities []*geojson.Feature
	for r.Next() {
		var ewkbPoint ewkb.Point
		f := new(facilities.MedicalFacility)
		err := r.Scan(&ewkbPoint, &f.Country, &f.City, &f.CapBeds, &f.Emergency, &f.RefDate, &f.HouseNumber,
			&f.PubDate, &f.Street, &f.Tel, &f.RefID, &f.FacilityType, &f.ListSpecs, &f.Email, &f.HospitalName, &f.Cc,
			&f.PublicPrivate, &f.Comments, &f.Postcode, &f.URL, &f.SiteName, &f.GeoQual)
		if err != nil {
			log.Fatal(err)
		}
		medFacilities = append(
			medFacilities, &geojson.Feature{
				Geometry: ewkbPoint.Point,
				Properties: map[string]interface{}{
					"Country":       &f.Country,
					"City":          &f.City,
					"CapBeds":       &f.CapBeds,
					"Emergency":     &f.Emergency,
					"RefDate":       &f.RefDate,
					"HouseNumber":   &f.HouseNumber,
					"PubDate":       &f.PubDate,
					"Street":        &f.Street,
					"Tel":           &f.Tel,
					"RefID":         &f.RefID,
					"FacilityType":  &f.FacilityType,
					"ListSpecs":     &f.ListSpecs,
					"Email":         &f.Email,
					"HospitalName":  &f.HospitalName,
					"Cc":            &f.Cc,
					"PublicPrivate": &f.PublicPrivate,
					"Comments":      &f.Comments,
					"Postcode":      &f.Postcode,
					"URL":           &f.URL,
					"SiteName":      &f.SiteName,
				},
			},
		)
	}
	if err := r.Err(); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	fc := geojson.FeatureCollection{Features: medFacilities}
	bits, err := fc.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	var dat map[string]interface{}
	err = json.Unmarshal(bits, &dat)
	if err != nil {
		log.Fatal(err)
	}
	return dat

}

// handleRequests handles all our http requests and routes them using gorilla/mux package
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/all-facilities", allFacilities).Methods("GET")
	log.Fatal(http.ListenAndServe(":5000", myRouter))
}
func main() {
	fmt.Println("Program Initialization")
	sqlDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {

		}
	}(sqlDB)

	err = facilities.CreateDB(sqlDB)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		err := facilities.AddMedicalFacilities(sqlDB)
		if err != nil {
			log.Fatal(err)
		}
	}()
	handleRequests()
	fmt.Println("***...Exiting....***")
}

package facilities

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkbhex"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// MedicalFacility a data mapping for medical facilities from our API
type MedicalFacility struct {
	gorm.Model
	Geometry      json.RawMessage `json:"geometry"`
	Country       string          `json:"country"`
	City          string          `json:"city"`
	CapBeds       sql.NullString          `json:"cap_beds,omitempty"`
	Emergency     string          `json:"emergency"`
	RefDate       string          `json:"ref_date"`
	HouseNumber   string          `json:"house_number"`
	PubDate       string          `json:"pub_date"`
	Street        string          `json:"street"`
	Tel           string          `json:"tel"`
	RefID         string          `json:"ref_id"`
	FacilityType  string          `json:"facility_type"`
	ListSpecs     string          `json:"list_specs"`
	Email         string          `json:"email"`
	HospitalName  string          `json:"hospital_name"`
	Cc            string          `json:"cc"`
	PublicPrivate string          `json:"public_private"`
	Comments      string          `json:"comments"`
	Postcode      string          `json:"postcode"`
	URL           string          `json:"url"`
	SiteName      string          `json:"site_name"`
	GeoQual       string          `json:"geo_qual"`
}

// CreateDB connects to our database, creates the medical_facilities table that can store MedicalFacility
//goland:noinspection SqlNoDataSourceInspection
func CreateDB(db *sql.DB) error {
	fmt.Println("creating table")
	_, err := db.Exec(`
			CREATE EXTENSION IF NOT EXISTS postgis;
			DROP TABLE IF EXISTS medical_facilities;
			CREATE TABLE IF NOT EXISTS medical_facilities (
			    id SERIAL PRIMARY KEY,
			    geom geometry(POINT, 4326) NOT NULL,
			    country TEXT NOT NULL,
			    city TEXT NOT NULL,
			    cap_beds TEXT,
			    emergency TEXT NOT NULL,
			    ref_date TEXT NOT NULL,
			    house_number TEXT NOT NULL,
			    pub_date TEXT NOT NULL,
			    street TEXT NOT NULL,
			    tel TEXT NOT NULL,
			    ref_id TEXT NOT NULL,
			    facility_type TEXT NOT NULL,
			    list_specs TEXT NOT NULL,
			    email TEXT NOT NULL,
			    hospital_name TEXT NOT NULL,
			    cc TEXT NOT NULL,
			    public_private TEXT NOT NULL,
			    comments TEXT NOT NULL,
			    postcode TEXT NOT NULL,
			    url TEXT NOT NULL,
			    site_name TEXT NOT NULL,
			    geo_qual TEXT NOT NULL
			);
		`)
	fmt.Println("created db")
	return err
}

//AddMedicalFacilities reads file or fetches data from API and inserts it into a database table
func AddMedicalFacilities(db *sql.DB) error {
	//filename, err := os.Open("Europe_Medical_Facilities-1623932428315.json")
	//defer func(filename *os.File) {
	//	err := filename.Close()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}(filename)
	//data, err := ioutil.ReadAll(filename)

	resp, err := http.Get("https://gisco-services.ec.europa.eu/pub/healthcare/geojson/all.geojson")
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {

		}
	}(tx)
	stmt, err := tx.Prepare(pq.CopyIn("medical_facilities", "geom", "country", "city", "cap_beds",
		"emergency", "ref_date", "house_number", "pub_date", "street", "tel", "ref_id", "facility_type", "list_specs",
		"email", "hospital_name", "cc", "public_private", "comments", "postcode", "url", "site_name", "geo_qual"))
	if err != nil {
		return err
	}

	var i map[string]interface{}
	err = json.Unmarshal(data, &i)
	if err != nil {
		log.Fatal(err)
	}
	features := i["features"].([]interface{})

	for i := range features {
		Properties := features[i].(map[string]interface{})["properties"]
		country := Properties.(map[string]interface{})["country"].(string)
		city := Properties.(map[string]interface{})["city"].(string)
		capBeds := Properties.(map[string]interface{})["cap_beds"]
		emergency := Properties.(map[string]interface{})["emergency"].(string)
		refDate := Properties.(map[string]interface{})["ref_date"].(string)
		houseNumber := Properties.(map[string]interface{})["house_number"].(string)
		pubDate := Properties.(map[string]interface{})["pub_date"].(string)
		street := Properties.(map[string]interface{})["street"].(string)
		tel := Properties.(map[string]interface{})["tel"].(string)
		refID := Properties.(map[string]interface{})["id"].(string)
		facilityType := Properties.(map[string]interface{})["facility_type"].(string)
		listSpecs := Properties.(map[string]interface{})["list_specs"].(string)
		email := Properties.(map[string]interface{})["email"].(string)
		hospitalName := Properties.(map[string]interface{})["hospital_name"].(string)
		cc := Properties.(map[string]interface{})["cc"].(string)
		publicPrivate := Properties.(map[string]interface{})["public_private"].(string)
		comments := Properties.(map[string]interface{})["comments"].(string)
		postcode := Properties.(map[string]interface{})["postcode"].(string)
		url := Properties.(map[string]interface{})["url"].(string)
		siteName := Properties.(map[string]interface{})["site_name"].(string)
		geoQual := Properties.(map[string]interface{})["geo_qual"].(string)
		Lat := Properties.(map[string]interface{})["lat"].(float64)
		Lng := Properties.(map[string]interface{})["lon"].(float64)
		point := geom.NewPoint(geom.XY).MustSetCoords([]float64{Lng, Lat}).SetSRID(4326)
		ewkbHexGeom, err := ewkbhex.Encode(point, ewkbhex.NDR)

		if err != nil {
			return err
		}
		if _, err := stmt.Exec(ewkbHexGeom, country, city, capBeds, emergency, refDate, houseNumber, pubDate, street, tel,
			refID, facilityType, listSpecs, email, hospitalName, cc, publicPrivate, comments, postcode, url, siteName,
			geoQual,
		); err != nil {
			return err
		}

		fmt.Printf("Inserted %s medical facility\n", refID)
	}
	if _, err := stmt.Exec(); err != nil {
		return err
	}
	fmt.Println("Done loading data")
	return tx.Commit()
}

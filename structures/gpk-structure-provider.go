package structures

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

//NsiProperties is a reflection of the JSON feature property attributes from the NSI-API
type NsiProperties struct {
	Name      string  `json:"fd_id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Occtype   string  `json:"occtype"`
	FoundHt   float64 `json:"found_ht"`
	FoundType string  `json:"found_type"`
	DamCat    string  `json:"st_damcat"`
	StructVal float64 `json:"val_struct"`
	ContVal   float64 `json:"val_cont"`
	CB        string  `json:"cbfips"`
	Pop2amu65 int32   `json:"pop2amu65"`
	Pop2amo65 int32   `json:"pop2amo65"`
	Pop2pmu65 int32   `json:"pop2pmu65"`
	Pop2pmo65 int32   `json:"pop2pmo65"`
}

//NsiFeature is a feature which contains the properties of a structure from the NSI API
type NsiFeature struct {
	Properties NsiProperties `json:"properties"`
}

//NsiInventory is a slice of NsiFeature that describes a complete json feature array return or feature collection return
type NsiInventory struct {
	Features []NsiFeature
}

//SQLDataSet is a simple struct to store a sql dataset
type SQLDataSet struct {
	db *sql.DB
}

//OpenSQLNSIDataSet opens a sqldataset with the NSI data
func OpenSQLNSIDataSet(nsiLoc string) SQLDataSet {
	db, _ := sql.Open("sqlite3", nsiLoc)
	db.SetMaxOpenConns(1)
	return SQLDataSet{db: db}
}

var nsiLoc string = "./nsiv2_29.gpkg?cache=shared&mode=rwc" //this targets the location of the NSI - maybe get some way to prompt the user for this... cache=shared comes from https://github.com/mattn/go-sqlite3/issues/274 but unsure if it does anything

//GetByFips returns an NsiInventory for a FIPS code
func GetByFips(fips string) NsiInventory {
	nsi := OpenSQLNSIDataSet(nsiLoc)
	rows, err1 := nsi.db.Query("SELECT fd_id, x, y, cbfips, occtype, found_ht, found_type, st_damcat, val_struct, val_cont, pop2amu65, pop2amo65, pop2pmu65, pop2pmo65 FROM nsi WHERE cbfips LIKE '" + fips + "'%")
	if err1 != nil {
		log.Fatal(err1)
	}
	defer rows.Close()

	var inventory NsiInventory
	for rows.Next() { // Iterate and fetch the records from result cursor
		feature := NsiFeature{}
		err2 := rows.Scan(&feature.Properties.Name, &feature.Properties.X, &feature.Properties.Y, &feature.Properties.CB, &feature.Properties.Occtype, &feature.Properties.FoundHt, &feature.Properties.FoundType, &feature.Properties.DamCat, &feature.Properties.StructVal, &feature.Properties.ContVal, &feature.Properties.Pop2amu65, &feature.Properties.Pop2amo65, &feature.Properties.Pop2pmu65, &feature.Properties.Pop2pmo65)
		if err2 != nil {
			panic(err2)
		}
		inventory.Features = append(inventory.Features, feature)
	}
	return inventory
}

//GetByBbox returns an NsiInventory for a Bounding Box
func GetByBbox(bbox string) NsiInventory {
	url := fmt.Sprintf("%s?bbox=%s&fmt=fa", apiURL, bbox)
	return nsiAPI(url)
}

//NsiStreamProcessor is a function used to process an in memory NsiFeature through the NsiStreaming service endpoints
type NsiStreamProcessor func(str NsiFeature)

/*
memory effecient structure compute methods
*/

//GetByFipsStream a streaming service for NsiFeature based on a FIPs code
func GetByFipsStream(fips string, nsp NsiStreamProcessor) error {
	// if we are not behind the USACE Firewall, we access a local NSI database
	nsi := OpenSQLNSIDataSet(nsiLoc)
	rows, err1 := nsi.db.Query("SELECT fd_id, x, y, cbfips, occtype, found_ht, found_type, st_damcat, val_struct, val_cont, pop2amu65, pop2amo65, pop2pmu65, pop2pmo65 FROM nsi WHERE cbfips LIKE '" + fips + "%'")

	if err1 != nil {
		log.Fatal(err1)
	}
	defer rows.Close()

	for rows.Next() { // Iterate and fetch the records from result cursor
		feature := NsiFeature{}
		err2 := rows.Scan(&feature.Properties.Name, &feature.Properties.X, &feature.Properties.Y, &feature.Properties.CB, &feature.Properties.Occtype, &feature.Properties.FoundHt, &feature.Properties.FoundType, &feature.Properties.DamCat, &feature.Properties.StructVal, &feature.Properties.ContVal, &feature.Properties.Pop2amu65, &feature.Properties.Pop2amo65, &feature.Properties.Pop2pmu65, &feature.Properties.Pop2pmo65)
		if err2 != nil {
			panic(err2)
		}
		nsp(feature)
	}
	return nil
}

//GetByBboxStream a streaming service for NsiFeature based on a bounding box
func GetByBboxStream(bbox string, nsp NsiStreamProcessor) error {
	url := fmt.Sprintf("%s?bbox=%s&fmt=fs", apiURL, bbox)
	return nsiAPIStream(url, nsp)
}

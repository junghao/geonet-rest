package main

// Loads lookup information. This is for things that don't change very often like regions etc.

import (
	"database/sql"
	"log"
)

var (
	qrV1GeoJSON []byte
	quakeRegion map[string]int
	allRegion   map[string][]byte
)

func init() {
	// Connect to the DB for this init func.
	// Due to the defered close this will still need to happen again in main when the app starts.
	var err error
	db, err = sql.Open("postgres", "connect_timeout=1 user="+config.DataBase.User+" password="+config.DataBase.Password+" dbname=hazard sslmode=disable")
	if err != nil {
		log.Println("Problem with DB config.")
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()

	if err != nil {
		log.Println("Problem pinging DB - is it up and contactable.")
		log.Fatal(err)
	}

	// quake regions
	var reg string
	quakeRegion = make(map[string]int)

	rows, err := db.Query("select regionname FROM qrt.region where groupname in ('region', 'north', 'south')")
	if err != nil {
		log.Println("Problem loading quake region query lookups.")
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&reg)
		if err != nil {
			log.Println("Problem loading quake region query lookups.")
			log.Fatal(err)
		}
		quakeRegion[reg] = 1
	}
	err = rows.Err()
	if err != nil {
		log.Println("Problem loading quake region query lookups.")
		log.Fatal(err)
	}
	rows.Close()

	// all regions (quake and volcano)
	allRegion = make(map[string][]byte)

	rows, err = db.Query("select regionname FROM qrt.region")
	if err != nil {
		log.Println("Problem loading region query lookups.")
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&reg)
		if err != nil {
			log.Println("Problem loading region query lookups.")
			log.Fatal(err)
		}
		d, err := regionV1GJ(reg)
		if err != nil {
			log.Println("Problem loading region query lookups.")
			log.Fatal(err)
		}
		allRegion[reg] = d
	}
	err = rows.Err()
	if err != nil {
		log.Println("Problem loading region query lookups.")
		log.Fatal(err)
	}
	rows.Close()

	qrV1GeoJSON, err = quakeRegionsV1GJ()
	if err != nil {
		log.Println("Problem loading quake region geojson.")
		log.Fatal(err)
	}
}

package main

import (
	//"log"

	_ "github.com/mattn/go-sqlite3"
)

func insert(secure bool, domain string, path string, fromID int) (int, int){

	var domainID int = -1
	db.QueryRow(`
		SELECT id FROM domains WHERE domain = ?
	`, domain).Scan(&domainID)

	if domainID == -1 {

		res, err := db.Exec(`
			INSERT INTO domains (domain) VALUES (?)
		`, domain)
		if err != nil {
			//log.Fatal(err)
		}
	
		id, err := res.LastInsertId()
		if err != nil {
			//log.Fatal(err)
		}

		domainID = int(id)
	}

	var pathID int = -1
	db.QueryRow(`
		SELECT id FROM paths WHERE domain = ? AND path = ?
		`, domainID, path).Scan(&pathID)

	if pathID == -1 {

		res, err := db.Exec(`
			INSERT INTO paths (domain, path, secure, scanned, onHold) VALUES (?, ?, ?, 0, 0)
		`, domainID, path, secure)
		if err != nil {
			//log.Fatal(err)
			
		}

		id, err := res.LastInsertId()
		if err != nil {
			//log.Fatal(err)
			
		}

		pathID = int(id)

		if fromID != -1 {
			_, err := db.Exec(`
				INSERT INTO links (parent, child) VALUES (?, ?)
			`, fromID, pathID)
			if err != nil {
				//log.Fatal(err)
			
			}
		}

	}


	return domainID, pathID
}

func markScanned(pathID int) {
	_, err := db.Exec(`
		UPDATE paths SET scanned = 1, onHold = 0 WHERE id = ?
	`, pathID)
	if err != nil {
		//log.Fatal(err)
		
	}
}

func next() (string, error) {

	var domain, path string
	var secure bool

	err := db.QueryRow(`
		SELECT domain, path, secure FROM paths WHERE scanned = 0 and onHold = 0 ORDER BY RANDOM() LIMIT 1
	`).Scan(&domain, &path, &secure)

	if err != nil {
		//log.Fatal(err)
		return "", err
	}

	err = db.QueryRow(`
		SELECT domain FROM domains WHERE id = ? LIMIT 1
	`, domain).Scan(&domain)

	if err != nil {
		//log.Fatal(err)
		return "", err
	}

	_, err = db.Exec(`
		UPDATE paths SET onHold = 1 WHERE id = ?
	`, path)
	if err != nil {
		//log.Fatal(err)
		return "", err
	}

	url := "http"

	if secure {
		url += "s"
	}

	url += "://" + domain + path

	return url, nil
}

func stats() (int, int, int) {
	var total, scanned, sites int
	db.QueryRow(`
		SELECT COUNT(*) FROM paths
	`).Scan(&total)

	db.QueryRow(`
		SELECT COUNT(*) FROM paths WHERE scanned = 1
	`).Scan(&scanned)

	db.QueryRow(`
		SELECT COUNT(*) FROM domains
	`).Scan(&sites)

	return total, scanned, sites
}
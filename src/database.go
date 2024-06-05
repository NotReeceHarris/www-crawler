package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS domains (
            id INTEGER PRIMARY KEY, 
            domain TEXT UNIQUE
        )
    `); err != nil {
		return nil, err
	}

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS paths (
			id INTEGER PRIMARY KEY, 
			domain INTEGER, 
			path TEXT,
			secure BOOLEAN,
			httpCode TEXT,
			scanned BOOLEAN,
			onHold BOOLEAN,
			FOREIGN KEY(domain) REFERENCES domains(id),
			UNIQUE(domain, path)
		)
    `); err != nil {
		return nil, err
	}

	if _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS links (
            id INTEGER PRIMARY KEY,  
            parent INTEGER,
			child INTEGER,
            FOREIGN KEY(parent) REFERENCES paths(id)
			FOREIGN KEY(child) REFERENCES paths(id)
        )
    `); err != nil {
		return nil, err
	}

	var domains int

	db.QueryRow(`
		SELECT COUNT(*) from domains
	`).Scan(&domains)

	if (domains == 0) {
		insert(true, "www.codegalaxy.co.uk", "/", -1)
	}

	db.Exec(`
		UPDATE paths SET onHold = 0 WHERE onHold = 1
	`)

	return db, nil

}

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
			return -1, -1
		}
	
		id, err := res.LastInsertId()
		if err != nil {
			return -1, -1
		}

		domainID = int(id)
	}

	var pathID int = -1
	db.QueryRow(`
		SELECT id FROM paths WHERE domain = ? AND path = ?
		`, domainID, path).Scan(&pathID)

	if pathID == -1 {

		res, err := db.Exec(`
			INSERT INTO paths (domain, path, secure, scanned, onHold, httpCode) VALUES (?, ?, ?, 0, 0, "")
		`, domainID, path, secure)
		if err != nil {
			return -1, -1
		}

		id, err := res.LastInsertId()
		if err != nil {
			return -1, -1
		}

		pathID = int(id)

		if fromID != -1 {
			_, err := db.Exec(`
				INSERT INTO links (parent, child) VALUES (?, ?)
			`, fromID, pathID)
			if err != nil {
				return -1, -1
			}
		}

	}


	return domainID, pathID
}

func markScanned(pathID int, httpCode int) {
	_, err := db.Exec(`
		UPDATE paths SET scanned = 1, onHold = 0, httpCode = ? WHERE id = ?
	`, httpCode, pathID)
	if err != nil {
		//log.Fatal(err)
		
	}
}

var picks int = 0
var approach int = 0

func next() (string, error) {

	/* 
		Hybrid, every x picks switch approach start with randomly picking from the database and after
		x picks switch to picking the oldest path that has not been scanned, this is to ensure that the 
	*/

	var domain, path string
	var secure bool
	var pathID int

	if approach == 0 {
		err := db.QueryRow(`
			SELECT id, domain, path, secure FROM paths WHERE scanned = 0 and onHold = 0 ORDER BY RANDOM() LIMIT 1
		`).Scan(&pathID, &domain, &path, &secure)

		if err != nil {
			return "", err
		}
	} else if approach == 1 {
		err := db.QueryRow(`
			SELECT id, domain, path, secure FROM paths WHERE scanned = 0 and onHold = 0 ORDER BY id ASC LIMIT 1
		`).Scan(&pathID, &domain, &path, &secure)

		if err != nil {
			return "", err
		}
	} else if approach == 2 {
		err := db.QueryRow(`
			SELECT id, domain, path, secure FROM paths WHERE scanned = 0 and onHold = 0 ORDER BY id DESC LIMIT 1
		`).Scan(&pathID, &domain, &path, &secure)

		if err != nil {
			return "", err
		}
	}

	_, err := db.Exec(`
		UPDATE paths SET onHold = 1 WHERE id = ?
	`, pathID)
	if err != nil {
		return "", nil
	}

	picks++

	if picks <= 100 {
		approach = 1
	} else if picks <= 200 {
		approach = 2
	} else if picks <= 300 {
		approach = 0
		picks = 0
	}

	err = db.QueryRow(`
		SELECT domain FROM domains WHERE id = ? LIMIT 1
	`, domain).Scan(&domain)

	if err != nil {
		return "", nil
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
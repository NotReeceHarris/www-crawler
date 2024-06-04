package main

import (
	"fmt"
)

func cleanup() {
	if db != nil {
		db.Close()
	}
	fmt.Println("Cleaned up and exiting.")
}

func crawl(url string, fromID int) {
	scheme, domain, path, err := parseURL(url)
	if err != nil {
		//fmt.Println(err)
		return
	}

	secure := scheme == "https"
	_, pathID := insert(secure, domain, path, fromID)

	links, err := get(url, pathID)
	if err != nil {
		//fmt.Println(err)
		return
	}

	for _, link := range links {
		scheme, domain, path, err := parseURL(link)
		if err != nil {
			//fmt.Println(err)
			continue
		}
		insert(scheme == "https", domain, path, pathID)
	}
	
}
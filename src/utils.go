package main

import (
	"fmt"
	"strconv"
    "strings"
	"time"
)

func cleanup() {
	if db != nil {
		db.Close()
	}
	fmt.Println("Cleaned up and exiting.")
}

func crawl(url string, fromID int) {
	scheme, domain, path, err := parseURL(url)
	if err == nil {
		secure := scheme == "https"
		_, pathID := insert(secure, domain, path, fromID)

		links, err := get(url, pathID)
		if err == nil {
			for _, link := range links {
				scheme, domain, path, err := parseURL(link)
				if err == nil {
					insert(scheme == "https", domain, path, pathID)
				}
			}
		}
	}
}

func addCommasToNumber(num int64) string {
    in := strconv.FormatInt(num, 10)
    var out strings.Builder
    l := len(in)
    for i, v := range in {
        out.WriteRune(v)
        if (l-i-1)%3 == 0 && i < l-1 {
            out.WriteRune(',')
        }
    }
    return out.String()
}

func ticker() func() {
	app_start := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	first := true

	spider := []string{
		"    /      \\     ",
		" \\  \\  ,,  /  /  ",
		"  '-.`\\()/`.-'	 ",
		" .--_'(  )'_--.  ",
		"/ /` /`\"\"`\\ `\\ \\ ",
		" |  |  \033[31m><\033[0m  |  |  ",
		" \\  \\      /  /  ",
		"     '.__.'      ",
	}

	return func() {
        for range ticker.C {

			total, scanned, sites := stats()
			var totalCrawlsPerMinute float64 = 0
			var average time.Duration =  time.Now().Sub(time.Now())
			var currentApproach string = "Random           "

			if !first {
				fmt.Printf("\033[%dA", 8)
				fmt.Print("\033[2K")

				var sum time.Duration
				for _, duration := range workersCrawlTime {
					sum += duration
				}

				if len(workersCrawlTime) != 0 {
					average = time.Duration(int64(sum) / int64(len(workersCrawlTime)))

					if average != 0 {
						averageInMinutes := float64(average) / float64(time.Minute)
						crawlsPerMinutePerWorker := 1 / averageInMinutes
						totalCrawlsPerMinute = crawlsPerMinutePerWorker * float64(numWorkers)
					}
				}

				if approach == 0 {
					currentApproach = "Random           "
				} else if approach == 1 {
					currentApproach = "Sequential (ASC) "
				} else if approach == 2 {
					currentApproach = "Sequential (DESC)"
				}

			} else {
				fmt.Println("")
			}

			fmt.Println(spider[0], "Pages total         :", addCommasToNumber(int64(total)))
			fmt.Println(spider[1], "Domains total       :", addCommasToNumber(int64(sites)))
			fmt.Println(spider[2], "Pages crawled       :", addCommasToNumber(int64(scanned)))
			fmt.Println(spider[3], "Average crawl time  :", average)
			fmt.Println(spider[4], "Crawls per minute   :", strings.Split(strconv.FormatFloat(totalCrawlsPerMinute, 'f', -1, 64), ".")[0])
			fmt.Println(spider[5], "Approach            :", currentApproach)
			fmt.Println(spider[6], "Workers             :", addCommasToNumber(int64(numWorkers)))
			fmt.Println(spider[7], "Elapsed time        :", time.Now().Sub(app_start))

			first = false

        }
    }
}
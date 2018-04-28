package unpaywall_test

import (
	"fmt"
	"log"

	unpaywall "github.com/jvecsei/go-unpaywall"
)

func ExampleUnpaywall_RequestByDOI() {
	// your email address
	var email string
	// DOI
	var doi string
	u, _ := unpaywall.New(email)
	// Request example
	result, err := u.RequestByDOI(doi)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Search result: %v", result)
}

func ExampleUnpaywall_DownloadByDOI() {
	// your email address
	var email string
	// DOI
	var doi string
	u, _ := unpaywall.New(email)
	var targetPath = "./"

	file, err := u.DownloadByDOI(doi, targetPath)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Success! Downloaded file to %s", file)
}

func ExampleUnpaywall_RequestByDOIs() {
	// your email address
	var email string

	// DOIs
	var dois []string

	u, _ := unpaywall.New(email)
	res, err := u.RequestByDOIs(dois)

	for err != nil || res != nil {
		select {
		case r, ok := <-res:
			if !ok {
				res = nil
				continue
			}
			fmt.Printf("Found: %s \n", r.BestOaLocation.URLForPdf)
		case e, ok := <-err:
			if !ok {
				err = nil
				continue
			}
			fmt.Printf("Error: %s \n", e)
		}
	}
}

func ExampleUnpaywall_DownloadByDOIs() {
	// your email address
	var email string

	// DOIs
	var dois []string

	// target directory
	var target string

	u, _ := unpaywall.New(email)
	res, err := u.DownloadByDOIs(dois, target)

	for err != nil || res != nil {
		select {
		case r, ok := <-res:
			if !ok {
				res = nil
				continue
			}
			fmt.Printf("Downloaded file to: %s \n", r)
		case e, ok := <-err:
			if !ok {
				err = nil
				continue
			}
			fmt.Printf("Error: %s \n", e)
		}
	}
}

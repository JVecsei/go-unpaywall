package main

import (
	"flag"
	"log"

	unpaywall "github.com/JVecsei/go-unpaywall"
)

func main() {
	email := flag.String("email", "", "your email address")
	doi := flag.String("doi", "", "the DOI for the file to download")
	flag.Parse()
	if *email == "" || *doi == "" {
		log.Fatalf("Email and doi cant be empty. Add '-h' to get help.")
	}
	u, _ := unpaywall.New(*email)
	file, err := u.DownloadByDOI(*doi, "./")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Success! Downloaded file to %s", file)
}

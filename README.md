# go-unpaywall

[![GoDoc](https://godoc.org/github.com/JVecsei/go-unpaywall?status.svg)](https://godoc.org/github.com/JVecsei/go-unpaywall) [![Build Status](https://travis-ci.org/JVecsei/go-unpaywall.svg?branch=master)](https://travis-ci.org/JVecsei/go-unpaywall)

Simple unofficial library to send requests to the unpaywall API ( http://unpaywall.org/products/api ) and automatically download documents. 

##### Is Unpaywall legal?

"Yes! We harvest content from **legal** sources including repositories run by universities, governments, and scholarly societies, as well as open content hosted by publishers themselves." ([source](http://unpaywall.org/faq), April 28, 2018)

## Examples

Both Methods are also available for multiple requests / downloads (see [godoc](http://godoc.org/github.com/JVecsei/go-unpaywall)), which will be executed concurrently. By default 5 workers are started simultaneously. 

### Example Request

```go
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
```



### Example Download

```go
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
```



### Example Multi-Request



```go
// your email address
var email string
//DOIs
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
```





## CLI-Tool

The CLI tool in the `cmd` folder can be used like this:

`unpaywall-cmd -doi "your-DOI" -email "your-email-address"`

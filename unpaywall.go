//Package unpaywall provides utitlity functions for the unpaywall.org api
package unpaywall

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const (
	unpaywallAPIURL = "https://api.unpaywall.org/v2/%s?email=%s"
	workersCount    = 5
)

//Unpaywall provides methods to request the unpaywall api
type Unpaywall struct {
	email  string
	client *http.Client
}

//New returns a new unpaywall object to send requests to the API
func New(email string) (*Unpaywall, error) {
	//regex from https://www.w3.org/TR/html5/forms.html#valid-e-mail-address
	regex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !regex.MatchString(email) {
		return nil, errors.New("invalid email address")
	}
	cookieJar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: cookieJar,
	}
	u := &Unpaywall{
		email,
		client,
	}
	return u, nil
}

//SearchResult describes the returned data structure from unpaywall
type SearchResult struct {
	StatusCode     int
	BestOaLocation struct {
		Evidence          string `json:"evidence"`
		HostType          string `json:"host_type"`
		IsBest            bool   `json:"is_best"`
		License           string `json:"license"`
		PmhID             string `json:"pmh_id"`
		Updated           string `json:"updated"`
		URL               string `json:"url"`
		URLForLandingPage string `json:"url_for_landing_page"`
		URLForPdf         string `json:"url_for_pdf"`
		Version           string `json:"version"`
	} `json:"best_oa_location"`
	DataStandard    int    `json:"data_standard"`
	Doi             string `json:"doi"`
	DoiURL          string `json:"doi_url"`
	Genre           string `json:"genre"`
	IsOa            bool   `json:"is_oa"`
	JournalIsInDoaj bool   `json:"journal_is_in_doaj"`
	JournalIsOa     bool   `json:"journal_is_oa"`
	JournalIssns    string `json:"journal_issns"`
	JournalName     string `json:"journal_name"`
	OaLocations     []struct {
		Evidence          string `json:"evidence"`
		HostType          string `json:"host_type"`
		IsBest            bool   `json:"is_best"`
		License           string `json:"license"`
		PmhID             string `json:"pmh_id"`
		Updated           string `json:"updated"`
		URL               string `json:"url"`
		URLForLandingPage string `json:"url_for_landing_page"`
		URLForPdf         string `json:"url_for_pdf"`
		Version           string `json:"version"`
	} `json:"oa_locations"`
	PublishedDate               string        `json:"published_date"`
	Publisher                   string        `json:"publisher"`
	Title                       string        `json:"title"`
	Updated                     string        `json:"updated"`
	XReportedNoncompliantCopies []interface{} `json:"x_reported_noncompliant_copies"`
	Year                        int           `json:"year"`
	ZAuthors                    []struct {
		Family string `json:"family"`
		Given  string `json:"given"`
	} `json:"z_authors"`
}

//RequestByDOI sends a new request to unpaywall with the given DOI
func (u *Unpaywall) RequestByDOI(doi string) (*SearchResult, error) {
	searchResult := new(SearchResult)
	if u == nil {
		return searchResult, errors.New("check for errors while initializing Unpaywall")
	}
	requestURL := fmt.Sprintf(unpaywallAPIURL, doi, u.email)
	res, err := u.client.Get(requestURL)

	if err != nil {
		return searchResult, err
	}
	defer res.Body.Close()

	searchResult.StatusCode = res.StatusCode
	if res.StatusCode != http.StatusOK {
		return searchResult, fmt.Errorf("unsuccessful request %s", res.Status)
	}

	response := json.NewDecoder(res.Body)

	if err != nil {
		return searchResult, err
	}

	err = response.Decode(&searchResult)

	return searchResult, err
}

//RequestByDOIs searches for the papers by DOI and returns all found results / errors
func (u *Unpaywall) RequestByDOIs(dois []string) (<-chan *SearchResult, <-chan error) {
	var wg sync.WaitGroup
	workerChan := make(chan string, len(dois))
	resultChan := make(chan *SearchResult, int(len(dois)/2))
	errorChan := make(chan error, int(len(dois)/2))

	for i := 0; i < workersCount; i++ {
		go func() {
			for doi := range workerChan {
				func(doi string) {
					defer wg.Done()
					res, err := u.RequestByDOI(doi)
					if err == nil {
						resultChan <- res
					} else {
						errorChan <- err
					}
				}(doi)
			}
		}()
	}
	wg.Add(len(dois))
	for _, doi := range dois {
		workerChan <- doi
	}
	close(workerChan)
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()
	return resultChan, errorChan
}

//DownloadByDOI searches for the paper by DOI and downloads it to the target path
//returns the filename / error
func (u *Unpaywall) DownloadByDOI(doi string, targetPath string) (string, error) {
	res, err := u.RequestByDOI(doi)
	if err != nil {
		return "", err
	}
	if res.BestOaLocation.URLForPdf == "" {
		return "", errors.New("could not find valid pdf")
	}
	pdfRes, err := u.client.Get(res.BestOaLocation.URLForPdf)

	if err != nil {
		return "", err
	}

	fileContent, err := ioutil.ReadAll(pdfRes.Body)
	if err != nil {
		return "", err
	}
	rand.Seed(time.Now().Unix())
	targetFilename := generateRandomFilename()

	cleanFilenameRegex := regexp.MustCompile(`\W`)

	if res.Title != "" {
		targetFilename = cleanFilenameRegex.ReplaceAllString(res.Title, "_")
	}

	targetFilePath := fmt.Sprintf(filepath.Dir(targetPath)+"%s%s", string(os.PathSeparator), targetFilename+".pdf")
	// Add random numbers to filename if file already exists
	for fileExists(targetFilePath) {
		targetFilePath = fmt.Sprintf(filepath.Dir(targetPath)+"%s%s", string(os.PathSeparator), targetFilename+generateRandomFilename()+".pdf")
	}

	err = ioutil.WriteFile(targetFilePath, fileContent, 777)

	return targetFilePath, err
}

//DownloadByDOIs searches for the papers by DOI and downloads all found documents to the target path
//returns the filenames / errors
func (u *Unpaywall) DownloadByDOIs(dois []string, targetPath string) (<-chan string, <-chan error) {
	var wg sync.WaitGroup
	workerChan := make(chan string, len(dois))
	resultChan := make(chan string, int(len(dois)/2))
	errorChan := make(chan error, int(len(dois)/2))

	for i := 0; i < workersCount; i++ {
		go func() {
			for doi := range workerChan {
				func(doi string, targetPath string) {
					defer wg.Done()
					res, err := u.DownloadByDOI(doi, targetPath)
					if res != "" {
						resultChan <- res
					}
					if err != nil {
						errorChan <- err
					}
				}(doi, targetPath)
			}
		}()
	}
	wg.Add(len(dois))
	for _, doi := range dois {
		workerChan <- doi
	}
	close(workerChan)
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()
	return resultChan, errorChan
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func generateRandomFilename() string {
	rand.Seed(time.Now().Unix())
	return strconv.Itoa(rand.Intn(100000000) + 100000000)
}

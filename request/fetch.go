package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"unicode/utf8"
)

func Fetch(url string, retdat bool) *FetchResponse {
	/*
	   Fetch the html document found at url.
	   Return download time and file size data as part of FetchResponse
	*/

	// Set up the http request
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")

	// Start Request Timer
	start := time.Now()

	//Fetch the url
	resp, err := client.Do(req)
	if err != nil {
		// TODO This needs to return a FetchResponse with err as the url (or think of something better)
		fmt.Println("Error: ", err)
		os.Exit(0)
	}

	// End Request Timer
	end := time.Now()

	// Read the html 'body' content from the response object
	body, err := ioutil.ReadAll(resp.Body)
	responseBody := string(body)

	output := FetchResponse{
		Url:     url,
		Body:    responseBody,
		Headers: resp.Header,
		Bytes:   len(responseBody),
		Runes:   utf8.RuneCountInString(responseBody),
		Time:    end.Sub(start),
		Status:  resp.StatusCode,
	}

	if !retdat { // we don't want the document data or headers
		output.Body = ``
		output.Headers = make(map[string][]string)
	}
	//Close the response body and return the output
	resp.Body.Close()
	return &output
}

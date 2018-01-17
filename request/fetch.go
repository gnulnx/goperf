package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"unicode/utf8"
)

func DefineAssetUrl(baseurl string, asseturl string) string {
	if asseturl[0] == '/' {
		asseturl = baseurl + asseturl
	}
	return asseturl
}

func FetchAsset(baseurl string, asseturl string, retdat bool) *FetchResponse {
	asset_url := DefineAssetUrl(baseurl, asseturl)
	return Fetch(asset_url, retdat)
}

func Fetch(url string, retdat bool) *FetchResponse {
	/*
	   Simple method that fetches a url and returns a FetchOutput structure
	   retdata if True then we return the Body and the Headers
	*/

	// Set up the http request
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")

	//Start Timer
	start := time.Now()

	//Fetch the url
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
	//End Timer
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

	if !retdat {
		output.Body = ``
		output.Headers = make(map[string][]string)
	}
	//Close the response body and return the output
	resp.Body.Close()
	return &output
}

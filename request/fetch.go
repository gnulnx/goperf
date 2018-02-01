package request

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"unicode/utf8"
)

func Fetch(input FetchInput) *FetchResponse {
	/*
	   Fetch document at url size and time data.
	   If retdat is true you also return the http.Response.Body
	*/

	url := input.BaseUrl
	retdat := input.Retdat
	cookies := input.Cookies

	// Set up the http request
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", "Chrome/61.0.3163.100 Mobile Safari/537.36")
	//req.Header.Add("cookie", "sessionid_vagrant=5i4bgzvc0vy8xjgf1flfoh89cwsg74hz; csrftoken_vagrant=taZjH9jskTjfbvDDq7OzdtQnTaB72zIk")
	fmt.Println("cookies: ", cookies)
	req.Header.Add("cookie", cookies)

	//Fetch the url and time the request
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		//color.Red(err.Error())
		return &FetchResponse{
			Url:    err.Error(),
			Status: -100,
			Error:  "There was a problem with you request.  Please double check your url",
		}
	}
	defer resp.Body.Close()
	responseTime := time.Now().Sub(start)

	// Read the html 'body' content from the response object
	body, err := ioutil.ReadAll(resp.Body)
	Error := ""
	if err != nil {
		body = []byte("")
		Error = err.Error()
	}
	responseBody := string(body)

	output := FetchResponse{
		Url:     url,
		Body:    responseBody,
		Headers: resp.Header,
		Bytes:   len(responseBody),
		Runes:   utf8.RuneCountInString(responseBody),
		Time:    responseTime,
		Status:  resp.StatusCode,
		Error:   Error,
	}

	if !retdat { // we don't want the document data or headers
		output.Body = ``
		output.Headers = make(map[string][]string)
	}
	//Close the response body and return the output
	return &output
}

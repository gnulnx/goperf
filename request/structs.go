package request

import (
	"net/http"
	"time"
)

type FetchResponse struct {
	url     string
	body    string
	headers map[string][]string
	bytes   int
	runes   int
	time    time.Duration
	status  int
}

type FetchAllResponse struct {
	baseUrl      *FetchResponse
	jsResponses  []FetchResponse
	imgResponses []FetchResponse
	cssResponses []FetchResponse

	body   string
	time   time.Duration
	status int
}

/*
   Structure to hold the input variables to 'go Run(...)' method call
*/
type Input struct {
	Url        string
	Threads    int
	Iterations int
	Output     int
	Index      int // Also the channel number
	Verbose    bool
}

/*
   Structure used to create web request Channel.  This is how we get the results
   back from the 'go Run(...) method call
*/
type Result struct {
	Total     time.Duration
	Average   time.Duration
	Channel   int
	Responses []*http.Response
}

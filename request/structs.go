package request

import (
	"net/http"
	"time"
)

type FetchResponse struct {
	Url     string
	Body    string
	Headers map[string][]string
	Bytes   int
	Runes   int
	Time    time.Duration
	Status  int
}

type FetchAllResponse struct {
	BaseUrl      *FetchResponse
	JSResponses  []FetchResponse
	IMGResponses []FetchResponse
	CSSResponses []FetchResponse

	Body   string
	Time   time.Duration
	Status int
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

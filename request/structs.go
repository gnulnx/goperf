package request

import (
	"net/http"
	"time"
)

type FetchResponse struct {
	Url     string              `json:"url"`
	Body    string              `json:"body"`
	Headers map[string][]string `json:"headers"`
	Bytes   int                 `json:"bytes"`
	Runes   int                 `json:"runes"`
	Time    time.Duration       `json:"time"`
	Status  int                 `json:"status"`
}

type FetchAllResponse struct {
	BaseUrl      *FetchResponse  `json:"baseUrl"`
	JSResponses  []FetchResponse `json:"jsResponses"`
	IMGResponses []FetchResponse `json:"imgResponses"`
	CSSResponses []FetchResponse `json:"cssResponses"`

	Body   string        `json:"body"`
	Time   time.Duration `json:"time"`
	Status int           `json:"status"`
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

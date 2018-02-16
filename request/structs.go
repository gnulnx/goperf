package request

import (
	"net/http"
	"time"
)

type IterateReqResp struct {
	Url         string          `json:"url"`
	Status      []int           `json:"status"`
	RespTimes   []time.Duration `json:"resp_times"`
	NumRequests int             `json:"num_requests"`
	Bytes       int             `json:"bytes"`
}

type IterateReqRespAll struct {
	AvgTotalRespTime       time.Duration    `json:"avgTotalRespTime"`
	AvgTotalLinearRespTime time.Duration    `json:"avgTotalLinearRespTime"`
	BaseURL                IterateReqResp   `json:"baseURL"`
	JSResps                []IterateReqResp `json:"jsResps"`
	CSSResps               []IterateReqResp `json:"cssResps"`
	IMGResps               []IterateReqResp `json:"imgResps"`
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

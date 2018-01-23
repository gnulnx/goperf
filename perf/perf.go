package perf

import (
	"encoding/json"
	"fmt"
	"github.com/gnulnx/color"
	"github.com/gnulnx/goperf/request"
	"strconv"
	"time"
)

type Init struct {
	Url        string
	Threads    int
	Seconds    int
	Iterations int
	Output     int
	Index      int // Also the channel number
	Verbose    bool
	Results    *request.IterateReqRespAll
}

func (input *Init) Basic() request.IterateReqRespAll {
	// Create slice of channels to hold results
	// Fire off anonymous go routine using newly created channel
	chanslice := []chan request.IterateReqRespAll{}
	for i := 0; i < input.Threads; i++ {
		chanslice = append(chanslice, make(chan request.IterateReqRespAll))
		go func(c chan request.IterateReqRespAll) {
			c <- iterateRequest(input.Url, input.Seconds)
		}(chanslice[i])
	}

	// Wait on all the channels
	results := []request.IterateReqRespAll{}
	for _, ch := range chanslice {
		results = append(results, <-ch)
	}

	input.Results = request.Combine(results)
	return *input.Results

}

func (input Init) Json() {
	tmp, _ := json.MarshalIndent(input.Results, "", "  ")
	fmt.Println(string(tmp))
}

func (input Init) Print() {
	results := input.Results
	yellow := color.New(color.FgYellow).SprintfFunc()
	green := color.New(color.FgGreen).SprintfFunc()
	color.Red("Base Url Results")
	fmt.Printf(" - %-30s %-25s\n", "Url:", green(results.BaseUrl.Url))
	fmt.Printf(" - %-30s %-25s\n", "Number of Requests:", green(strconv.Itoa(len(results.BaseUrl.Status))))
	fmt.Printf(" - %-30s %s\n", "Total Bytes:", green(strconv.Itoa(results.BaseUrl.Bytes)))
	fmt.Printf(" - %-30s %s\n", "Avg Page Resp Time:", green(results.AvgTotalRespTime.String()))

	avg, statusResults := procResult(&results.BaseUrl)
	fmt.Printf(" - %-30s %s\n", "Average Time to First Byte:", green(avg))
	fmt.Printf(" - %-30s %s\n", "Status:", green(statusResults))

	color.Red("JS Results")
	for _, resp := range results.JSResps {
		avg, statusResults := procResult(&resp)
		fmt.Printf(" - %-22s %-20s %-10s\n", green(avg), yellow(statusResults), resp.Url)
	}

	color.Red("CSS Results")
	for _, resp := range results.CSSResps {
		avg, statusResults := procResult(&resp)
		fmt.Printf(" - %-22s %-20s %-10s\n", green(avg), yellow(statusResults), resp.Url)
	}

	color.Red("IMG Results")
	for _, resp := range results.IMGResps {
		avg, statusResults := procResult(&resp)
		fmt.Printf(" - %-22s %-20s %-10s\n", green(avg), yellow(statusResults), resp.Url)
	}
}

func procResult(resp *request.IterateReqResp) (string, string) {
	totalTime := time.Duration(0)
	for _, val := range resp.RespTimes {
		totalTime += val
	}
	avg := time.Duration(int64(totalTime) / int64(len(resp.Status))).String()

	statusCodes := map[string][]int{}
	for _, val := range resp.Status {
		status := strconv.Itoa(val)
		statusCodes[status] = append(statusCodes[status], val)
	}

	statusResults := make(map[string]int)
	for key, _ := range statusCodes {
		statusResults[key] = len(statusCodes[key])
	}
	tmp, _ := json.Marshal(statusResults)
	status := string(tmp)
	return avg, status
}

func iterateRequest(url string, sec int) request.IterateReqRespAll {
	/*
		Continuously fetch 'url' for 'sec' second and return the results.
	*/
	start := time.Now()
	maxTime := time.Duration(sec) * time.Second
	elapsedTime := maxTime

	resp := request.IterateReqResp{
		Url: url,
	}
	jsMap := map[string]*request.IterateReqResp{}
	cssMap := map[string]*request.IterateReqResp{}
	imgMap := map[string]*request.IterateReqResp{}

	var totalRespTimes int64 = 0
	var count int64 = 0 // TODO for loop counter instead???
	for {
		//Fetch the url and all the js, css, and img assets
		fetchAllResp := request.FetchAll(url, false)

		// Set base resp properties
		resp.Status = append(resp.Status, fetchAllResp.BaseUrl.Status)
		resp.RespTimes = append(resp.RespTimes, fetchAllResp.BaseUrl.Time)
		resp.Bytes = fetchAllResp.TotalBytes

		totalRespTimes += int64(fetchAllResp.TotalLinearTime)

		gatherAllStats(fetchAllResp, jsMap, cssMap, imgMap)

		elapsedTime = time.Now().Sub(start)
		count += 1
		if elapsedTime > maxTime {
			break
		}
	}

	avgTotalRespTimes := time.Duration(totalRespTimes / count)

	// TODO Clean this up.  Perhaps some benchmark tests
	// to see if its faster as go routines or not
	jsResps := []request.IterateReqResp{}
	for _, val := range jsMap {
		jsResps = append(jsResps, *val)
	}

	cssResps := []request.IterateReqResp{}
	for _, val := range cssMap {
		cssResps = append(cssResps, *val)
	}

	imgResps := []request.IterateReqResp{}
	for _, val := range imgMap {
		imgResps = append(imgResps, *val)
	}

	output := request.IterateReqRespAll{
		BaseUrl:          resp,
		AvgTotalRespTime: avgTotalRespTimes,
		JSResps:          jsResps,
		CSSResps:         cssResps,
		IMGResps:         imgResps,
	}
	return output
}

func gatherAllStats(resp *request.FetchAllResponse, jsMap map[string]*request.IterateReqResp, cssMap map[string]*request.IterateReqResp, imgMap map[string]*request.IterateReqResp) {
	/*
		Gather all the asset stuff.
		NOTE:  You benchmarked this and the 3 go routine method was way slower so you removed the method
		BenchmarkGatherAllStatsGo-8   	  500000	      2764 ns/op
		BenchmarkGatherAllStats-8     	 2000000	       638 ns/op
	*/
	gatherStats(resp.JSResponses, jsMap)
	gatherStats(resp.CSSResponses, cssMap)
	gatherStats(resp.IMGResponses, imgMap)
}

func gatherStats(Resps []request.FetchResponse, respMap map[string]*request.IterateReqResp) {
	// gather all the responses
	for resp := 0; resp < len(Resps); resp++ {
		url2 := Resps[resp].Url
		bytes := Resps[resp].Bytes
		status := Resps[resp].Status
		respTime := Resps[resp].Time
		_, ok := respMap[url2]
		if !ok {
			respMap[url2] = &request.IterateReqResp{
				Url:         url2,
				Bytes:       bytes,
				Status:      []int{status},
				RespTimes:   []time.Duration{respTime},
				NumRequests: 1,
			}
		} else {
			respMap[url2].Status = append(respMap[url2].Status, status)
			respMap[url2].RespTimes = append(respMap[url2].RespTimes, respTime)
			respMap[url2].NumRequests += 1
		}
	}
}

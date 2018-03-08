package perf

import (
	"encoding/json"
	"fmt"
	"github.com/gnulnx/color"
	"github.com/gnulnx/goperf/request"
	"net/http"
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
	Cookies    string
	UserAgent  string
}

func (input *Init) Basic() request.IterateReqRespAll {
	// Create slice of channels to hold results
	// Fire off anonymous go routine using newly created channel
	chanslice := []chan request.IterateReqRespAll{}
	for i := 0; i < input.Threads; i++ {
		chanslice = append(chanslice, make(chan request.IterateReqRespAll))
		go func(c chan request.IterateReqRespAll) {

			// Make an initial GET request to get and set cookies so we can accurately simulate a user.
			// This effectivly sets up a user session.  If this is commented out
			// then each request will simulate a new user.
			// TODO This should be a parameter the user can set.
			resp1, _ := http.Get(input.Url)
			if resp1 == nil {
				fmt.Println("Error connecting to url: ", input.Url)
				return
			}
			if len(resp1.Header["Set-Cookie"]) > 0 {
				input.Cookies = resp1.Header["Set-Cookie"][0]
			}
			// TODO Just pass the Input in
			c <- iterateRequest(input)
		}(chanslice[i])
		//time.Sleep(time.Duration(1000))
	}

	// Wait on all the channels
	results := []request.IterateReqRespAll{}
	for _, ch := range chanslice {
		results = append(results, <-ch)
	}

	input.Results = request.Combine(results)
	return *input.Results

}

func iterateRequest(input *Init) request.IterateReqRespAll {
	/*
		Continuously fetch 'url' for 'sec' second and return the results.
	*/
	url := input.Url
	sec := input.Seconds
	cookies := input.Cookies
	useragent := input.UserAgent
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
	var totalLinearRespTimes int64 = 0
	var count int64 = 0 // TODO for loop counter instead???
	for {
		//Fetch the url and all the js, css, and img assets
		fetchAllResp := request.FetchAll(request.FetchInput{
			BaseURL:   url,
			Retdat:    false,
			Cookies:   cookies,
			UserAgent: useragent,
		})

		// Set base resp properties
		resp.Status = append(resp.Status, fetchAllResp.BaseURL.Status)
		resp.RespTimes = append(resp.RespTimes, fetchAllResp.BaseURL.Time)
		resp.Bytes = fetchAllResp.TotalBytes

		totalRespTimes += int64(fetchAllResp.TotalTime)
		totalLinearRespTimes += int64(fetchAllResp.TotalLinearTime)

		gatherAllStats(fetchAllResp, jsMap, cssMap, imgMap)

		elapsedTime = time.Now().Sub(start)
		count += 1
		if elapsedTime > maxTime {
			break
		}
	}

	avgTotalRespTimes := time.Duration(totalRespTimes / count)
	avgTotalLinearRespTimes := time.Duration(totalLinearRespTimes / count)

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
		BaseURL:                resp,
		AvgTotalRespTime:       avgTotalRespTimes,
		AvgTotalLinearRespTime: avgTotalLinearRespTimes,
		JSResps:                jsResps,
		CSSResps:               cssResps,
		IMGResps:               imgResps,
	}
	return output
}

type BaseURL struct {
	Url                 string         `json:"base_url"`
	Numreqs             int            `json:"num_reqs"`
	TotBytes            int            `json:"total_bytes"`
	AvgPageRespTime     time.Duration  `json:"avg_page_resp_time"`
	AvgTimeToFirsttByte time.Duration  `json:"avg_time_to_first_byte"`
	Status              map[string]int `json:"status"`
}

type AssetResult struct {
	Url         string         `json:"url"`
	AvgRespTime time.Duration  `json:"avg_resp_time"`
	Status      map[string]int `json:"status"`
}

type Output struct {
	BaseURL    BaseURL       `json:"base_url"`
	JSResults  []AssetResult `json:"js_assets"`
	CSSResults []AssetResult `json:"css_assets"`
	IMGResults []AssetResult `json:"img_assets"`
}

func (input Init) JsonAll() {
	/*
		Prints out all of the JSON data.  Useful mainful for debugging
	*/
	tmp, _ := json.MarshalIndent(input.Results, "", "  ")
	fmt.Println(string(tmp))
}

func (input Init) JsonResults() string {
	/*
		Return only the performance result json.
	*/
	results := input.Results

	avg, statusResults := procResult(&results.BaseURL)
	output := Output{
		BaseURL: BaseURL{
			Url:                 results.BaseURL.Url,
			Numreqs:             len(results.BaseURL.Status),
			TotBytes:            results.BaseURL.Bytes,
			AvgPageRespTime:     results.AvgTotalRespTime,
			AvgTimeToFirsttByte: avg,
			Status:              statusResults,
		},
		JSResults:  buildAssetSlice(results.JSResps),
		CSSResults: buildAssetSlice(results.CSSResps),
		IMGResults: buildAssetSlice(results.IMGResps),
	}

	tmp, _ := json.MarshalIndent(output, "", "    ")
	output_json := string(tmp)
	return output_json
}

func buildAssetSlice(resps []request.IterateReqResp) []AssetResult {
	results := []AssetResult{}
	for _, resp := range resps {
		avg, statusResults := procResult(&resp)
		result := AssetResult{
			Url:         resp.Url,
			AvgRespTime: avg,
			Status:      statusResults,
		}
		results = append(results, result)
	}
	return results
}

func (input Init) Print() {
	results := input.Results
	yel := color.New(color.FgHiYellow).SprintfFunc()
	yellow := color.New(color.FgHiYellow, color.Underline).SprintfFunc()
	grey := color.New(color.FgHiBlack).SprintfFunc()
	white := color.New(color.FgWhite).SprintfFunc()

	color.Red("Base Url Results")
	fmt.Printf(" - %-45s %-25s\n", yel("Url:"), white(results.BaseURL.Url))
	fmt.Printf(" - %-45s %-25s\n", yel("Number of Requests:"), white(strconv.Itoa(len(results.BaseURL.Status))))
	fmt.Printf(" - %-45s %s\n", yel("Total Bytes:"), white(strconv.Itoa(results.BaseURL.Bytes)))
	fmt.Printf(" - %-45s %s\n", yel("Avg Page Resp Time:"), white(results.AvgTotalRespTime.String()))

	// TODO Remove this and associated code
	// This shows the total time to do a linear fetch asset by asset.
	//fmt.Printf(" - %-45s %s\n", yel("Avg Linear Resp Time:"), white(results.AvgTotalLinearRespTime.String()))

	// This is useful for comparing the decrease in resp time from linear to go routines
	// decrease := float64(results.AvgTotalLinearRespTime) - float64(results.AvgTotalRespTime)
	// percentDecrease := (float64(decrease) / float64(results.AvgTotalLinearRespTime) * 100.00)
	// fmt.Printf(" - %-45s %s\n", yel("percentDecrease:"), white(strconv.FormatFloat(percentDecrease, 'g', 5, 64)))

	avg, statusResults := procResultString(&results.BaseURL)
	fmt.Printf(" - %-45s %s\n", yel("Average Time to First Byte:"), white(avg))
	fmt.Printf(" - %-45s %s\n", yel("Status:"), white(statusResults))

	printAssets := func(title string, results []request.IterateReqResp) {
		color.Red(title)
		fmt.Printf(" - %-28s %-30s %-21s %-10s\n", yellow("Average"), yellow("Status"), yellow("Bytes"), yellow("Url"))
		for i, resp := range results {
			avg, statusResults := procResultString(&resp)
			if i%2 == 0 {
				fmt.Printf(" - %-26s %-28s %-19s %-10s\n", grey(avg), grey(statusResults), grey(strconv.Itoa(resp.Bytes)), grey(resp.Url))
			} else {
				fmt.Printf(" - %-26s %-28s %-19s %-10s\n", white(avg), white(statusResults), white(strconv.Itoa(resp.Bytes)), white(resp.Url))
			}
		}
	}
	printAssets("JS Results", results.JSResps)
	printAssets("CSS Results", results.CSSResps)
	printAssets("IMG Results", results.IMGResps)
}

func procResultString(resp *request.IterateReqResp) (string, string) {
	avg, statusResults := procResult(resp)
	tmp, _ := json.Marshal(statusResults)
	status := string(tmp)
	return avg.String(), status
}

func procResult(resp *request.IterateReqResp) (time.Duration, map[string]int) {
	totalTime := time.Duration(0)
	for _, val := range resp.RespTimes {
		totalTime += val
	}
	avg := time.Duration(int64(totalTime) / int64(len(resp.Status)))

	statusCodes := map[string][]int{}
	for _, val := range resp.Status {
		status := strconv.Itoa(val)
		statusCodes[status] = append(statusCodes[status], val)
	}

	statusResults := make(map[string]int)
	for key, _ := range statusCodes {
		statusResults[key] = len(statusCodes[key])
	}

	return avg, statusResults
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

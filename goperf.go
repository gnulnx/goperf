package main

/*
goperf is a load testing tool.

** Use Case 1:  Fetch a url and report stats

This command will return all information for a given url.
./goperf -url http://qa.teaquinox.com -fetchall -printjson

When fetchall is provided the returned struct will contain
url, time, size, and data info.

You can do a simpler request that leaves the data and headers out like this
./goperf -url http://qa.teaquinox.com -fetchall -printjson


** Use Case 2: Load testing


*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/gnulnx/goperf/request"
	"os"
	"runtime/pprof"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	// I ❤️  the way go handles command line arguments
	fetch := flag.Bool("fetch", false, "Fetch -url and report it's stats. Does not return resources")
	fetchall := flag.Bool("fetchall", false, "Fetch -url and report stats  return all assets (js, css, img)")
	printjson := flag.Bool("printjson", false, "Print json output")
	threads := flag.Int("connections", 1, "Number of concurrent connections")
	url := flag.String("url", "https://qa.teaquinox.com", "url to test")
	seconds := flag.Int("sec", 2, "Number of seconds each concurrant thread/user should make requests")
	iterations := flag.Int("iter", 1000, "Iterations per thread")
	output := flag.Int("output", 5, "Show thread output every {n} iterations")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	flag.Parse()

	if *fetch || *fetchall {

		color.Green("~~ Fetching a single url and printing info ~~")
		resp := request.FetchAll(*url, *fetchall)

		if *printjson {
			tmp, _ := json.MarshalIndent(resp, "", "    ")
			fmt.Println(string(tmp))
		}

		request.PrintFetchAllResponse(resp)

		os.Exit(1)
	}

	// TODO Declare an inline parameter struct...
	input := request.Input{
		Iterations: *iterations,
		Threads:    *threads,
		Url:        *url,
		Output:     *output,
		Verbose:    *verbose,
		Seconds:    *seconds,
	}
	f, _ := os.Create(*cpuprofile)
	pprof.StartCPUProfile(f)
	perf(input)

	defer pprof.StopCPUProfile()
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
			fmt.Println(url2)
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

func iterateRequest(url string, sec int) request.IterateReqRespAll {
	start := time.Now()
	maxTime := time.Duration(sec) * time.Second
	elapsedTime := maxTime

	resp := request.IterateReqResp{
		Url: url,
	}
	jsMap := map[string]*request.IterateReqResp{}
	cssMap := map[string]*request.IterateReqResp{}
	imgMap := map[string]*request.IterateReqResp{}

	// Remove this when you can
	resps := []request.FetchAllResponse{}
	for {
		//Fetch the url
		fetchAllResp := request.FetchAll(url, false)
		//tmp, _ := json.MarshalIndent(fetchAllResp, "", "    ")
		//fmt.Println(string(tmp))
		//fmt.Println(fetchAllResp.Status)

		// Set base resp properties
		resp.Status = append(resp.Status, fetchAllResp.BaseUrl.Status)
		resp.RespTimes = append(resp.RespTimes, fetchAllResp.BaseUrl.Time)
		resp.Bytes = fetchAllResp.TotalBytes

		gatherStats(fetchAllResp.JSResponses, jsMap)
		gatherStats(fetchAllResp.CSSResponses, cssMap)
		gatherStats(fetchAllResp.IMGResponses, imgMap)

		// This is the old way... it will be removed
		resps = append(resps, *fetchAllResp)

		elapsedTime = time.Now().Sub(start)
		if elapsedTime > maxTime {
			break
		}
	}

	// NOTE Do you want to return the elapsed time?
	/*
		fmt.Println("----------------------------")
		fmt.Println(" - total: ", elapsedTime)
		fmt.Println(" - Num of Requests: ", int64(count))
		avg := time.Duration(int64(elapsedTime) / int64(count))
		fmt.Println(" - avg: ", avg)
		fmt.Println("----------------------------")
	*/

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
		BaseUrl:  resp,
		JSResps:  jsResps,
		CSSResps: cssResps,
		IMGResps: imgResps,
	}
	return output
}

func perf(input request.Input) time.Duration {
	// Print input params as JSON
	tmp, _ := json.MarshalIndent(input, "", "    ")
	fmt.Println(string(tmp))

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
	totalReqs := 0
	for ch := 0; ch < len(chanslice); ch++ {
		results = append(results, <-chanslice[ch])
	}

	/*
		TODO Next steps.
		1) Figureout why BaseUrl doesn't have status
		Combine all IterateReqResp in results into a single
		IterateReqResp struct.  That is what we want to return from here
	*/

	tmp, _ = json.MarshalIndent(results, "", "    ")
	fmt.Println(string(tmp))

	f, _ := os.Create("./results.json")
	f.WriteString(string(tmp))
	color.Magenta("json results in results.json")
	color.Yellow("len results: %d", len(results))
	color.Yellow("total reqs: %d", totalReqs)
	for i := 0; i < len(results); i++ {

	}
	return 0
}

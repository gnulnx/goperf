package request

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gnulnx/goperf/httputils"
	"strconv"
	"time"
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

func FetchAllAssetArray(files []string, baseurl string, retdat bool, resp chan []FetchResponse) {
	responses := []FetchResponse{}

	// TODO  What if this was go routines instead?
	// NOTE: Then you end up hyper threaded which is perfect right?
	// Look wait group example below
	// https://nathanleclaire.com/blog/2014/02/15/how-to-wait-for-all-goroutines-to-finish-executing-before-continuing/
	for _, asset_url := range files {
		responses = append(
			responses,
			*FetchAsset(baseurl, asset_url, retdat),
		)
	}

	resp <- responses
}

func FetchAll(baseurl string, retdat bool) *FetchAllResponse {
	/*
	   Fetch the url and then fetch all of it's assets.
	   Assets currently refer to script, style, and img tags.
	   Each asset class is fetched in it's own go routine

	   If retdata is False we don't return the Body or Header
	   This is useful if you only want the timing data.
	   For instance you might find it useful to fetch with retdat=true
	   the first time around to get all the data and write to file.
	   The subsequet requests could be used as part of a perf test where
	   you only need the raw timing and size data.  In those cases
	   you can set retdat=false to effectivly cut down on the verbosity
	*/
	// Fetch initial url
	start := time.Now()
	output := Fetch(baseurl, true)

	// Now parse output for js, css, img urls
	jsfiles, imgfiles, cssfiles := httputils.ParseAllAssets(output.body)

	// Now lets create some go routines and fetch all the js, img, css files
	c1 := make(chan []FetchResponse)
	c2 := make(chan []FetchResponse)
	c3 := make(chan []FetchResponse)

	go FetchAllAssetArray(jsfiles, baseurl, retdat, c1)
	go FetchAllAssetArray(imgfiles, baseurl, retdat, c2)
	go FetchAllAssetArray(cssfiles, baseurl, retdat, c3)

	jsResponses := []FetchResponse{}
	imgResponses := []FetchResponse{}
	cssResponses := []FetchResponse{}

	for i := 0; i < 3; i++ {
		select {
		case jsResponses = <-c1:
		case imgResponses = <-c2:
		case cssResponses = <-c3:
		}
	}

	if !retdat {
		output.body = ``
		output.headers = make(map[string][]string)
	}

	resp := FetchAllResponse{
		baseUrl:      output,
		time:         time.Now().Sub(start),
		jsResponses:  jsResponses,
		imgResponses: imgResponses,
		cssResponses: cssResponses,
	}

	return &resp
}

// Why is this here?  move this to the request.fetch.go modules
func PrintFetchAllResponse(resp *FetchAllResponse) {
	color.Red("Base Url Results")

	if resp.baseUrl.status == 200 {
		color.Green(" - Status: " + strconv.Itoa(resp.baseUrl.status))
	} else {
		color.Red(" - Status: " + strconv.Itoa(resp.baseUrl.status))
	}

	total := resp.baseUrl.time
	color.Yellow(" - Url: " + resp.baseUrl.url)
	color.Yellow(" - Time to first byte: " + total.String())
	color.Yellow(" - Bytes: " + strconv.Itoa(resp.baseUrl.bytes))
	color.Yellow(" - Runes: " + strconv.Itoa(resp.baseUrl.runes))

	// This part will work for a single response
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// This really needs to be factored back into the request module
	calcTotal := func(resp []FetchResponse) time.Duration {
		total := time.Duration(0)
		for _, val := range resp {
			total += val.time
		}
		return total
	}

	total += calcTotal(resp.jsResponses)
	total += calcTotal(resp.cssResponses)
	total += calcTotal(resp.imgResponses)

	color.Magenta(" - Total Time: %s", total.String())

	color.Red("JS Responses")
	fmt.Printf(" - %-22s %-15s %-50s \n", green("Time"), green("Bytes"), green("Url"))
	for _, val := range resp.jsResponses {
		total += val.time
		fmt.Printf(" - %-22s %-15s %-50s \n", green(val.time.String()), yellow(strconv.Itoa(val.bytes)), val.url)
	}

	color.Red("CSS Responses")
	fmt.Printf(" - %-22s %-15s %-50s \n", green("Time"), green("Bytes"), green("Url"))
	for _, val := range resp.cssResponses {
		total += val.time
		fmt.Printf(" - %-22s %-15s %-50s \n", green(val.time.String()), yellow(strconv.Itoa(val.bytes)), val.url)
	}

	color.Red("IMG Responses")
	fmt.Printf(" - %-22s %-15s %-50s \n", green("Time"), green("Bytes"), green("Url"))
	for _, val := range resp.imgResponses {
		total += val.time
		fmt.Printf(" - %-22s %-15s %-50s \n", green(val.time.String()), yellow(strconv.Itoa(val.bytes)), val.url)
	}
}

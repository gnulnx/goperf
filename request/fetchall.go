package request

import (
	"fmt"
	"github.com/gnulnx/color"
	"github.com/gnulnx/goperf/httputils"
	"net/url"
	"strconv"
	"time"
	"unicode"
)

func FetchAll(input FetchInput) *FetchAllResponse {
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
	baseurl := input.BaseUrl
	retdat := input.Retdat
	//cookies := input.Cookies

	// Fetch initial url
	start := time.Now()
	input.Retdat = true
	output := Fetch(input)

	output.Body = StripBody(output.Body)

	// Now parse output for js, css, img urls
	jsfiles, imgfiles, cssfiles := httputils.ParseAllAssets(output.Body)

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
	totalTime2 := time.Now().Sub(start)

	if !retdat {
		output.Body = ``
		output.Headers = make(map[string][]string)
	}

	calcTotal := func(resp []FetchResponse) (time.Duration, int) {
		totalTime := time.Duration(0)
		totalBytes := 0
		for _, val := range resp {
			totalTime += val.Time
			totalBytes += val.Bytes
		}
		return totalTime, totalBytes
	}

	jsTime, jsBytes := calcTotal(jsResponses)
	cssTime, cssBytes := calcTotal(cssResponses)
	imgTime, imgBytes := calcTotal(imgResponses)

	totalLinearTime := output.Time + jsTime + cssTime + imgTime
	totalBytes := output.Bytes + jsBytes + cssBytes + imgBytes

	resp := FetchAllResponse{
		BaseUrl:         output,
		Time:            output.Time,
		TotalTime:       totalTime2,
		TotalLinearTime: totalLinearTime,
		TotalBytes:      totalBytes,
		JSResponses:     jsResponses,
		IMGResponses:    imgResponses,
		CSSResponses:    cssResponses,
	}

	return &resp
}

func PrintFetchAllResponse(resp *FetchAllResponse) {
	yel := color.New(color.FgHiYellow).SprintfFunc()
	yellow := color.New(color.FgHiYellow, color.Underline).SprintfFunc()
	grey := color.New(color.FgHiBlack).SprintfFunc()
	white := color.New(color.FgWhite).SprintfFunc()
	red := color.New(color.FgRed).SprintfFunc()
	green := color.New(color.FgGreen).SprintfFunc()
	total := resp.BaseUrl.Time

	color.Red("Base Url Results")
	if resp.BaseUrl.Status == 200 {
		fmt.Printf(" - %-34s %-25s\n", yel("Status:"), green(strconv.Itoa(resp.BaseUrl.Status)))
	} else {
		fmt.Printf(" - %-34s %-25s\n", yel("Status:"), red(strconv.Itoa(resp.BaseUrl.Status)))
	}
	fmt.Printf(" - %-34s %-25s\n", yel("Url:"), white(resp.BaseUrl.Url))
	fmt.Printf(" - %-34s %-25s\n", yel("Time to first byte"), total.String())
	fmt.Printf(" - %-34s %-25s\n", yel("Bytes"), strconv.Itoa(resp.BaseUrl.Bytes))
	fmt.Printf(" - %-34s %-25s\n", yel("Runes"), strconv.Itoa(resp.BaseUrl.Runes))
	fmt.Printf(" - %-34s %-25s\n", yel("TotalTime"), resp.TotalTime.String())
	fmt.Printf(" - %-34s %-25s\n", yel("TotalBytes"), strconv.Itoa(resp.TotalBytes))

	printAssets := func(title string, results []FetchResponse) {
		color.Red(title)
		fmt.Printf(" - %-24s %-22s %-21s\n", yellow("Time"), yellow("Bytes"), yellow("Url"))
		for i, val := range results {
			if i%2 == 0 {
				fmt.Printf(" - %-22s %-20s %-10s \n", white(val.Time.String()), white(strconv.Itoa(val.Bytes)), white(val.Url))
			} else {
				fmt.Printf(" - %-22s %-20s %-10s \n", grey(val.Time.String()), grey(strconv.Itoa(val.Bytes)), grey(val.Url))
			}
		}
	}

	printAssets("JS Responses", resp.JSResponses)
	printAssets("CSS Responses", resp.CSSResponses)
	printAssets("IMG Responses", resp.IMGResponses)
}

func StripBody(input string) string {
	var output string = ""
	for _, c := range input {
		if !unicode.IsSpace(c) {
			output = output + string(c)
		}
	}
	return output
}

func DefineAssetUrl(baseurl string, asseturl string) string {
	/*
		If the url starts with a / we know it's a local resource
		so we prepend the baseurl to it
	*/
	if asseturl[:4] == "http" {
		return asseturl
	}

	u, err := url.Parse(baseurl)
	if err != nil {
		panic(err)
	}
	baseurl = fmt.Sprintf("%s://%s", u.Scheme, u.Hostname())

	if asseturl[0] == '/' {
		asseturl = baseurl + asseturl
	} else {
		asseturl = baseurl + "/" + asseturl
	}
	return asseturl
}

func FetchAsset(baseurl string, asseturl string, retdat bool) *FetchResponse {
	// TODO This method would be cleaner if you passed the FetchInput struct in
	asset_url := DefineAssetUrl(baseurl, asseturl)
	return Fetch(FetchInput{
		BaseUrl: asset_url,
		Retdat:  retdat,
	})
}

func FetchAllAssetArray(files []string, baseurl string, retdat bool, resp chan []FetchResponse) {
	responses := []FetchResponse{}

	// TODO  Look into turning this into go routines
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

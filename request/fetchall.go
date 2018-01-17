package request

import (
	"github.com/fatih/color"
	"github.com/gnulnx/goperf/httputils"
	"strconv"
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

func FetchAllAssetArray(files []string, baseurl string, retdat bool, resp chan []FetchResponse) []FetchResponse {
	responses := []FetchResponse{}

	for i := 0; i < len(files); i++ {
		asset_url := (files)[i]
		responses = append(responses, *FetchAsset(baseurl, asset_url, retdat))
	}

	resp <- responses
	return responses
}

func FetchAll(baseurl string, retdat bool) *FetchAllResponse {
	/*
	   Fetch the url and then fetch all of it's assets.
	   Assets currently refer to script, style, and img tags.

	   If retdata is False we don't return the Body or Header
	   This is useful if you only want the timing data.
	   For instance you might find it useful to fetch with retdat=true
	   the first time around to get all the data and write to file.
	   The subsequet requests could be used as part of a perf test where
	   you only need the raw timing and size data.  In those cases
	   you can set retdat=false to effectivly cut down on the verbosity
	*/
	// Fetch initial url
	output := Fetch(baseurl, true)

	// Now parse for js, css, img urls
	jsfiles, imgfiles, cssfiles, bundle := httputils.Resources(output.Body)

	c1 := make(chan []FetchResponse)
	c2 := make(chan []FetchResponse)
	c3 := make(chan []FetchResponse)

	// Not sure a browser kicks off 3 threads.  1 for each asset class
	// But I'm pretty sure for load testing purposes this will be decent
	go FetchAllAssetArray(*jsfiles, baseurl, retdat, c1)
	go FetchAllAssetArray(*imgfiles, baseurl, retdat, c2)
	go FetchAllAssetArray(*cssfiles, baseurl, retdat, c3)

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
		output.Body = ``
		output.Headers = make(map[string][]string)
	}
	outputall := FetchAllResponse{
		BaseUrl:      output,
		JSReponses:   jsResponses,
		IMGResponses: imgResponses,
		CSSResponses: cssResponses,
	}

	// TODO You need to control output with a cmd line switch
	color.Red("Fetching: " + output.Url)
	if output.Status == 200 {
		color.Green(" - Status: " + strconv.Itoa(output.Status))
	} else {
		color.Red(" - Status: " + strconv.Itoa(output.Status))
	}
	color.Yellow(" - Time to first byte: " + output.Time.String())
	color.Yellow(" - Bytes: " + strconv.Itoa(output.Bytes))
	color.Yellow(" - Runes: " + strconv.Itoa(output.Runes))
	log("Javascript files", jsfiles)
	log("CSS files", cssfiles)
	log("IMG files", imgfiles)
	log("Full Bundle", bundle)

	return &outputall
}

package request

import (
	"github.com/gnulnx/goperf/httputils"
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

	if !retdat {
		output.Body = ``
		output.Headers = make(map[string][]string)
	}

	resp := FetchAllResponse{
		BaseUrl:      output,
		Time:         time.Now().Sub(start),
		JSResponses:  jsResponses,
		IMGResponses: imgResponses,
		CSSResponses: cssResponses,
	}

	return &resp
}

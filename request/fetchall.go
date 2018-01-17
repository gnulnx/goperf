package request

import (
	"github.com/fatih/color"
	"github.com/gnulnx/goperf/httputils"
	"strconv"
)

func FetchAll(baseurl string, retdat bool) *FetchAllResponse {
	/*
	   Try to simulate a request.
	   1) Fetch base_url
	   2) Parse for script, css, and img tags
	   3) Fetch other resources with go threads
	   4) compile results and return as json
	*/
	// Fetch initial url
	output := Fetch(baseurl, true)

	// Now parse for js, css, img urls
	jsfiles, imgfiles, cssfiles, bundle := httputils.Resources(output.Body)

	// TODO:  This needs to be done as Go routines to simulate a real browser
	jsResponses := []FetchResponse{}
	files := *jsfiles
	for i := 0; i < len(files); i++ {
		asset_url := (files)[i]
		jsResponses = append(jsResponses, *FetchAsset(baseurl, asset_url, retdat))
		color.Magenta(asset_url)
	}

	imgResponses := []FetchResponse{}
	files = *imgfiles
	for i := 0; i < len(files); i++ {
		asset_url := (files)[i]
		imgResponses = append(imgResponses, *FetchAsset(baseurl, asset_url, retdat))
		color.Magenta(asset_url)
	}

	cssResponses := []FetchResponse{}
	files = *cssfiles
	for i := 0; i < len(files); i++ {
		asset_url := (files)[i]
		cssResponses = append(cssResponses, *FetchAsset(baseurl, asset_url, retdat))
		color.Magenta(asset_url)
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
	//tmp, _ := json.MarshalIndent(outputall, "", "    ")
	//fmt.Println(string(tmp))

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

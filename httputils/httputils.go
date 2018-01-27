package httputils

import (
	"regexp"
)

func ParseAllAssetsSequential(body string) (js []string, img []string, css []string) {
	jsfiles := GetJS(body)
	cssfiles := GetCSS(body)
	imgfiles := GetIMG(body)
	return jsfiles, imgfiles, cssfiles
}

func ParseAllAssets(body string) (js []string, img []string, css []string) {
	/*
		Parse string of text (typically from a http.Response.Body)
		and return it's assets:  js, css, img.

		Note:  In go it is literally faster to start seperate go routines for each asset rather than
			fetch them sequetially.  The go routine overhead is miniscule.  Go literally fucking rocks...
	*/

	// make some channels
	c1 := make(chan []string)
	c2 := make(chan []string)
	c3 := make(chan []string)

	//kick off our annonymous go routines.
	go func() { c1 <- GetJS(body) }()
	go func() { c2 <- GetIMG(body) }()
	go func() { c3 <- GetCSS(body) }()

	//collect our results
	jsfiles := []string{}
	imgfiles := []string{}
	cssfiles := []string{}

	for i := 0; i < 3; i++ {
		select {
		case jsfiles = <-c1:
		case imgfiles = <-c2:
		case cssfiles = <-c3:
		}
	}

	return jsfiles, imgfiles, cssfiles
}

func GetJS(body string) []string {
	return runregex(`<script.*?src=["'\''](.*?)["'\''].*?script>`, body)
	//return runregex(`<script.*?src=["'\''](.*?)["'\'']`, body)
	//return runregex(`<script.*?src="(.*?)"`, body)
}

func GetCSS(body string) []string {
	return runregex(`<link.*?href=["'\''](.*?)["'\''].*?>`, body)
}

func GetIMG(body string) []string {
	backgroundimgs := runregex(`background-image: url\(["'\''](.*?)["'\'']`, body)
	imgs := runregex(`<img(?s:.)*?src=["'\''](.*?)["'\'']`, body)
	all := append(imgs, backgroundimgs...)
	return all
}

func runregex(expr string, body string) []string {
	r, _ := regexp.Compile(expr)
	match := r.FindAllStringSubmatch(body, -10)
	files := make([]string, 0)
	for j := 0; j < len(match); j++ {
		files = append(files, match[j][1])
	}
	return files
}

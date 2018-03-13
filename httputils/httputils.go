package httputils

import (
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

/*
	ParseAllAssetsSequential takes a string of text (typically from a http.Response.Body)
	and return the urls for the page <script> <link> and <img> tag.
	The method runs lineearly.
	In benchmark you will see that ParseAllAssets is generally faster and GetAssets is faster still
*/
func ParseAllAssetsSequential(body string) (js []string, img []string, css []string) {
	jsfiles := GetJS(body)
	cssfiles := GetCSS(body)
	imgfiles := GetIMG(body)
	return jsfiles, imgfiles, cssfiles
}

/*
GetAssets takes a string of test from an http.Response.Body and returns the
urls for the page <script>, <link>, and <img> tags.
It makes use of the goquery library and is currently the fastest method
*/
func GetAssets(body string) (js []string, img []string, css []string) {
	utfBody := strings.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		log.Println("Unable to parse document with goquery.  Make sure it is utf8")
		return ParseAllAssets(body)
	}
	goroutine := false
	jsfiles := []string{}
	imgfiles := []string{}
	cssfiles := []string{}

	if goroutine {
		c1 := make(chan []string)
		c2 := make(chan []string)
		c3 := make(chan []string)

		go func() { c1 <- getAttr(doc, "script", "src") }()
		go func() { c2 <- getAttr(doc, "img", "src") }()
		go func() { c3 <- getAttr(doc, "link", "href") }()

		for i := 0; i < 3; i++ {
			select {
			case jsfiles = <-c1:
			case imgfiles = <-c2:
			case cssfiles = <-c3:
			}
		}
	} else {
		jsfiles = getAttr(doc, "script", "src")
		imgfiles = getAttr(doc, "img", "src")
		cssfiles = getAttr(doc, "link", "href")
	}

	return jsfiles, imgfiles, cssfiles
}

/*
	geAttr takes a *goquery.Document a html tag and attr
	and returns a list of those attributes
*/
func getAttr(doc *goquery.Document, tag string, attr string) []string {
	files := []string{}
	doc.Find(tag).Each(func(i int, s *goquery.Selection) {
		value, exists := s.Attr(attr)
		if exists {
			files = append(files, value)
		}
	})
	return files
}

/*
	ParseAllAssets takes string of text (typically from a http.Response.Body)
	and return the urls for the page <script> <link> and <img> tag.
	The method uses seperate go routines for each asset class.
	It is faster than ParseAllAssetsSequentially, but still slow that GetAssets
*/
func ParseAllAssets(body string) (js []string, img []string, css []string) {
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

// GetJS uses regex to parse a body of text and return the script src attributes
func GetJS(body string) []string {
	return runregex(`<script.*?src=["'\''](.*?)["'\''].*?script>`, body)
}

// GetCSS uses regex to parse a body of text and return the <link> href attributes
func GetCSS(body string) []string {
	return runregex(`<link.*?href=["'\''](.*?)["'\''].*?>`, body)
}

// GetIMG uses regex to parse a body of text and return the <img> src attributes
func GetIMG(body string) []string {
	backgroundimgs := runregex(`background-image: url\(["'\''](.*?)["'\'']`, body)
	imgs := runregex(`<img(?s:.)*?src=["'\''](.*?)["'\'']`, body)
	all := append(imgs, backgroundimgs...)
	return all
}

// Take a regex expression that returns the matched object
// and return an array of the matched text
func runregex(expr string, body string) []string {
	r, _ := regexp.Compile(expr)
	match := r.FindAllStringSubmatch(body, -10)
	files := make([]string, 0)
	for j := 0; j < len(match); j++ {
		files = append(files, match[j][1])
	}
	return files
}

package httputils

import (
	"regexp"
)

func ParseAllAssets(body string) (js []string, img []string, css []string) {

	c1 := make(chan []string)
	c2 := make(chan []string)
	c3 := make(chan []string)

	go func() {
		c1 <- GetJS(body)
	}()
	go func() {
		c2 <- GetIMG(body)
	}()
	go func() {
		c3 <- GetCSS(body)
	}()

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
	return runregex(`<script.*?src="(.*?)"`, body)
}

func GetCSS(body string) []string {
	return runregex(`<link.*?href="(.*?)"`, body)
}

func GetIMG(body string) []string {
	return runregex(`<img.*?src="(.*?)"`, body)
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

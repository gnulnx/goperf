package httputils

import (
	"regexp"
)

func Resources(body string) (*[]string, *[]string, *[]string, *[]string) {

	c1 := make(chan *[]string)
	c2 := make(chan *[]string)
	c3 := make(chan *[]string)

	go func() {
		c1 <- Getjs(body)
	}()
	go func() {
		c2 <- Getimg(body)
	}()
	go func() {
		c3 <- Getcss(body)
	}()

	jsfiles := &[]string{}
	imgfiles := &[]string{}
	cssfiles := &[]string{}

	for i := 0; i < 3; i++ {
		select {
		case jsfiles = <-c1:
		case imgfiles = <-c2:
		case cssfiles = <-c3:
		}
	}

	// Worth perf testing this because the below code is really trivial
	//jsfiles := Getjs(body)
	//imgfiles := Getimg(body)
	//cssfiles := Getcss(body)

	// Create a full Bundle... Maybe this is all you need?
	// TODO It's not effecient to regex them twice like this...
	bundle := append(
		append(*jsfiles, *Getcss(body)...),
		*Getimg(body)...,
	)

	return jsfiles, imgfiles, cssfiles, &bundle
}

func Getjs(body string) *[]string {
	return runregex(`<script.*?src="(.*?)"`, body)
}

func Getimg(body string) *[]string {
	return runregex(`<img.*?src="(.*?)"`, body)
}

func Getcss(body string) *[]string {
	return runregex(`<link.*?href="(.*?)"`, body)
}

func runregex(expr string, body string) *[]string {
	r, _ := regexp.Compile(expr)
	match := r.FindAllStringSubmatch(body, -10)
	files := make([]string, 0)
	for j := 0; j < len(match); j++ {
		files = append(files, match[j][1])
	}
	return &files
}

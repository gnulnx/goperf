package httputils

import (
	"regexp"
)

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

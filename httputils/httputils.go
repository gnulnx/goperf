package main

import "fmt"

func Getjs(body string) *[]string {
	return runregex(`<script.*?src="(.*?)"`, body)
}

func Getimg(body string) *[]string {
	return runregex(`<img.*?src="(.*?)"`, body)
}

func Getcss(body string) *[]string {
	return runregex(`<link.*?href="(.*?)"`, body)
}

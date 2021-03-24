package httputils

import (
	"io/ioutil"
	"testing"

	"github.com/gnulnx/color"
	"gopkg.in/fatih/set.v0"
)

/*
getTestBody reads the test html file from the test_data directory
*/
func getTestBody() string {
	bodyBytes, _ := ioutil.ReadFile("test_data/test_basic.html")
	numBytes := len(bodyBytes)
	body := string(bodyBytes[:numBytes])
	return body
}

func testEquality(input []string, testdata []string, t *testing.T) bool {
	sInput := set.New(set.ThreadSafe)
	sExpected := set.New(set.ThreadSafe)
	for _, i := range input {
		sInput.Add(i)
	}
	for _, i := range testdata {
		sExpected.Add(i)
	}

	extra := set.Difference(sInput, sExpected)
	missing := set.Difference(sExpected, sInput)
	if !extra.IsEmpty() {
		t.Error("EXTRA", extra)
	}
	if !missing.IsEmpty() {
		t.Error("Missing", missing)
	}

	return true
}

func TestParseAllAssets(t *testing.T) {
	color.Green("~~ TestParseAll ~~")
	body := getTestBody()
	jsfiles, imgfiles, cssfiles := ParseAllAssets(body)

	// Test js results
	testData := []string{
		`/static/tcart/js/test1.min.js`,
		`/static/tcart/js/bundle_kldsf2334.min.js`,
	}
	testEquality(jsfiles, testData, t)

	// Test img results
	testData = []string{
		`/media//teaquinox_header_2.svg`,
		`/media/cart.svg`,
		`/media/banners/1-12-2018/SnowyTea_50percent.jpg`,
		`/media/banners/1-12-2018/SnowyTea_lowres2.jpg`,
		`/media/product_11/Shou_Mei_M.jpeg`,
		`/media/product_36/Turmeric_Chai_M.jpeg`,
		`/media/product_45/Luian_Gua_Pian_M.jpeg`,
		`/media/product_58/NEB_new_m.jpg`,
		`/media/product_71/Black_Dragon_Pearls_M.jpg`,
		`/media/product_None/Moroccan_Mint_M.jpg`,
		`/static/tcart/img/stripe_badges/outline_dark/powered_by_stripe.png`,
	}
	testEquality(imgfiles, testData, t)

	//Test css Results
	testData = []string{
		`/media/manifest.webmanifest`,
		`/static/vendor/icomoon/style.css`,
		`/media/favicon_94S_icon.ico`,
		`/static/vendor/bootstrap/bootstrap.min.css`,
		`/static/tcart/css/styles.min.css`,
	}
	testEquality(cssfiles, testData, t)
}

func BenchmarkParseAllAssets(b *testing.B) {
	body := getTestBody()
	for i := 0; i < b.N; i++ {
		ParseAllAssets(body)
	}
}

func BenchmarkParseAllAssetsSequential(b *testing.B) {
	body := getTestBody()
	for i := 0; i < b.N; i++ {
		ParseAllAssetsSequential(body)
	}
}

func BenchmarkGetAssets(b *testing.B) {
	body := getTestBody()
	for i := 0; i < b.N; i++ {
		GetAssets(body)
	}
}

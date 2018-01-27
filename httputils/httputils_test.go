package httputils

import (
	"github.com/gnulnx/color"
	"gopkg.in/fatih/set.v0"
	"io/ioutil"
	"reflect"
	"testing"
)

func get_test_body() string {
	body_bytes, _ := ioutil.ReadFile("test_data/test.html")
	num_bytes := len(body_bytes)
	body := string(body_bytes[:num_bytes])
	return body
}

func test_deep_equal(input []string, testdata []string, t *testing.T) bool {
	sInput := set.New()
	sExpected := set.New()
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
		//fmt.Println("MISSING", missing)
		t.Error("Missing", missing)
	}
	//fmt.Println("INPUT")

	if reflect.DeepEqual(input, testdata) != true {
		/*
			fmt.Println("\n-----------------------")
			color.Green("- input")
			for _, i := range input {
				color.Green(i)
			}
			//fmt.Println("- input: ", input)
			fmt.Println("- expected")
			for _, i := range testdata {
				color.Red(i)
			}
			t.Error("Slices above are not equal")
			return false
		*/
	}
	return true
}

func TestParseAllAssets(t *testing.T) {
	color.Green("~~ TestParseAll ~~")
	body := get_test_body()
	jsfiles, imgfiles, cssfiles := ParseAllAssets(body)

	// Test js results
	test_data := []string{
		`/static/tcart/js/test1.min.js`,
		`/static/tcart/js/bundle_kldsf2334.min.js`,
	}
	test_deep_equal(jsfiles, test_data, t)

	// Test img results
	test_data = []string{
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
	test_deep_equal(imgfiles, test_data, t)

	//Test css Results
	test_data = []string{
		`/media/manifest.webmanifest`,
		`/static/vendor/icomoon/style.css`,
		`/media/favicon_94S_icon.ico`,
		`/static/vendor/bootstrap/bootstrap.min.css`,
		`/static/tcart/css/styles.min.css`,
	}
	test_deep_equal(cssfiles, test_data, t)
}

func BenchmarkParseAllAssets(b *testing.B) {
	body := get_test_body()
	for i := 0; i < b.N; i++ {
		ParseAllAssets(body)
	}
}

func BenchmarkParseAllAssetsSequential(b *testing.B) {
	body := get_test_body()
	for i := 0; i < b.N; i++ {
		ParseAllAssetsSequential(body)
	}
}

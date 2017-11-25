package amazonjp

import (
	"fmt"
	"net/http"
	"testing"
)

func TestIsValidProductID(t *testing.T) {
	if !IsValidProductID("B01DUC3V14") {
		t.Errorf("\"B01DUC3V14\" is a valid ProductID")
	}
	if IsValidProductID("_") {
		t.Errorf("\"_\" is an invalid ProductID")
	}
}

func TestExtractProductIDFromURL(t *testing.T) {

	acceptableBaseProductURLs := []string{
		"https://www.amazon.co.jp/dp/product/",
		"https://www.amazon.co.jp/gp/product/",
		DefaultBaseProductURL,
	}
	// ベースURLパターン
	productID := "B01DUC3V14"
	for i, baseURL := range acceptableBaseProductURLs {
		url := baseURL + productID
		id, ok := ExtractProductIDFromURL(url)
		if !ok {
			t.Errorf("failed to extract ProductID from %v", url)
		}
		if id != "B01DUC3V14" {
			t.Errorf("\n[%d]\nExpected : %v\nActual : %v\nURL : %v", i, productID, id, url)
		}
	}
	// 末尾にパラメータあり。
	url := "https://www.amazon.co.jp/gp/product/" + productID + "/ref=series_rw_dp_sw"
	id, ok := ExtractProductIDFromURL(url)
	if !ok {
		t.Errorf("failed to extract ProductID from %v", url)
	}
	if id != "B01DUC3V14" {
		t.Errorf("Expected : %v\nActual : %v", productID, id)
	}
	// ランキングからのリンク。gp/dp の前に商品名
	url = "https://www.amazon.co.jp/%E5%AE%87%E5%AE%99%E5%85%84%E5%BC%9F%EF%BC%88%EF%BC%93%EF%BC%92%EF%BC%89-%E3%83%A2%E3%83%BC%E3%83%8B%E3%83%B3%E3%82%B0%E3%82%B3%E3%83%9F%E3%83%83%E3%82%AF%E3%82%B9-%E5%B0%8F%E5%B1%B1%E5%AE%99%E5%93%89-ebook/dp/B077G328Y2/ref=zg_bs_2293143051_1?_encoding=UTF8&psc=1&refRID=Z6D6K8STPQ853V1V0CB2"
	id, ok = ExtractProductIDFromURL(url)
	if !ok {
		t.Errorf("failed to extract ProductID from %v", url)
	}
	if id != "B077G328Y2" {
		t.Errorf("Expected : %v\nActual : %v", "B077G328Y2", id)
	}
	// ランキングからのリンク（相対アドレス）
	url = "/%E5%AE%87%E5%AE%99%E5%85%84%E5%BC%9F%EF%BC%88%EF%BC%93%EF%BC%92%EF%BC%89-%E3%83%A2%E3%83%BC%E3%83%8B%E3%83%B3%E3%82%B0%E3%82%B3%E3%83%9F%E3%83%83%E3%82%AF%E3%82%B9-%E5%B0%8F%E5%B1%B1%E5%AE%99%E5%93%89-ebook/dp/B077G328Y2/ref=zg_bs_2293143051_1?_encoding=UTF8&psc=1&refRID=7QM6SEMR96PRVP3BV9FX"
	id, ok = ExtractProductIDFromURL(url)
	if !ok {
		t.Errorf("failed to extract ProductID from %v", url)
	}
	if id != "B077G328Y2" {
		t.Errorf("Expected : %v\nActual : %v", "B077G328Y2", id)
	}
}

func TestNewProductFromURL(t *testing.T) {
	product, err := NewProductFromURL("https://www.amazon.co.jp/gp/product/B01DUC3V14")
	if err != nil {
		t.Error(err)
	}
	if product.ID != "B01DUC3V14" {
		t.Errorf("product.ID != B01DUC3V14")
	}
}

func TestNewProductFromId(t *testing.T) {
	product, err := NewProductFromID("B01DUC3V14")
	if err != nil {
		t.Error(err)
	}
	if product.ID != "B01DUC3V14" {
		t.Errorf("product.ID != B01DUC3V14")
	}
}

func TestProduct_Update(t *testing.T) {
	product, err := NewProductFromURL("https://www.amazon.co.jp/gp/product/B01DUC3V14")
	if err != nil {
		t.Error(err)
	}
	err = product.Update(http.DefaultClient)
	if err != nil {
		t.Error(err)
	}
	expectedTitle := "AIの遺電子　１ (少年チャンピオン・コミックス)"
	if product.Title != expectedTitle {
		t.Errorf("Expected : %v\nActual : %v", expectedTitle, product.Title)
	}
	fmt.Print(product)
}

package amazonjp

import (
	"fmt"
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
	// ベースURLパターン
	productID := "B01DUC3V14"
	for i, baseURL := range AcceptableBaseProductURLs {
		url := baseURL + productID
		id, err := ExtractProductIDFromURL(url)
		if err != nil {
			t.Error(err)
		}
		if id != "B01DUC3V14" {
			t.Errorf("\n[%d]\nExpected : B01DUC3V14\nActual : %v\nURL : %v", i, id, url)
		}
	}
	// 末尾にパラメータあり。
	id, err := ExtractProductIDFromURL("https://www.amazon.co.jp/gp/product/" + productID + "/ref=series_rw_dp_sw")
	if err != nil {
		t.Error(err)
	}
	if id != "B01DUC3V14" {
		t.Errorf("Expected : B01DUC3V14\nActual : %v", id)
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
	err = product.Update()
	if err != nil {
		t.Error(err)
	}
	expectedTitle := "AIの遺電子　１ (少年チャンピオン・コミックス)"
	if product.Title != expectedTitle {
		t.Errorf("Expected : %v\nActual : %v", expectedTitle, product.Title)
	}
	fmt.Print(product)
}

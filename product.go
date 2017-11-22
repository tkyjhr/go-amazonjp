package amazonjp

import (
	"fmt"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	// DefaultBaseProductURL は Amazon の商品情報ページのベース URL です。
	// 「ベース URL + 商品 ID」で商品情報ページにアクセスできます。
	DefaultBaseProductURL = "https://www.amazon.co.jp/dp/"
)

// IsValidProductID は id が有効な商品 ID かどうかを返します。有効かどうかは 0-9 及びアルファベットのみで構成されているかで判断されます。
func IsValidProductID(id string) bool {
	ok, _ := regexp.MatchString(`[^0-9a-zA-Z]`, id)
	return !ok
}

// GetProductURL は指定した商品 ID の商品情報ページの URL を返します。
func GetProductURL(id string) (string, bool) {
	if !IsValidProductID(id) {
		return "", false
	}
	return DefaultBaseProductURL + id, true
}

// ExtractProductIDFromURL は商品情報ページの URL から商品 ID を返します。
// 例えば「https://www.amazon.co.jp/gp/product/B00KYEH7GW?ref_=msw_list_shoveler_media_mangatop_0&storeType=ebooks」のような URL からは「B00KYEH7GW」が帰ります。
func ExtractProductIDFromURL(url string) (string, bool) {
	pattern := regexp.MustCompile(`https://www.amazon.co.jp/.*(dp|gp)/(product/)?([0-9a-zA-Z]+)/?.*`)
	matches := pattern.FindStringSubmatch(url)
	if matches == nil || len(matches) < 4 {
		return "", false
	}
	return matches[3], true
}

// Product は商品情報を表す構造体です。
type Product struct {
	ID       string
	Title    string
	Category string
	Price    int
	Point    int
}

// NewProductFromID は商品 ID から Product を作成します。ID 以外の情報は設定されません。
func NewProductFromID(id string) (Product, error) {
	item := Product{}
	if !IsValidProductID(id) {
		return item, fmt.Errorf("%s is not a valid product id", id)
	}
	item.ID = id
	return item, nil
}

// NewProductFromURL は商品の URL から Product を作成します。ID 以外の情報は設定されません。
func NewProductFromURL(url string) (Product, error) {
	item := Product{}
	var ok bool
	item.ID, ok = ExtractProductIDFromURL(url)
	if !ok {
		return item, fmt.Errorf("failed to extract ProductID from %s", url)
	}
	return item, nil
}

// GetURL は商品の URL を返します。
func (p Product) GetURL() string {
	url, _ := GetProductURL(p.ID)
	return url
}

func (p Product) String() string {
	return fmt.Sprintf(
		"[%s]\n"+
			"Title         : %s\n"+
			"Current Price : %d\n"+
			"Point         : %dpt\n"+
			"URL           : %s\n",
		p.ID, p.Title, p.Price, p.Point, p.GetURL())
}

// Update は商品情報ページにアクセスして Product の内容を更新します。
func (p *Product) Update() error {
	resp, err := http.Get(p.GetURL())
	if err != nil {
		return err
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	titleMatcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Span && scrape.Attr(n, "id") == "ebooksProductTitle" {
			return true
		}
		if n.DataAtom == atom.Span && scrape.Attr(n, "id") == "btAsinTitle" {
			return true
		}
		if n.DataAtom == atom.Span && scrape.Attr(n, "id") == "productTitle" {
			return true
		}
		return false
	}

	categoryMatcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Div && scrape.Attr(n, "id") == "nav-subnav" {
			return true
		}
		return false
	}

	priceMatcher := func(n *html.Node) bool {
		// 画面右側に「Kindle 価格: 値段」というタイプ
		// 例： https://www.amazon.co.jp/dp/B004R9QACC
		if n.DataAtom == atom.Td && scrape.Attr(n, "class") == "a-color-price a-size-medium a-align-bottom" {
			return true
		}
		// 画面右側の「1-Click で今すぐ買う」のボックスに値段表記がない（緑色のボックスの）タイプ
		// 例： https://www.amazon.co.jp/dp/B01GI5F2FS
		if n.DataAtom == atom.B && scrape.Attr(n, "class") == "priceLarge" {
			return true
		}
		if n.DataAtom == atom.Span && scrape.Attr(n, "id") == "priceblock_ourprice" {
			return true
		}
		return false
	}

	pointMatcher := func(n *html.Node) bool {
		// 画面右側に「Kindle 価格: 値段」というタイプ（例： https://www.amazon.co.jp/dp/B004R9QACC ）
		if n.DataAtom == atom.Td && scrape.Attr(n, "class") == "a-color-price a-align-bottom" &&
			n.Parent != nil && n.Parent.DataAtom == atom.Tr && scrape.Attr(n.Parent, "class") == "loyalty-points" &&
			strings.Contains(scrape.Text(n), "pt") {
			return true
		}
		// 画面右側の「1-Click で今すぐ買う」のボックスに値段表記がない（緑色のボックスの）タイプ（例： https://www.amazon.co.jp/dp/B01GI5F2FS/ ）
		if n.DataAtom == atom.Td && scrape.Attr(n, "class") == "price" &&
			strings.Contains(scrape.Text(n), "pt") {
			return true
		}
		return false
	}

	titleNode, ok := scrape.Find(root, titleMatcher)
	if !ok {
		return fmt.Errorf("titleNode for %s was not found", p.GetURL())
	}
	p.Title = scrape.Text(titleNode)

	categoryNode, ok := scrape.Find(root, categoryMatcher)
	if ok {
		// カテゴリが見つからないことは許容する。
		p.Category = scrape.Attr(categoryNode, "data-category")
	}

	priceNode, ok := scrape.Find(root, priceMatcher)
	if !ok {
		return fmt.Errorf("priceNode for %s was not found", p.GetURL())
	}
	// 値段のノードに "\ XXの割引（N%)" のような表記が <p> タグで含まれる場合があるため、直下のテキストだけ取得するように TextJoin を使う。
	priceText := scrape.TextJoin(priceNode, func(s []string) string { return s[0] })
	// 数値以外（\記号や , など）を削除。
	re := regexp.MustCompile(`[^0-9]`)
	price, err := strconv.Atoi(re.ReplaceAllString(priceText, ""))
	if err != nil {
		return fmt.Errorf("priceNode for %s was found, but had unexpected text format : %s", p.GetURL(), priceText)
	}
	p.Price = price

	if point, ok := scrape.Find(root, pointMatcher); ok {
		pointText := scrape.Text(point)
		ptIndex := strings.Index(pointText, "pt")
		if ptIndex < 0 {
			return fmt.Errorf("pointNode for %s was found, but had unexpected text format : %s", p.GetURL(), pointText)
		}
		point, err := strconv.Atoi(pointText[:ptIndex])
		if err != nil {
			return fmt.Errorf("pointNode for %s was found, but had unexpected text format : %s", p.GetURL(), pointText)
		}
		p.Point = point
	}

	return nil
}

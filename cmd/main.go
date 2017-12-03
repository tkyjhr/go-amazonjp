package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pkg/browser"
	"github.com/tkyjhr/go-amazonjp"
)

const (
	defaultProductsFile = "products.json"
)

type productWithNotification struct {
	amazonjp.Product
	NotifyPrice int
}

func (p productWithNotification) shouldNotify() bool {
	if p.NotifyPrice >= 0 && p.NotifyPrice >= p.Price-p.Point {
		return true
	}
	return false
}

func readProductsJson(filePath string) ([]productWithNotification, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	products := make([]productWithNotification, 0)
	err = json.NewDecoder(f).Decode(&products)
	return products, err
}

func exitOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func writePrettyJson(filePath string, v interface{}) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	return enc.Encode(v)
}

func addProductCommand(args []string) {
	var productsFile, productID string
	var price int
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	fs.StringVar(&productsFile, "file", defaultProductsFile, "")
	fs.StringVar(&productID, "id", "", "")
	fs.IntVar(&price, "price", -1, "")
	fs.Parse(args)
	if productsFile == "" {
		log.Fatal("-file must be specified.")
	}

	if productID == "" {
		log.Fatal("-id must be specified.")
	}

	products, err := readProductsJson(productsFile)
	if err != nil && os.IsNotExist(err) {
		log.Fatal(err)
	}
	for _, p := range products {
		if p.ID == productID {
			log.Fatalf("%s (%s) is already in the file.", p.ID, p.Title)
		}
	}

	p, err := amazonjp.NewProductFromID(productID)
	exitOnErr(err)
	exitOnErr(p.Update(http.DefaultClient))

	products = append(products, productWithNotification{p, price})

	exitOnErr(writePrettyJson(productsFile, products))

	log.Printf("Succeeded to add \"%s\".\n", p.Title)
}

func checkProductsCommand(args []string) {
	var productsFile string
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	fs.StringVar(&productsFile, "file", defaultProductsFile, "")
	fs.Usage = func() {
		fs.PrintDefaults()
	}
	fs.Parse(args)

	products, err := readProductsJson(productsFile)
	exitOnErr(err)
	for _, p := range products {
		log.Printf("Checking \"%s\" ...\n", p.Title)
		if err := p.Update(http.DefaultClient); err != nil {
			log.Printf("Failed to update \"%s\" (%s)\n", p.Title, p.GetURL())
			continue
		}
		if p.shouldNotify() {
			browser.OpenURL(p.GetURL())
		}
	}
	exitOnErr(writePrettyJson(productsFile, products))
}

const (
	mainUsage = `Usage: <Command> [Options]

Command:
	add -id <id> -price <price> [-file <file>]
	check [-file <file>]
`
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Printf("Error : No command\n")
		fmt.Println(mainUsage)
		os.Exit(1)
	}
	command := args[0]
	subargs := args[1:]
	switch command {
	case "add":
		addProductCommand(subargs)
	case "check":
		checkProductsCommand(subargs)
	default:
		fmt.Printf("Error : Unknown command : %s\n", args[0])
		fmt.Println(mainUsage)
		os.Exit(1)
	}
}

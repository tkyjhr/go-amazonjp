package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"

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
	var productsFile, productID, productURL string
	var price int
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	fs.StringVar(&productsFile, "file", defaultProductsFile, "The file that list the products. Create a new file if the specified file does not exist.")
	fs.StringVar(&productID, "id", "", "ID of the product.")
	fs.StringVar(&productURL, "url", "", "URL of the product.")
	fs.IntVar(&price, "price", -1, "The price that you want to be notified when running 'check' command. If not specified, set to -1 (never notify).")
	fs.Parse(args)
	if productsFile == "" {
		log.Fatal("-file must be specified.")
	}

	if productID == "" && productURL == "" {
		log.Fatal("-id or -url must be specified.")
	}
	if productID != "" && productURL != "" {
		log.Fatal("-id and -url must not be specified at once. Use only one of them.")
	}
	if productURL != "" {
		var ok bool
		productID, ok = amazonjp.ExtractProductIDFromURL(productURL)
		if !ok {
			log.Fatalf("failed to extract product id from %s", productURL)
		}
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

func countNonEmptyString(ss ...string) int {
	count := 0
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			count++
		}
	}
	return count
}

func sendMail(from, password, to, subject, message string) error {
	auth := smtp.PlainAuth(
		"",
		from,
		password,
		"smtp.gmail.com",
	)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		from,
		[]string{to},
		[]byte(
			fmt.Sprintf(`To: %s
Subject: %s

%s
`, to, subject, message)))
}

func checkProductsCommand(args []string) {
	var productsFile, from, password, to string

	fs := flag.NewFlagSet("check", flag.ExitOnError)
	fs.StringVar(&productsFile, "file", defaultProductsFile, "The file that listed products to check. Use 'add' command to create.")
	fs.StringVar(&from, "from", "", "email address from which send a notification. Must be gmail.")
	fs.StringVar(&password, "password", "", "password of the email address specified in -from")
	fs.StringVar(&to, "to", "", "email address to which send a notification. ")
	fs.Usage = func() {
		fs.PrintDefaults()
	}
	fs.Parse(args)

	notifyByEmail := false
	switch countNonEmptyString(from, password, to) {
	case 0:
		break
	case 3:
		notifyByEmail = true
	default:
		log.Fatal("-from, -password and -to must be specified at once.")
	}

	products, err := readProductsJson(productsFile)
	exitOnErr(err)
	var notifyProducts []productWithNotification
	for _, p := range products {
		log.Printf("Checking \"%s\" ...\n", p.Title)
		if err := p.Update(http.DefaultClient); err != nil {
			log.Printf("Failed to update \"%s\" (%s)\n", p.Title, p.GetURL())
			continue
		}
		if p.shouldNotify() {
			notifyProducts = append(notifyProducts, p)
		}
	}
	if len(notifyProducts) > 0 {
		if notifyByEmail {
			msg := ""
			for _, p := range notifyProducts {
				msg += fmt.Sprintf("%s is now \\%d (%d pt)\n%s\n\n", p.Title, p.Price, p.Point, p.GetURL())
			}
			exitOnErr(sendMail(from, password, to, "Amazon Low Price Notification", msg))
		} else {
			for _, p := range notifyProducts {
				browser.OpenURL(p.GetURL())
			}
		}
	}
	exitOnErr(writePrettyJson(productsFile, products))
}

const (
	mainUsage = `Usage: <Command> [Options]

Command:
	add [-id <id>|-url <url>] [-price <price>] [-file <file>]
	check [-file <file>] [-from <mail-address(must be gmail)> -password <password> -to <mail-address>]
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
		fmt.Printf("Error : Unknown command : %s\n", command)
		fmt.Println(mainUsage)
		os.Exit(1)
	}
}

package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Item struct {
	Title     string `json:"title"`
	Condition string `json:"condition"`
	Price     string `json:"price"`
	Url       string `json:"product_url"`
	id        string
}

func Start(url string, conditionFilterString string) {
	var wg sync.WaitGroup
	var ok bool = true

	if err := os.Mkdir("data", os.ModePerm); err != nil {
		fmt.Println(err)
	}

	for ok {
		responce, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
		}

		if responce.StatusCode > 400 {
			fmt.Println("error code: ", responce.StatusCode)
		}

		defer responce.Body.Close()
		doc, err := goquery.NewDocumentFromReader(responce.Body)
		if err != nil {
			fmt.Println(err)
		}

		url, ok = doc.Find("a.pagination__next").Attr("href")

		wg.Add(1)
		go parseAndWriteItemInfo(doc, &wg, conditionFilterString)
	}

	wg.Wait()
}

func parseAndWriteItemInfo(doc *goquery.Document, wg *sync.WaitGroup, conditionFilterString string) {
	defer wg.Done()
	items := scrapePageData(doc, conditionFilterString)
	writeItemToJsonFile(&items)
}

func scrapePageData(doc *goquery.Document, conditionFilterString string) []Item {
	var items []Item
	doc.Find("ul.srp-results>li.s-item").Each(func(index int, item *goquery.Selection) {
		title := item.Find("div.s-item__title>span")
		price := item.Find("div.s-item__detail>span.s-item__price")
		itemLink, _ := item.Find("a.s-item__link").Attr("href")
		condition := item.Find("span.SECONDARY_INFO")

		if filterCondition(condition.Text(), conditionFilterString) {
			i := Item{
				Title:     title.Text(),
				Condition: condition.Text(),
				Price:     strings.Trim(price.Text(), "$"),
				Url:       itemLink,
				id:        extractItemIdFromUrl(itemLink),
			}

			items = append(items, i)
		}
	})

	return items
}

func writeItemToJsonFile(items *[]Item) {
	for _, item := range *items {
		s, err := json.Marshal(item)
		if err != nil {
			fmt.Println(err)
		}

		fileName := fmt.Sprintf("data/%s.json", item.id)

		f, err := os.Create(fileName)
		if err != nil {
			fmt.Println(err)
		}

		_, err = f.Write(s)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func filterCondition(condition, filter string) bool {
	if filter == "" || filter == "all" {
		return true
	}

	return condition == filter
}

func extractItemIdFromUrl(url string) string {
	idString := strings.Split(url, "?")
	id := idString[0][25:]

	return id
}

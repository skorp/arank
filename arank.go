package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type items struct {
	index    int
	rank     int
	url      string
	response *http.Response
	err      error
}

func getRanks(urls []string) <-chan items {
	ch := make(chan items, len(urls)) // buffered
	for i, url := range urls {

		go func(url string, i int) {
			alexaurl := "https://www.alexa.com/minisiteinfo/" + url
			resp, err := http.Get(alexaurl)
			defer resp.Body.Close()

			lastrank := 0
			if resp.StatusCode == 200 {
				doc, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				doc.Find("table#siteStats").Each(func(i int, s *goquery.Selection) {

					rank := s.Find("tr").First().Find("a").First().Text()
					if endrank, err := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(rank, ",", ""))); err == nil {
						lastrank = endrank
					}
				})
			}

			ch <- items{i, lastrank, url, resp, err}
		}(url, i)
	}
	return ch
}

func main() {

	urls := os.Args[1:]
	var output []items
	results := getRanks(urls)
	for _ = range urls {
		res := <-results
		if res.rank > 0 {
			output = append(output, res)
		}
		if res.err != nil {
			fmt.Println(res.url, " error")
		}
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].rank < output[j].rank
	})

	for _, element := range output {
		fmt.Println(element.url, "=>", element.rank)
	}
}

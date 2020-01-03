package ytn

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type LastIndex struct {
	Ytn IndexSuffix `json:"ytn"`
}
type IndexSuffix struct {
	Index  []int    `json:"index"`
	Suffix []string `json:"suffix"`
}

const (
	YTN_URL  = "https://radio.ytn.co.kr/program/"
	filePath = "scripts/ytn/"
)

func Crawler(currentTime time.Time) {

	var wait sync.WaitGroup
	wait.Add(3)

	// Parsing with structs
	var li LastIndex
	parseLastIndex(&li)

	for i := 0; i < len(li.Ytn.Index); i++ {
		go func(iAddr *int, lastIndex int, suffix string) {
			latestIndex, contentUrls := getAttrHrefsAfterLastIndex(YTN_URL+suffix, lastIndex)

			contents := make([]string, 0)
			for _, u := range contentUrls {
				content := getContent(u)
				contents = append(contents, content)
			}

			if len(contents) != 0 {
				writeFile(filePath, getFileName(currentTime, lastIndex, latestIndex), strings.Join(contents, "\n"))

				*iAddr = latestIndex
			}

			defer wait.Done()
		}(&li.Ytn.Index[i], li.Ytn.Index[i], li.Ytn.Suffix[i])
	}

	wait.Wait()

	updateLastIndex(li)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getFileName(currentT time.Time, lastI int, latestI int) string {
	return currentT.Format("[2006.01.02]") + strconv.Itoa(lastI) + "-" + strconv.Itoa(latestI) + ".txt"
}

func parseLastIndex(li *LastIndex) {
	// Read json file
	b, err := ioutil.ReadFile("src/lastIndex.json")
	check(err)

	// Parsing with structs
	if err := json.Unmarshal(b, li); err != nil {
		check(err)
	}
}

func updateLastIndex(li LastIndex) {
	bJson, _ := json.Marshal(li)
	if err := ioutil.WriteFile("src/lastIndex.json", bJson, os.ModePerm); err != nil {
		check(err)
	}
}

func writeFile(path string, fileName string, content string) {
	f, err := os.Create(path + fileName)
	check(err)

	defer f.Close()

	_, err = f.WriteString(content)
	check(err)

	f.Sync()
}

func getContent(url string) string {
	doc := getDocument(url)

	// find content text
	return doc.Find(".gray_table_view > tbody").Text()
}

func getDocument(url string) *goquery.Document {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

// func getAttrHrefs(url string) []string {
// 	// Get document from body
// 	doc := getDocument(url)

// 	hrefs := make([]string, 0)

// 	// Find the tr tags
// 	doc.Find(".gray_table > tbody >  tr").Each(func(i int, s *goquery.Selection) {
// 		// Filter notice
// 		if s.Find("td").First().HasClass("default") {
// 			// Get a tag's href
// 			href, exists := s.Find("a").Attr("href")

// 			if exists {
// 				hrefs = append(hrefs, YTN_URL+href)
// 			}
// 		}
// 	})

// 	return hrefs
// }

func getAttrHrefsAfterLastIndex(url string, lastIndex int) (int, []string) {
	// Get document from body
	doc := getDocument(url)

	hrefs := make([]string, 0)

	latestNum := 0

	// Find the tr tags
	doc.Find(".gray_table > tbody >  tr").Each(func(i int, s *goquery.Selection) {
		// Filter tr tags after last index
		contentNum := s.Find("td").First().Text()
		num, err := strconv.Atoi(contentNum)
		// Exception handling 공지사항
		if err == nil && num > lastIndex {
			// Update lastIndex for endpoint
			if latestNum < num {
				latestNum = num
			}

			// Get a tag's href
			href, exists := s.Find("a").Attr("href")

			if exists {
				hrefs = append(hrefs, YTN_URL+href)
			}
		}
	})

	return latestNum, hrefs
}

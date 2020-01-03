package imbc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type LastIndex struct {
	Imbc IndexSuffix `json:"imbc"`
}
type IndexSuffix struct {
	Index  []int    `json:"index"`
	Suffix []string `json:"suffix"`
}

const (
	IMBC_URL = "http://www.imbc.com/broad/radio/fm/"
	filePath = "scripts/ytn/"
)

func Crawler(currentTime time.Time) {

	// Parsing with structs
	var li LastIndex
	parseLastIndex(&li)

	contextVar, cancelFunc := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancelFunc()

	contextVar, cancelFunc = context.WithTimeout(contextVar, 10*time.Second)
	defer cancelFunc()

	var iframeUrl string
	var exists bool

	if err := chromedp.Run(contextVar,
		chromedp.Navigate("http://www.imbc.com/broad/radio/fm/look/interview/"),
		chromedp.WaitVisible("#imbc_content"),
		chromedp.AttributeValue("div.sub-content > iframe", "src", &iframeUrl, &exists),
	); err != nil {
		panic(err)
	}

	if !exists {
		log.Fatal("No iframe!!")
	}

	var hrefs []*cdp.Node
	if err := chromedp.Run(contextVar,
		chromedp.Navigate(iframeUrl),
		chromedp.Nodes("span[id$=lblListNum] > a.bbs_list_text_title", &hrefs),
	); err != nil {
		panic(err)
	}

	res := make([]string, len(hrefs))
	for i := 0; i < len(hrefs); i++ {
		fmt.Println(hrefs[i])
		if err := chromedp.Run(contextVar,
			chromedp.MouseClickNode(hrefs[i]),
			chromedp.Text("div.bbs_view_content", &res[i]),
			chromedp.NavigateBack(),
			chromedp.Nodes("span[id$=lblListNum] > a.bbs_list_text_title", &hrefs),
		); err != nil {
			panic(err)
		}
	}

	// fmt.Println(strings.Join(res, ""))

}

func check(e error) {
	if e != nil {
		panic(e)
	}
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

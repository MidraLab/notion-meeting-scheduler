package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type notionResponse struct {
	Results []struct {
		ID string `json:"id"`
	} `json:"Results"`
}

const daysUntilNextThursday = 11

func main() {
	c := make(chan string)
	go readPageID(c)
	patchPageTitle(<-c)
}

func readPageID(c chan string) {
	notionDatabaseURL := loadEnv("NOTION_DATABASE_URL")
	url := "https://api.notion.com/v1/databases/" + notionDatabaseURL + "/query"

	payload := strings.NewReader(`{
    "filter": {
        "property": "会議種別",
        "multi_select": {
            "contains": "定例"
        }
    },
    "page_size": 1
}`)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+loadEnv("NOTION_API"))
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Add("content-type", "application/json")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	var notionRes notionResponse
	if err := json.NewDecoder(res.Body).Decode(&notionRes); err != nil {
		log.Fatal(err)
	}
	pageID := notionRes.Results[0].ID

	c <- pageID
}

func patchPageTitle(id string) {
	url := "https://api.notion.com/v1/pages/" + id

	nextThursday := time.Now().AddDate(0, 0, (daysUntilNextThursday-int(time.Now().Weekday()))%7)
	nextThursdayStartStr := nextThursday.Format("2006-01-02")
	nextThursdayTitleStr := nextThursday.Format("01/02")

	payload := strings.NewReader(fmt.Sprintf(`{
    "properties": {
        "名前": {
            "title": [
                {
                    "text": {
                        "content": "定例%s"
                    }
                }
            ]
        },
		"会議開催日": {
            "date": {
   				 "start": "%s",
   				 "end": null
  			}
        }
    }
}`, nextThursdayTitleStr, nextThursdayStartStr))

	req, _ := http.NewRequest("PATCH", url, payload)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+loadEnv("NOTION_API"))
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Add("content-type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))
}

func loadEnv(keyName string) string {
	err := godotenv.Load(".env")
	// もし err がnilではないなら、"読み込み出来ませんでした"が出力されます。
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}
	// .envの SAMPLE_MESSAGEを取得して、messageに代入します。
	message := os.Getenv(keyName)

	return message
}

package seoultechbot

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func GetWebData(url string) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("error getting response from website,", err)
		return
	}
	defer response.Body.Close()
	html, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		fmt.Println("error reading website,", err)
		return
	}
	fmt.Print(html.Text())
}

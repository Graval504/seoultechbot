package seoultechbot

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func FindIndex(arr [25]string, value interface{}) (found bool, index int) {
	for i, v := range arr {
		if v == value {
			return true, i
		}
	}
	return false, -1
}

func GetWebInfo(url string) (WebInfo *goquery.Document, err error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	html, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}
	return html, nil
}

/*
func HtmlToImage(html *goquery.Selection) (image []byte, occuredError error) {
	var buf []byte
	htmlCode, htmlerr := html.Html()
	if htmlerr != nil {
		return buf, htmlerr
	}
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	if err := chromedp.Run(ctx, FullScreenshot(`data:text/html,`+`<html lang="ko">`+htmlCode, 90, &buf)); err != nil {
		return nil, err
	}
	return buf, nil
}

func FullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(res, quality),
	}
}
*/

func SaveImageFile(image []byte, name string) (err error) {
	if err := os.WriteFile(name+".png", image, 0o644); err != nil {
		fmt.Println("error writing file,", err)
		return err
	}
	fmt.Println("wrote " + name + ".png")
	return nil
}

func ContentsToImage(url string, selector string) (image []byte, occuredError error) {
	var buf []byte
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector, chromedp.ByID),
		chromedp.Screenshot(selector, &buf, chromedp.ByID),
	); err != nil {
		return nil, err
	}
	return buf, nil
}

func FullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(res, quality),
	}
}

package seoultechbot

import (
	"context"
	"encoding/base64"
	"io"
	"log"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

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

func SaveImageFile(image io.Reader, name string) (err error) {
	file, err := os.Create(name + ".png")
	if err != nil {
		log.Println("error creating file,", err)
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, image)
	if err != nil {
		log.Println("error saving image,", err)
		return err
	}
	log.Println("wrote " + name + ".png")
	return nil
}

func ContentsToImage(url string, selector string) (image []byte, occuredError error) {
	var buf []byte
	ctx, cancel := chromedp.NewContext(context.Background())
	var nodes []*cdp.Node
	defer cancel()
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector, chromedp.ByID),
		chromedp.Nodes(selector, &nodes, chromedp.ByID),
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

func ImageToUrl(img []byte) (imageUrl string) {
	imgBase64 := base64.StdEncoding.EncodeToString(img)
	return "data:image/png;base64," + imgBase64
}

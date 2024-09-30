package seoultechbot

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const AAI string = "https://aai.seoultech.ac.kr/information/bulletin/"
const COSS string = "https://coss.seoultech.ac.kr/community/notice/"
const SEOULTECH string = "https://www.seoultech.ac.kr/service/info/notice/"
const LISTSIZE int = 50

func Scrap(url string) (isUpdated bool, bulletinList []bulletin, err error) {
	html, err := GetWebInfo(url)
	if err != nil {
		log.Println("error scraping web,", err)
		return false, nil, err
	}
	var titlelist [LISTSIZE]string
	var urllist [LISTSIZE]string
	html.Find(DecideTitleSelector(url)).Each(
		func(i int, s *goquery.Selection) {
			titlelist[i], urllist[i] = strings.TrimSpace(s.Text()), s.AttrOr("href", "None")
		})
	isUpdated, newTitleList := TitleList.CheckWebUpdate(titlelist, url)
	if !isUpdated {
		return false, nil, nil
	}
	for _, title := range newTitleList {
		found, index := FindIndex(titlelist, title)
		if found {
			image, err := GetNoticeContents(url, urllist[index])
			if err != nil {
				return true, nil, err
			}
			bulletinList = append(bulletinList, bulletin{url + urllist[index], title, image})
		}
	}
	return true, bulletinList, nil
}

func DecideTitleSelector(url string) string {
	switch url {
	case COSS:
		return "#sub > div > div.board_container > table > tbody > tr > td.body_col_title.dn2 > div:nth-child(1) > a"
	case AAI:
		return "#sub > div > div.board_container > table > tbody > tr > td.body_col_title.dn2 > div:nth-child(1) > a"
	case SEOULTECH:
		return "#hcms_content > div.wrap_list > table > tbody > tr > td.tit.dn2 > a"
	default:
		return "error"
	}
}

func DecideContentsSelector(url string) string {
	switch url {
	case COSS:
		return "#sub > div > div.board_container > div > table > tbody > tr:nth-child(4)"
	case AAI:
		return "#sub > div > div.board_container > div > table > tbody > tr:nth-child(4)"
	case SEOULTECH:
		return "#hcms_content > div.wrap_view > table > tbody > tr:nth-child(4)"
	default:
		return "error"
	}
}

type formertitlelist struct {
	COSSTitleList      [LISTSIZE]string
	AAITitleList       [LISTSIZE]string
	SeoulTechTitleList [LISTSIZE]string
}

var TitleList formertitlelist

func init() {
	TitleList = formertitlelist{}
	TitleList.LoadFormerTitles()
}

func (t *formertitlelist) CheckWebUpdate(currentTitles [LISTSIZE]string, url string) (isUpdated bool, updatedTitles []string) {
	var formerTitles *[LISTSIZE]string
	var found bool
	newTitles := []string{}
	switch url {
	case COSS:
		formerTitles = &t.COSSTitleList
	case AAI:
		formerTitles = &t.AAITitleList
	case SEOULTECH:
		formerTitles = &t.SeoulTechTitleList
	default:
		return false, newTitles
	}
	for _, currentTitle := range currentTitles {
		found = false
		for _, formerTitle := range *formerTitles {
			if strings.Compare(currentTitle, formerTitle) != 0 {
				continue
			} else {
				found = true
				break
			}
		}
		if !found {
			newTitles = append(newTitles, currentTitle)
		}
	}
	for i := range formerTitles {
		(*formerTitles)[i] = currentTitles[i]
	}
	if len(newTitles) == 0 {
		return false, newTitles
	}
	return true, newTitles
}

func GetNoticeContents(url string, contentsUrl string) (image []byte, err error) {
	image, err = ContentsToImage(url+contentsUrl, DecideContentsSelector(url))
	if err != nil {
		log.Println("error converting html into image,", err)
		return image, err
	}
	return image, nil
}

func FindIndex(arr [LISTSIZE]string, value interface{}) (found bool, index int) {
	for i, v := range arr {
		if v == value {
			return true, i
		}
	}
	return false, -1
}

func GetWebInfo(url string) (WebInfo *goquery.Document, err error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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

type bulletin struct {
	Url   string
	Title string
	Image []byte
}

func (list formertitlelist) SaveFormerTitles() error {
	file, err := json.Marshal(list)
	if err != nil {
		log.Println("error converting into json,", err)
		return err
	}
	err = os.WriteFile("./formerTitleList.json", file, os.FileMode(0644))
	if err != nil {
		log.Println("error writing file,", err)
		return err
	}
	return nil
}

func (list *formertitlelist) LoadFormerTitles() error {
	file, err := os.ReadFile("./formerTitleList.json")
	if err != nil {
		log.Println("error reading file,", err)
		return err
	}
	err = json.Unmarshal(file, &list)
	if err != nil {
		log.Println("error converting into struct,", err)
		return err
	}
	return nil
}

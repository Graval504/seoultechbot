package seoultechbot

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const AAI string = "https://aai.seoultech.ac.kr/information/bulletin/"
const COSS string = "https://coss.seoultech.ac.kr/community/notice/"
const SEOULTECH string = "https://www.seoultech.ac.kr/service/info/notice/"

func Scrap(url string) (isUpdated bool, newUrlList []string, newTitleList []string, imageList [][]byte, err error) {
	html, err := GetWebInfo(url)
	if err != nil {
		fmt.Println("error scraping web,", err)
		return false, nil, nil, nil, err
	}
	var titlelist [25]string
	var urllist [25]string
	html.Find(DecideTitleSelector(url)).Each(
		func(i int, s *goquery.Selection) {
			titlelist[i], urllist[i] = strings.TrimSpace(s.Text()), s.AttrOr("href", "None")
		})
	isUpdated, newTitleList = TitleList.CheckWebUpdate(titlelist, url)
	if !isUpdated {
		return false, nil, nil, nil, nil
	}
	for _, title := range newTitleList {
		found, index := FindIndex(titlelist, title)
		if found {
			image, err := GetNoticeContents(url, urllist[index])
			if err != nil {
				return true, nil, nil, nil, err
			}
			imageList = append(imageList, image)
			newUrlList = append(newUrlList, urllist[index])
		}
	}
	return true, newUrlList, newTitleList, imageList, nil
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
	COSSTitleList      [25]string
	AAITitleList       [25]string
	SeoulTechTitleList [25]string
}

var TitleList formertitlelist

func Init() {
	TitleList = formertitlelist{}
}

func (t formertitlelist) CheckWebUpdate(currentTitles [25]string, url string) (isUpdated bool, updatedTitles []string) {
	var formerTitles *[25]string
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
		for _, formerTitle := range formerTitles {
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
	if len(newTitles) == 0 {
		return false, newTitles
	}
	return true, newTitles
}

func GetNoticeContents(url string, contentsUrl string) (image []byte, err error) {
	image, err = ContentsToImage(url+contentsUrl, DecideContentsSelector(url))
	if err != nil {
		fmt.Println("error converting html into image,", err)
		return image, err
	}
	return image, nil
}

package parsers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/theverysameliquidsnake/sales-bot/internal/types"
)

func ParseBackloggdWishlist(profileUrl string) ([]string, error) {
	// Obtain wishlist link
	doc, err := getGoqueryDoc(profileUrl)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("parser: could not get response from: %s", profileUrl), err)
	}

	gamesUrl, isExists := doc.Find("a[href^='/u/'][href$='/games/']").Attr("href")
	if !isExists {
		return nil, fmt.Errorf("parser: could not find games url: %s", profileUrl)
	}

	gamesUrl, err = resolvePartialUrl(gamesUrl)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("parser: could not resolve partial url: %s", gamesUrl), err)
	}

	doc, err = getGoqueryDoc(gamesUrl)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("parser: could not get response from: %s", gamesUrl), err)
	}

	wishlistUrl, isExists := doc.Find("a[href^='/u/'][href$='/type:wishlist/']").Attr("href")
	if !isExists {
		return nil, fmt.Errorf("parser: could not find wishlist url: %s", gamesUrl)
	}

	wishlistUrl, err = resolvePartialUrl(wishlistUrl)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("parser: could not resolve partial url: %s", wishlistUrl), err)
	}

	// Collect pages url to parse for games
	pagesUrl := types.NewSet()
	//pagesUrl.Add(wishlistUrl)

	doc, err = getGoqueryDoc(wishlistUrl)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("parser: could not get response from: %s", wishlistUrl), err)
	}

	doc.Find("nav[aria-label='Pages'] > a[href^='/page=']").Each(func(i int, s *goquery.Selection) {
		pageUrl, isExists := s.Attr("href")
		if isExists {
			pagesUrl.Add(pageUrl)
		}
	})

	// Parse each page and collect game slugs
	slugs := types.NewSet()

	doc.Find("div[id='game-lists'] a[href^='/games/']").Each(func(i int, s *goquery.Selection) {
		gameUrl, isExists := s.Attr("href")
		if isExists {
			slugs.Add(strings.Split(gameUrl, "/")[2])
		}
	})

	for _, pageUrl := range pagesUrl.Values() {
		pageUrl, err = resolvePartialUrl(wishlistUrl)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("parser: could not resolve partial url: %s", pageUrl), err)
		}

		doc, err = getGoqueryDoc(pageUrl)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("parser: could not get response from: %s", pageUrl), err)
		}

		doc.Find("div[id='game-lists'] a[href^='/games/']").Each(func(i int, s *goquery.Selection) {
			gameUrl, isExists := s.Attr("href")
			if isExists {
				slugs.Add(strings.Split(gameUrl, "/")[2])
			}
		})
	}

	return slugs.Values(), nil
}

func getGoqueryDoc(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("request error: %d %s", res.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func resolvePartialUrl(partialUrl string) (string, error) {
	fullUrl, err := url.JoinPath("https://backloggd.com/", partialUrl)
	if err != nil {
		return "", err
	}

	return fullUrl, nil
}

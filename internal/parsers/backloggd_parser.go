package parsers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/playwright-community/playwright-go"
	"github.com/theverysameliquidsnake/sales-bot/internal/configs"
	"github.com/theverysameliquidsnake/sales-bot/internal/types"
)

func ParseBackloggdWishlistPlaywright(profileUrl string) ([]string, error) {
	browser := configs.GetBrowser()

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36') Chrome/85.0.4183.121 Safari/537.36"
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{UserAgent: &userAgent})
	if err != nil {
		return nil, fmt.Errorf("parser: could not create new browser page: %w", err)
	}
	defer page.Close()

	if _, err = page.Goto(profileUrl, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
		return nil, fmt.Errorf("parser: could not go to profile url: %w", err)
	}

	gamesUrl, err := page.Locator("a[href^='/u/'][href$='/games/']").First().GetAttribute("href")
	if err != nil {
		return nil, fmt.Errorf("parser: could not find games url: %w", err)
	}

	gamesUrl, err = resolvePartialUrl(gamesUrl)
	if err != nil {
		return nil, fmt.Errorf("parser: could not resolve partial url: %w", err)
	}

	if _, err = page.Goto(gamesUrl, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
		return nil, fmt.Errorf("parser: could not go to games url: %w", err)
	}

	wishlistUrl, err := page.Locator("a[href^='/u/'][href$='/type:wishlist/']").First().GetAttribute("href")
	if err != nil {
		return nil, fmt.Errorf("parser: could not find wishlist url: %w", err)
	}

	wishlistUrl, err = resolvePartialUrl(wishlistUrl)
	if err != nil {
		return nil, fmt.Errorf("parser: could not resolve partial url: %w", err)
	}

	if _, err = page.Goto(wishlistUrl, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
		return nil, fmt.Errorf("parser: could not go to games url: %w", err)
	}

	slugs := types.NewSet()
	entries, err := page.Locator("div[id='game-lists'] a[href^='/games/']").All()
	if err != nil {
		return nil, fmt.Errorf("parser: could not find games on page: %w", err)
	}

	for _, entry := range entries {
		entryUrl, err := entry.GetAttribute("href")
		if err != nil {
			return nil, fmt.Errorf("parser: could not get entry url: %w", err)
		}

		slugs.Add(strings.Split(entryUrl, "/")[2])
	}

	pagesUrl := types.NewSet()
	entries, err = page.Locator("nav[aria-label='Pages'] > a[href^='/page=']").All()
	if err != nil {
		return nil, fmt.Errorf("parser: could not find pages urls on page: %w", err)
	}

	for _, entry := range entries {
		entryUrl, err := entry.GetAttribute("href")
		if err != nil {
			return nil, fmt.Errorf("parser: could not get next pages url: %w", err)
		}

		pagesUrl.Add(entryUrl)
	}

	for _, pageUrl := range pagesUrl.Values() {
		newPageUrl, err := resolvePartialUrl(pageUrl)
		if err != nil {
			return nil, fmt.Errorf("parser: could not resolve partial url: %w", err)
		}

		if _, err = page.Goto(newPageUrl, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
			return nil, fmt.Errorf("parser: could not go to next pages url: %w", err)
		}

		entries, err := page.Locator("div[id='game-lists'] a[href^='/games/']").All()
		if err != nil {
			return nil, fmt.Errorf("parser: could not find games on page: %w", err)
		}

		for _, entry := range entries {
			entryUrl, err := entry.GetAttribute("href")
			if err != nil {
				return nil, fmt.Errorf("parser: could not get entry url: %w", err)
			}

			slugs.Add(strings.Split(entryUrl, "/")[2])
		}
	}

	return slugs.Values(), nil
}

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

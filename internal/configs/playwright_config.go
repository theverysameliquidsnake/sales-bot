package configs

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

var engine *playwright.Playwright
var browser playwright.Browser

func InstallPlaywright() error {
	if err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}}); err != nil {
		return fmt.Errorf("config: could not install playwright: %w", err)
	}

	return nil
}

func StartPlaywright() error {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("config: could not run engine: %w", err)
	}

	engine = pw

	br, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if err != nil {
		return fmt.Errorf("config: could not run browser: %w", err)
	}

	browser = br

	return nil
}

func StopPlaywright() error {
	if err := browser.Close(); err != nil {
		return fmt.Errorf("config: could not stop browser: %w", err)
	}

	if err := engine.Stop(); err != nil {
		return fmt.Errorf("config: could not stop engine: %w", err)
	}

	return nil
}

func GetBrowser() playwright.Browser {
	return browser
}

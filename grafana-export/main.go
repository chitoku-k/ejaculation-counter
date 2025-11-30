package main

import (
	"log"
	"net/http"
	"os"

	"github.com/playwright-community/playwright-go"
)

func main() {
	if len(os.Args) != 3 {
		println("Usage:")
		println("  " + os.Args[0] + " URL FILENAME")
		os.Exit(1)
	}

	url := os.Args[1]
	filename := os.Args[2]

	if err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true}); err != nil {
		log.Fatalf("Failed to install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Failed to run playwright: %v", err)
	}
	defer func() {
		_ = pw.Stop()
	}()

	browser, err := pw.Firefox.Launch()
	if err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	defer func() {
		_ = browser.Close()
	}()

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Failed to create a new page: %v", err)
	}
	defer func() {
		_ = page.Close()
	}()

	response, err := page.Goto(url)
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}
	if status := response.Status(); status != http.StatusOK {
		log.Fatalf("Failed to load (HTTP %v)", status)
	}

	toggle := page.Locator(`[data-testid="data-testid export as json externally switch"]`)
	if err := toggle.DispatchEvent("click", nil); err != nil {
		log.Fatalf("Failed to check the toggle: %v", err)
	}

	download, err := page.ExpectDownload(func() error {
		button := page.Locator(`[data-testid="data-testid export as json save to file button"]`)
		return button.DispatchEvent("click", nil)
	})
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}
	if err := download.SaveAs(filename); err != nil {
		log.Fatalf("Failed to save: %v", err)
	}

	log.Println("Exported.")
}

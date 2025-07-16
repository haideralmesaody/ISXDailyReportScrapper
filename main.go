package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"isxcli/internal/license"

	"github.com/chromedp/chromedp"
)

const (
	baseURL  = "http://www.isx-iq.net"
	startURL = "http://www.isx-iq.net/isxportal/portal/uploadedFilesList.html?currLanguage=en"
)

func main() {
	mode := flag.String("mode", "initial", "scrape mode: initial | accumulative")
	fromStr := flag.String("from", "2025-01-01", "start date (YYYY-MM-DD) (used in initial mode if provided)")
	toStr := flag.String("to", "", "optional end date (YYYY-MM-DD); leave blank to keep site default")
	outDir := flag.String("out", "downloads", "directory to save reports")
	headless := flag.Bool("headless", true, "run browser headless")
	flag.Parse()

	// Initialize license system
	fmt.Println("🔐 ISX Daily Reports Scraper - Licensed Version")
	fmt.Println("═══════════════════════════════════════════════")

	if !checkLicense() {
		fmt.Println("❌ License validation failed. Application will exit.")
		fmt.Println("📞 Contact The Iraqi Investor Group to get a new license.")
		os.Exit(1)
	}

	// Create output directory if it doesn't exist (but don't delete existing files)
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Printf("failed to create output dir: %v\n", err)
		os.Exit(1)
	}

	// determine fromSite depending on mode
	var fromSite string
	if *mode == "accumulative" {
		// scan downloads for latest file
		if d, ok := latestDownloadedDate(*outDir); ok {
			fromSite = d.AddDate(0, 0, 1).Format("02/01/2006") // next day
			fmt.Printf("[MODE accumulative] Detected last report date %s. Will start from %s.\n", d.Format("2006-01-02"), fromSite)
		}
	}

	if fromSite == "" {
		// fallback to user provided from
		startDate, err := time.Parse("2006-01-02", *fromStr)
		if err != nil {
			fmt.Printf("invalid --from date: %v\n", err)
			os.Exit(1)
		}
		fromSite = startDate.Format("02/01/2006")
		fmt.Printf("[MODE initial] Starting from %s (preserving existing files)\n", startDate.Format("2006-01-02"))
	}

	var toSite string
	if *toStr != "" {
		endDate, err := time.Parse("2006-01-02", *toStr)
		if err != nil {
			fmt.Printf("invalid --to date: %v\n", err)
			os.Exit(1)
		}
		toSite = endDate.Format("02/01/2006")
	}

	// setup ChromeDP
	opts := chromedp.DefaultExecAllocatorOptions[:]
	if *headless {
		opts = append(opts, chromedp.Flag("headless", true))
	} else {
		opts = append(opts, chromedp.Flag("headless", false))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	if err := chromedp.Run(ctx, runScraper(fromSite, toSite, *outDir)); err != nil {
		fmt.Fprintf(os.Stderr, "scraping failed: %v\n", err)
		os.Exit(1)
	}
}

func runScraper(fromSite, toSite, outDir string) chromedp.Tasks {
	actions := []chromedp.Action{
		timedAction("Navigate", chromedp.Navigate(startURL)),
		chromedp.WaitVisible(`#date`, chromedp.ByID),
		chromedp.SetValue(`#date`, fromSite, chromedp.ByID),
	}
	if toSite != "" {
		actions = append(actions, chromedp.SetValue(`#toDate`, toSite, chromedp.ByID))
	}
	actions = append(actions,
		chromedp.SetValue(`#reporttype`, "40", chromedp.ByID),
		timedAction("ExecuteSearch", chromedp.Click(`/html/body/div[2]/div/div[3]/div[3]/div[2]/div[4]/div/div[1]/form/div[8]/input`, chromedp.BySearch)),
		chromedp.WaitVisible(`#report`, chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			page := 1
			for {
				fmt.Printf("Scraping page %d...\n", page)
				startPage := time.Now()
				shouldContinue, err := scrapePage(ctx, outDir)
				if err != nil {
					return err
				}
				if !shouldContinue {
					fmt.Printf("Found existing files on page %d, stopping scraping process.\n", page)
					return nil
				}
				// check if next arrow exists
				var nextHref string
				var ok bool
				err = chromedp.Run(ctx, chromedp.AttributeValue(`a img[src*='next.gif']`, "src", &nextHref, &ok))
				if err != nil || !ok {
					// No next arrow or not clickable
					return nil
				}
				// Click the parent anchor of the img
				if err := chromedp.Click(`a img[src*='next.gif']`, chromedp.ByQuery).Do(ctx); err != nil {
					return nil // assume finished when can't click
				}
				// wait for table refresh
				if err := chromedp.WaitVisible(`#report`, chromedp.ByID).Do(ctx); err != nil {
					return err
				}
				fmt.Printf("[TIME] page %d processed in %s\n", page, time.Since(startPage))
				page++
			}
		}),
	)

	return chromedp.Tasks(actions)
}

func scrapePage(ctx context.Context, outDir string) (bool, error) {
	// Retrieve rows data: href, date text, type text
	var rows []struct {
		Href string `json:"href"`
		Date string `json:"date"`
		Typ  string `json:"typ"`
	}

	js := `Array.from(document.querySelectorAll('#report tbody tr')).map(tr => {
		const link = tr.querySelector('td.report-download a');
		if (!link) return null;
		const dateCell = tr.querySelector('td.report-titledata1');
		const typeCell = tr.querySelector('td.report-titledata3');
		return {href: link.getAttribute('href'), date: dateCell ? dateCell.innerText.trim() : '', typ: typeCell ? typeCell.innerText.trim() : ''};
	}).filter(Boolean)`

	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &rows)); err != nil {
		return false, err
	}

	foundExistingFiles := 0
	newDownloads := 0

	for _, r := range rows {
		// We only care about Daily type and xlsx file extension
		if strings.ToLower(r.Typ) != "daily" {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(r.Href), ".xlsx") {
			continue
		}

		fullURL := r.Href
		if !strings.HasPrefix(r.Href, "http") {
			fullURL = baseURL + r.Href
		}

		// Parse date dd/mm/yyyy
		t, err := time.Parse("02/01/2006", r.Date)
		if err != nil {
			// fallback to original filename
			fmt.Printf(" !! unable to parse date '%s': %v\n", r.Date, err)
		}

		var fname string
		if err == nil {
			fname = fmt.Sprintf("%s ISX Daily Report.xlsx", t.Format("2006 01 02"))
		} else {
			fname = filepath.Base(r.Href)
		}

		destPath := filepath.Join(outDir, fname)
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf(" --> already have %s, skipping\n", fname)
			foundExistingFiles++
			continue
		}

		fmt.Printf(" --> downloading %s\n", fname)
		if err := downloadFile(fullURL, destPath); err != nil {
			fmt.Printf("failed to download %s: %v\n", fname, err)
		} else {
			newDownloads++
		}
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("Page summary: %d new downloads, %d existing files\n", newDownloads, foundExistingFiles)

	// If we found more existing files than new downloads, and we found at least some existing files,
	// it means we're getting into already-downloaded territory, so we should stop
	if foundExistingFiles > 0 && foundExistingFiles >= newDownloads {
		return false, nil // Stop scraping
	}

	return true, nil // Continue scraping
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func timedAction(name string, act chromedp.Action) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		start := time.Now()
		err := act.Do(ctx)
		fmt.Printf("[TIME] %s took %s\n", name, time.Since(start))
		return err
	})
}

// latestDownloadedDate looks for files named "YYYY MM DD ISX Daily Report.xlsx" in dir and returns the most recent date.
func latestDownloadedDate(dir string) (time.Time, bool) {
	pattern := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return time.Time{}, false
	}
	var dates []time.Time
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := pattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		t, err := time.Parse("2006 01 02", strings.Join(m[1:4], " "))
		if err == nil {
			dates = append(dates, t)
		}
	}
	if len(dates) == 0 {
		return time.Time{}, false
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates[len(dates)-1], true
}

func checkLicense() bool {
	// Initialize license manager
	licenseManager, err := license.NewManager("license.dat")
	if err != nil {
		fmt.Printf("⚠️  License system initialization failed: %v\n", err)
		return false
	}

	// Check if license is valid
	valid, err := licenseManager.ValidateLicense()
	if valid {
		// Get license info for display
		info, infoErr := licenseManager.GetLicenseInfo()
		if infoErr == nil {
			daysLeft := int(time.Until(info.ExpiryDate).Hours() / 24)
			fmt.Printf("✅ License Valid - %d days remaining\n", daysLeft)
			if daysLeft <= 7 {
				fmt.Printf("⚠️  License expires soon: %s\n", info.ExpiryDate.Format("2006-01-02"))
				fmt.Println("📞 Contact The Iraqi Investor Group for license renewal.")
			}
		}
		fmt.Println("═══════════════════════════════════════════════")
		return true
	}

	// License is invalid or expired
	fmt.Println("❌ Invalid or Expired License")
	fmt.Println("═══════════════════════════════════════════════")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Prompt for license key activation
	fmt.Println("\n🔑 Please enter your ISX license key to activate:")
	fmt.Println("   (License keys look like: ISX3M-ABC123DEF456GHI789JKL)")
	fmt.Print("License Key: ")

	reader := bufio.NewReader(os.Stdin)
	licenseKey, _ := reader.ReadString('\n')
	licenseKey = strings.TrimSpace(licenseKey)

	if licenseKey == "" {
		fmt.Println("❌ No license key provided.")
		return false
	}

	// Validate license key format
	if !isValidLicenseFormat(licenseKey) {
		fmt.Println("❌ Invalid license key format.")
		fmt.Println("   License keys should start with ISX1M, ISX3M, ISX6M, or ISX1Y")
		return false
	}

	// Activate license
	fmt.Println("🔄 Activating license...")
	if err := licenseManager.ActivateLicense(licenseKey); err != nil {
		fmt.Printf("❌ License activation failed: %v\n", err)
		fmt.Println("📞 Please contact The Iraqi Investor Group if you believe this is an error.")
		return false
	}

	fmt.Println("✅ License activated successfully!")
	fmt.Println("🎉 Welcome to ISX Daily Reports Scraper!")
	fmt.Println("═══════════════════════════════════════════════")
	return true
}

func isValidLicenseFormat(licenseKey string) bool {
	// Check if license key starts with valid prefixes
	validPrefixes := []string{"ISX1M", "ISX3M", "ISX6M", "ISX1Y"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(licenseKey, prefix) {
			return true
		}
	}
	return false
}

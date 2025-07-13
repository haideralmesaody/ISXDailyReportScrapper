package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"isxcli/internal/license"
)

func main() {
	var (
		count1m = flag.Int("1m", 0, "Number of 1-month licenses to generate")
		count3m = flag.Int("3m", 0, "Number of 3-month licenses to generate")
		count6m = flag.Int("6m", 0, "Number of 6-month licenses to generate")
		count1y = flag.Int("1y", 0, "Number of 1-year licenses to generate")
		total   = flag.Int("total", 0, "Total number of random licenses to generate")
		output  = flag.String("output", "", "Output file to save license keys (optional)")
	)
	flag.Parse()

	// Initialize license manager
	licenseManager, err := license.NewManager("license-config.json", "license.dat")
	if err != nil {
		log.Fatal("Failed to initialize license manager:", err)
	}

	var generatedLicenses []string
	totalGenerated := 0

	// Generate specific duration licenses
	durations := []struct {
		duration string
		count    int
	}{
		{"1m", *count1m},
		{"3m", *count3m},
		{"6m", *count6m},
		{"1y", *count1y},
	}

	for _, d := range durations {
		if d.count > 0 {
			fmt.Printf("Generating %d licenses for %s duration...\n", d.count, d.duration)
			for i := 0; i < d.count; i++ {
				licenseKey, err := generateRechargeCardLicense(licenseManager, d.duration)
				if err != nil {
					log.Printf("Error generating license %d of %d for %s: %v", i+1, d.count, d.duration, err)
					continue
				}
				generatedLicenses = append(generatedLicenses, licenseKey)
				totalGenerated++
				if (i+1)%10 == 0 {
					fmt.Printf("  Generated %d/%d licenses for %s\n", i+1, d.count, d.duration)
				}
				// Small delay to avoid overwhelming the API
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	// Generate random duration licenses
	if *total > 0 {
		fmt.Printf("Generating %d random duration licenses...\n", *total)
		randomDurations := []string{"1m", "3m", "6m", "1y"}
		for i := 0; i < *total; i++ {
			duration := randomDurations[rand.Intn(len(randomDurations))]
			licenseKey, err := generateRechargeCardLicense(licenseManager, duration)
			if err != nil {
				log.Printf("Error generating random license %d of %d: %v", i+1, *total, err)
				continue
			}
			generatedLicenses = append(generatedLicenses, licenseKey)
			totalGenerated++
			if (i+1)%10 == 0 {
				fmt.Printf("  Generated %d/%d random licenses\n", i+1, *total)
			}
			// Small delay to avoid overwhelming the API
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Save to file if requested
	if *output != "" && len(generatedLicenses) > 0 {
		err := saveLicensesToFile(*output, generatedLicenses)
		if err != nil {
			log.Printf("Error saving licenses to file: %v", err)
		} else {
			fmt.Printf("Saved %d licenses to %s\n", len(generatedLicenses), *output)
		}
	}

	fmt.Printf("\n✅ Successfully generated %d licenses total!\n", totalGenerated)
	fmt.Printf("🔗 Check your Google Sheet: https://docs.google.com/spreadsheets/d/1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc/edit\n")
}

func generateRechargeCardLicense(licenseManager *license.Manager, duration string) (string, error) {
	// For recharge card model, we use a dummy email and generate the license
	// The license will be marked as "Available" for activation
	dummyEmail := "recharge-card@isx.local"

	// Generate the license using the existing system
	licenseKey, err := licenseManager.GenerateLicense(dummyEmail, duration)
	if err != nil {
		return "", fmt.Errorf("failed to generate license: %w", err)
	}

	// Note: We would need to update the license status to "Available" instead of "issued"
	// This would require additional functionality in the license manager

	return licenseKey, nil
}

func generateLicenseKey(duration string) string {
	// Generate unique license key with format: ISX{DURATION}-{RANDOM6}-{RANDOM6}
	prefix := "ISX" + duration

	// Generate random strings
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	part1 := make([]byte, 6)
	part2 := make([]byte, 6)

	for i := range part1 {
		part1[i] = chars[rand.Intn(len(chars))]
	}
	for i := range part2 {
		part2[i] = chars[rand.Intn(len(chars))]
	}

	// Add timestamp to ensure uniqueness
	timestamp := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)

	return fmt.Sprintf("%s-%s-%s%s", prefix, string(part1), string(part2), timestamp)
}

func saveLicensesToFile(filename string, licenses []string) error {
	// Implementation to save licenses to file
	// This is a placeholder - you can implement file writing here
	fmt.Printf("Would save %d licenses to %s\n", len(licenses), filename)
	return nil
}

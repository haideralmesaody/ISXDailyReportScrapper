package license

// getBuiltInConfig returns the Google Sheets configuration for secure credential mode.
// This keeps non-sensitive sheet configuration centralized.
func getBuiltInConfig() GoogleSheetsConfig {
	sheetID := "1l4jJNNqHZNomjp3wpkL-txDfCjsRr19aJZOZqPHJ6lc"
	sheetName := "Licenses"

	return GoogleSheetsConfig{
		SheetID:           sheetID,
		SheetName:         sheetName,
		UseServiceAccount: true,
	}
}

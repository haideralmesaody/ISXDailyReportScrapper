package license

import "strings"

// MaskLicenseKey masks a license key for display (ISX-XXXX-****-****) and stacked keys.
func MaskLicenseKey(key string) string {
	if len(key) < 8 {
		return "****"
	}

	if strings.Contains(key, "+") {
		parts := strings.Split(key, "+")
		masked := ""
		for i, part := range parts {
			if i > 0 {
				masked += "+"
			}
			masked += maskSingleKey(part)
		}
		return masked
	}

	return maskSingleKey(key)
}

// maskSingleKey masks a single license key for display.
func maskSingleKey(key string) string {
	if strings.Contains(key, "-") {
		parts := strings.Split(key, "-")
		if len(parts) >= 2 {
			masked := parts[0] + "-" + parts[1]
			for i := 2; i < len(parts); i++ {
				masked += "-****"
			}
			return masked
		}
	}

	if len(key) > 8 {
		return key[:8] + "****"
	}

	return "****"
}

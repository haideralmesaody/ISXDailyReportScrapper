#!/usr/bin/env python3
"""
Manual RSI Calculation Verification Script

This script manually calculates RSI using Wilder's smoothing method
to verify the accuracy of the API RSI values for BIIB.

RSI Calculation Steps:
1. Calculate price changes (gains/losses)
2. Calculate average gains and losses using Wilder's smoothing
3. Calculate RS (Relative Strength) = Average Gain / Average Loss
4. Calculate RSI = 100 - (100 / (1 + RS))

Wilder's Smoothing:
- First Average Gain/Average Loss = Simple average of first 14 periods
- Subsequent values = (Previous Average * 13 + Current Value) / 14
"""

def calculate_rsi_wilder(prices, period=14):
    """
    Calculate RSI using Wilder's smoothing method

    Args:
        prices: List of closing prices
        period: RSI period (default 14)

    Returns:
        List of RSI values (same length as prices, with None for first period-1 values)
    """
    if len(prices) < period:
        return [None] * len(prices)

    # Calculate price changes
    changes = []
    for i in range(1, len(prices)):
        change = prices[i] - prices[i-1]
        changes.append(change)

    # Initialize RSI list with None values
    rsi_values = [None] + [None] * (period - 1)

    # Calculate first average gain and loss (simple average)
    gains = [change if change > 0 else 0 for change in changes[:period]]
    losses = [-change if change < 0 else 0 for change in changes[:period]]

    avg_gain = sum(gains) / period
    avg_loss = sum(losses) / period

    # Calculate first RSI
    if avg_loss == 0:
        rsi = 100.0
    else:
        rs = avg_gain / avg_loss
        rsi = 100.0 - (100.0 / (1.0 + rs))

    rsi_values.append(rsi)

    # Calculate subsequent RSI values using Wilder's smoothing
    for i in range(period, len(changes)):
        # Current gain and loss
        current_change = changes[i]
        current_gain = current_change if current_change > 0 else 0
        current_loss = -current_change if current_change < 0 else 0

        # Wilder's smoothing: (Previous Average * (period-1) + Current Value) / period
        avg_gain = (avg_gain * (period - 1) + current_gain) / period
        avg_loss = (avg_loss * (period - 1) + current_loss) / period

        # Calculate RSI
        if avg_loss == 0:
            rsi = 100.0
        else:
            rs = avg_gain / avg_loss
            rsi = 100.0 - (100.0 / (1.0 + rs))

        rsi_values.append(rsi)

    return rsi_values

def main():
    # BIIB closing prices from CSV (last 50 records, oldest to newest)
    biib_data = [
        ("2025-08-28", 1.100),
        ("2025-08-29", 1.100),
        ("2025-08-30", 1.100),
        ("2025-08-31", 1.090),
        ("2025-09-01", 1.090),
        ("2025-09-02", 1.100),
        ("2025-09-03", 1.110),
        ("2025-09-04", 1.110),
        ("2025-09-05", 1.110),
        ("2025-09-06", 1.110),
        ("2025-09-07", 1.110),
        ("2025-09-08", 1.090),
        ("2025-09-09", 1.090),
        ("2025-09-10", 1.060),
        ("2025-09-11", 1.070),
        ("2025-09-12", 1.070),
        ("2025-09-13", 1.070),
        ("2025-09-14", 1.080),
        ("2025-09-15", 1.070),
        ("2025-09-16", 1.070),
        ("2025-09-17", 1.060),
        ("2025-09-18", 1.060),
        ("2025-09-19", 1.060),
        ("2025-09-20", 1.060),
        ("2025-09-21", 1.060),
        ("2025-09-22", 1.060),
        ("2025-09-23", 1.050),
        ("2025-09-24", 1.050),
        ("2025-09-25", 1.050),
        ("2025-09-26", 1.050),
        ("2025-09-27", 1.050),
        ("2025-09-28", 1.030),
        ("2025-09-29", 1.030),
        ("2025-09-30", 1.040),
        ("2025-10-01", 1.040),
        ("2025-10-02", 1.040),
        ("2025-10-03", 1.040),
        ("2025-10-04", 1.040),
        ("2025-10-05", 1.040),
        ("2025-10-06", 1.040),
        ("2025-10-07", 1.020),
        ("2025-10-08", 1.030),
        ("2025-10-09", 1.030),
        ("2025-10-10", 1.030),
        ("2025-10-11", 1.030),
        ("2025-10-12", 1.030),
        ("2025-10-13", 1.040),
        ("2025-10-14", 1.020),
        ("2025-10-15", 1.020),
        ("2025-10-16", 1.020),
    ]

    # Extract prices for calculation
    dates = [item[0] for item in biib_data]
    prices = [item[1] for item in biib_data]

    print("=" * 80)
    print("MANUAL RSI CALCULATION VERIFICATION FOR BIIB")
    print("=" * 80)
    print(f"Data Points: {len(prices)} closing prices")
    print(f"Date Range: {dates[0]} to {dates[-1]}")
    print(f"RSI Period: 14 (Wilder's Smoothing)")
    print()

    # Calculate RSI
    rsi_values = calculate_rsi_wilder(prices, period=14)

    # Print detailed results for last 10 days
    print("LAST 10 DAYS - MANUAL RSI CALCULATION:")
    print("-" * 80)
    print(f"{'Date':<12} {'Close':<8} {'Change':<8} {'Manual RSI':<12} {'API RSI':<10} {'Difference':<12}")
    print("-" * 80)

    # API RSI values from our previous curl test (last 10 values)
    api_rsi_values = [
        57.683590929743445,  # Oct 12
        89.64657299465205,   # Oct 13
        53.09452526610562,   # Oct 14
        53.09452526610562,   # Oct 15
        53.09452526610562,   # Oct 16
    ]

    # Calculate differences for last few days
    differences = []
    for i in range(len(prices)):
        if i > 0:
            change = prices[i] - prices[i-1]
        else:
            change = 0.0

        manual_rsi = rsi_values[i] if i < len(rsi_values) else None

        # Get API RSI for comparison (last 5 days only)
        api_rsi = None
        if i >= len(prices) - 5:
            api_index = 4 - (len(prices) - 1 - i)  # Reverse index
            if 0 <= api_index < len(api_rsi_values):
                api_rsi = api_rsi_values[api_index]

        # Calculate difference
        diff = None
        if manual_rsi is not None and api_rsi is not None:
            diff = manual_rsi - api_rsi

        # Store differences for analysis
        if diff is not None:
            differences.append(abs(diff))

        # Only print last 10 days
        if i >= len(prices) - 10:
            date_str = dates[i]
            close_str = f"{prices[i]:.3f}"
            change_str = f"{change:+.3f}"
            manual_str = f"{manual_rsi:.2f}" if manual_rsi is not None else "N/A"
            api_str = f"{api_rsi:.2f}" if api_rsi is not None else "N/A"
            diff_str = f"{diff:+.2f}" if diff is not None else "N/A"

            print(f"{date_str:<12} {close_str:<8} {change_str:<8} {manual_str:<12} {api_str:<10} {diff_str:<12}")

    print("-" * 80)
    print()

    # Analysis
    if differences:
        avg_diff = sum(differences) / len(differences)
        max_diff = max(differences)
        min_diff = min(differences)

        print("DIFFERENCE ANALYSIS (Manual vs API):")
        print("-" * 40)
        print(f"Average Difference: {avg_diff:.2f}")
        print(f"Maximum Difference: {max_diff:.2f}")
        print(f"Minimum Difference: {min_diff:.2f}")
        print()

        if avg_diff < 5.0:
            print("✅ GOOD: Average difference < 5.0 (API values are accurate)")
        elif avg_diff < 10.0:
            print("⚠️  MODERATE: Average difference 5-10 (API values have some variance)")
        else:
            print("❌ POOR: Average difference > 10 (API values may be inaccurate)")
    else:
        print("❌ No differences calculated for comparison")

    print()
    print("RSI RANGE ANALYSIS:")
    print("-" * 30)
    valid_manual_rsi = [r for r in rsi_values if r is not None]
    if valid_manual_rsi:
        print(f"Manual RSI Range: {min(valid_manual_rsi):.2f} - {max(valid_manual_rsi):.2f}")
        print(f"All values in 0-100 range: {all(0 <= r <= 100 for r in valid_manual_rsi)}")

    if api_rsi_values:
        print(f"API RSI Range: {min(api_rsi_values):.2f} - {max(api_rsi_values):.2f}")
        print(f"All values in 0-100 range: {all(0 <= r <= 100 for r in api_rsi_values)}")

if __name__ == "__main__":
    main()
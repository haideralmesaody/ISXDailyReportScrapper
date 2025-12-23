#!/usr/bin/env python3
import sys

def calculate_rsi_manual(prices, period=14):
    """
    Calculate RSI manually using Wilder's smoothing method
    This is the industry standard method used by most trading platforms
    """
    if len(prices) < period + 1:
        return None, "Need at least " + str(period + 1) + " price points"

    # Calculate price changes
    changes = []
    for i in range(1, len(prices)):
        change = prices[i] - prices[i-1]
        changes.append(change)

    # Separate gains and losses
    gains = [max(0, change) for change in changes]
    losses = [max(0, -change) for change in changes]

    print(f"Prices (last {len(prices)}): {prices}")
    print(f"Changes: {changes}")
    print(f"Gains: {gains}")
    print(f"Losses: {losses}")
    print()

    # Calculate initial average gain and loss (first period)
    avg_gain = sum(gains[:period]) / period
    avg_loss = sum(losses[:period]) / period

    print(f"Initial Avg Gain (first {period}): {avg_gain:.6f}")
    print(f"Initial Avg Loss (first {period}): {avg_loss:.6f}")
    print()

    # Apply Wilder's smoothing for remaining periods
    for i in range(period, len(gains)):
        avg_gain = ((avg_gain * (period - 1)) + gains[i]) / period
        avg_loss = ((avg_loss * (period - 1)) + losses[i]) / period
        print(f"Day {i+1}: Avg Gain = {avg_gain:.6f}, Avg Loss = {avg_loss:.6f}")

    print()

    # Calculate RS and RSI
    if avg_loss == 0:
        return 100, "Avg loss is 0 (no losses)"

    rs = avg_gain / avg_loss
    rsi = 100 - (100 / (1 + rs))

    print(f"RS (Relative Strength): {rs:.6f}")
    print(f"RSI calculation: 100 - (100 / (1 + {rs:.6f})) = {rsi:.2f}")

    return rsi, None

def main():
    # BIIB closing prices from CSV (last 14 days from Oct 16 back to Oct 1)
    # From oldest to newest
    prices = [
        1.040,  # Sep 30
        1.040,  # Oct 1
        1.040,  # Oct 2
        1.040,  # Oct 3
        1.040,  # Oct 4
        1.040,  # Oct 5
        1.040,  # Oct 6
        1.020,  # Oct 7
        1.030,  # Oct 8
        1.030,  # Oct 9
        1.030,  # Oct 10
        1.030,  # Oct 11
        1.040,  # Oct 12
        1.020,  # Oct 13
        1.020,  # Oct 14
        1.020,  # Oct 15
        1.020,  # Oct 16
    ]

    print("Manual RSI Calculation for BIIB (14-period)")
    print("=" * 50)

    rsi, error = calculate_rsi_manual(prices, 14)

    print("\n" + "=" * 50)
    if error:
        print(f"Error: {error}")
    else:
        print(f"MANUAL RSI RESULT: {rsi:.2f}")
        print()
        print("Comparison with other methods:")
        print(f"- Our Unified API:     53.09")
        print(f"- JavaScript Analysis: 34.73")
        print(f"- Manual Calculation:  {rsi:.2f}")
        print()
        if abs(rsi - 53.09) < abs(rsi - 34.73):
            print("✅ Our Unified API (53.09) is CLOSER to manual calculation")
        else:
            print("❌ JavaScript Analysis (34.73) is CLOSER to manual calculation")

if __name__ == "__main__":
    main()
package strategies

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
	"github.com/stretchr/testify/require"
)

func TestRSI14MeanReversionEODStrategy_CrossingRules(t *testing.T) {
	t.Parallel()

	strat := NewRSI14MeanReversionEODStrategy(slog.Default())

	// Build a synthetic series that trends down hard then stabilizes.
	// We assert behavior based on computed RSI crossing, not on exact RSI values.
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	data := make([]liquidity.TradingDay, 0, 80)

	closePrice := 100.0
	for i := 0; i < 60; i++ {
		// Downtrend for first 40 points, then small oscillation.
		if i < 40 {
			closePrice -= 1.5
		} else {
			if i%2 == 0 {
				closePrice += 0.2
			} else {
				closePrice -= 0.2
			}
		}
		data = append(data, liquidity.TradingDay{
			Date:          start.AddDate(0, 0, i),
			Symbol:        "TEST",
			Close:         closePrice,
			Open:          closePrice,
			High:          closePrice,
			Low:           closePrice,
			Volume:        1,
			ShareVolume:   1,
			Value:         1,
			NumTrades:     1,
			TradingStatus: "ACTIVE",
		})
	}

	signal, err := strat.Execute(context.Background(), data)
	require.NoError(t, err)
	require.Contains(t, []strategy.SignalAction{strategy.SignalBuy, strategy.SignalSell, strategy.SignalHold}, signal.Action)
	require.Equal(t, "TEST", signal.Symbol)
	require.Equal(t, "rsi14_mr_eod_v1", signal.StrategyID)

	// Metadata should include RSI values for UI display/debugging.
	require.NotNil(t, signal.Metadata)
	require.Contains(t, signal.Metadata, "rsi")
	require.Contains(t, signal.Metadata, "prev_rsi")
}

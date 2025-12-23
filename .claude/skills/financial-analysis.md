---
name: financial-analysis
type: domain-skill
expertise_level: expert
domain: isx-financial-markets
version: "1.0.0"
---

# Financial Analysis Skill - ISX Market Expertise

## Purpose
Specialized expertise in Iraqi Stock Exchange (ISX) financial data analysis, market behavior patterns, and trading indicator calculations for the ISX Pulse platform.

## Core Competencies

### ISX Market Structure Knowledge
- **Trading Calendar**: Sunday-Thursday trading week, Iraqi holidays
- **Market Hours**: 10:00 AM - 12:30 PM Iraq time (EOD focus)
- **Market Segments**: Regular market, ISX60 index, SME market
- **Settlement Cycle**: T+2 settlement for most securities
- **Currency**: All trading in Iraqi Dinar (IQD) with USD reporting

### Financial Data Processing
- **Price Data**: Open, High, Low, Close (OHLC) patterns
- **Volume Analysis**: Trading volume patterns and liquidity metrics
- **Market Cap Calculations**: Outstanding shares × current price
- **Sector Classification**: Banking, Industrial, Services, Agriculture, Insurance
- **Dividend Yield**: Annual dividends / current price × 100

### Technical Analysis Fundamentals
- **Support & Resistance**: Key price levels with buying/selling pressure
- **Trend Analysis**: Uptrend, downtrend, sideways market identification
- **Volume Patterns**: Volume confirmation of price movements
- **Market Sentiment**: Bullish/bearish indicators from price action
- **Volatility Analysis**: Price fluctuation patterns and risk assessment

## ISX-Specific Analysis Patterns

### 1. Banking Sector Dominance
- **Market Leaders**: Banks typically represent 60-70% of total market cap
- **Key Indicators**: P/E ratios, NPL ratios, loan growth metrics
- **Regulatory Impact**: Central Bank policies affecting banking stocks
- **Seasonal Patterns**: Dividend payouts affecting banking stock prices

### 2. Oil & Gas Sector Influence
- **Global Oil Prices**: Direct correlation with Iraqi economy
- **Government Revenue**: Oil revenue impact on market sentiment
- **Sector Weight**: Limited but influential sector representation
- **Currency Effects**: Oil price → USD/IQD → ISX valuations

### 3. Market Liquidity Analysis
- **Thin Trading**: Many stocks have limited daily volume
- **Price Impact**: Large orders can significantly move prices
- **Market Depth**: Limited order book depth in most securities
- **Institutional Activity**: Foreign vs local investor participation

## Integration with ISX Pulse Architecture

### SSOT Pattern Alignment
- **Data Consistency**: Ensure all calculations use same price source
- **Indicator Storage**: Pre-calculated indicators in ticker_indicators.csv
- **Update Frequency**: EOD updates for all indicators
- **Version Control**: Track methodology changes over time

### Pipeline Integration
- **Processing Stage**: Post-CSV validation, pre-indicator calculation
- **Error Handling**: Robust processing of missing or invalid data
- **Performance**: Efficient batch processing of all tickers
- **Audit Trail**: Log all calculation parameters and results

## Use Cases

### Market Analysis
- **Sector Performance**: Compare banking vs industrial sector returns
- **Market Breadth**: Advances/declines, new highs/lows analysis
- **Relative Strength**: Stock performance vs market index
- **Correlation Analysis**: Inter-stock relationship patterns

### Risk Assessment
- **Volatility Analysis**: Historical price volatility patterns
- **Drawdown Analysis**: Maximum price decline from peaks
- **Beta Calculation**: Stock sensitivity to market movements
- **Liquidity Risk**: Trading volume and market impact analysis

## Best Practices

### Data Processing
1. **Always validate** price ranges against historical patterns
2. **Handle missing data** with appropriate interpolation methods
3. **Document assumptions** for calculation methodologies
4. **Maintain audit trails** for all data transformations
5. **Test thoroughly** with historical data before deployment

### Analysis Standards
1. **Context awareness**: Always consider Iraqi market specifics
2. **Risk focus**: Emphasize risk assessment in volatile markets
3. **Liquidity consideration**: Account for thin trading patterns
4. **Regulatory compliance**: Follow ISX and CBI regulations
5. **Cultural sensitivity**: Understand local market practices
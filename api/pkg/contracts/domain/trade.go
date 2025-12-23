package domain

import (
	"time"
)

// Trade represents a stock trade transaction
type Trade struct {
	ID            string                 `json:"id" db:"id" validate:"required,uuid"`
	TradeNumber   string                 `json:"trade_number" db:"trade_number" validate:"required"`
	Symbol        string                 `json:"symbol" db:"symbol" validate:"required"`
	TradeDate     time.Time              `json:"trade_date" db:"trade_date"`
	TradeTime     time.Time              `json:"trade_time" db:"trade_time"`
	Price         float64                `json:"price" db:"price" validate:"required,min=0"`
	Quantity      int64                  `json:"quantity" db:"quantity" validate:"required,min=1"`
	Value         float64                `json:"value" db:"value" validate:"required,min=0"`
	Side          TradeSide              `json:"side" db:"side" validate:"required"`
	OrderType     OrderType              `json:"order_type" db:"order_type"`
	Settlement    SettlementType         `json:"settlement" db:"settlement"`
	Currency      string                 `json:"currency" db:"currency" validate:"required,len=3"`
	BrokerBuy     string                 `json:"broker_buy,omitempty" db:"broker_buy"`
	BrokerSell    string                 `json:"broker_sell,omitempty" db:"broker_sell"`
	MarketType    MarketType             `json:"market_type" db:"market_type"`
	TradeCondition string                `json:"trade_condition,omitempty" db:"trade_condition"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// TradeSide represents the side of a trade
type TradeSide string

const (
	TradeSideBuy  TradeSide = "buy"
	TradeSideSell TradeSide = "sell"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeMarket     OrderType = "market"
	OrderTypeLimit      OrderType = "limit"
	OrderTypeStop       OrderType = "stop"
	OrderTypeStopLimit  OrderType = "stop_limit"
)

// SettlementType represents the settlement type
type SettlementType string

const (
	SettlementT0 SettlementType = "T+0"
	SettlementT1 SettlementType = "T+1"
	SettlementT2 SettlementType = "T+2"
	SettlementT3 SettlementType = "T+3"
)

// MarketType represents the market type
type MarketType string

const (
	MarketTypeRegular    MarketType = "regular"
	MarketTypeNegotiated MarketType = "negotiated"
	MarketTypeOddLot     MarketType = "odd_lot"
	MarketTypeBlock      MarketType = "block"
)

// OrderBook represents the order book for a symbol
type OrderBook struct {
	Symbol    string          `json:"symbol"`
	Timestamp time.Time       `json:"timestamp"`
	Bids      []OrderBookEntry `json:"bids"`
	Asks      []OrderBookEntry `json:"asks"`
	Spread    float64         `json:"spread"`
	MidPrice  float64         `json:"mid_price"`
}

// OrderBookEntry represents a single entry in the order book
type OrderBookEntry struct {
	Price    float64 `json:"price" validate:"required,min=0"`
	Quantity int64   `json:"quantity" validate:"required,min=0"`
	Orders   int     `json:"orders" validate:"min=0"`
}

// MarketDepth represents market depth data
type MarketDepth struct {
	Symbol          string    `json:"symbol"`
	Timestamp       time.Time `json:"timestamp"`
	TotalBidVolume  int64     `json:"total_bid_volume"`
	TotalAskVolume  int64     `json:"total_ask_volume"`
	TotalBidValue   float64   `json:"total_bid_value"`
	TotalAskValue   float64   `json:"total_ask_value"`
	BidLevels       int       `json:"bid_levels"`
	AskLevels       int       `json:"ask_levels"`
	ImbalanceRatio  float64   `json:"imbalance_ratio"`
}

// TradeStatistics represents trade statistics for a period
type TradeStatistics struct {
	Symbol        string        `json:"symbol"`
	Period        string        `json:"period"` // daily, weekly, monthly
	StartDate     time.Time     `json:"start_date"`
	EndDate       time.Time     `json:"end_date"`
	TotalTrades   int64         `json:"total_trades"`
	TotalVolume   int64         `json:"total_volume"`
	TotalValue    float64       `json:"total_value"`
	AvgTradeSize  float64       `json:"avg_trade_size"`
	AvgTradeValue float64       `json:"avg_trade_value"`
	VWAP          float64       `json:"vwap"`
	HighPrice     float64       `json:"high_price"`
	LowPrice      float64       `json:"low_price"`
	PriceRange    float64       `json:"price_range"`
	Volatility    float64       `json:"volatility"`
	BuyVolume     int64         `json:"buy_volume"`
	SellVolume    int64         `json:"sell_volume"`
	BuySellRatio  float64       `json:"buy_sell_ratio"`
}

// IntradayData represents intraday trading data
type IntradayData struct {
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	Interval  string    `json:"interval"` // 1m, 5m, 15m, 30m, 1h
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	Value     float64   `json:"value"`
	Trades    int       `json:"trades"`
	VWAP      float64   `json:"vwap"`
}

// MarketBreadth represents market breadth indicators
type MarketBreadth struct {
	Date            time.Time `json:"date"`
	AdvancingStocks int       `json:"advancing_stocks"`
	DecliningStocks int       `json:"declining_stocks"`
	UnchangedStocks int       `json:"unchanged_stocks"`
	NewHighs        int       `json:"new_highs"`
	NewLows         int       `json:"new_lows"`
	AdvanceVolume   int64     `json:"advance_volume"`
	DeclineVolume   int64     `json:"decline_volume"`
	ADRatio         float64   `json:"ad_ratio"` // Advance/Decline Ratio
	ADLine          float64   `json:"ad_line"`  // Advance/Decline Line
	TRIN            float64   `json:"trin"`     // Trading Index (Arms Index)
	McClellanIndex  float64   `json:"mcclellan_index"`
}

// TradeFilter represents filters for trade queries
type TradeFilter struct {
	Symbols        []string       `json:"symbols,omitempty"`
	DateFrom       *time.Time     `json:"date_from,omitempty"`
	DateTo         *time.Time     `json:"date_to,omitempty"`
	MinPrice       float64        `json:"min_price,omitempty"`
	MaxPrice       float64        `json:"max_price,omitempty"`
	MinQuantity    int64          `json:"min_quantity,omitempty"`
	MaxQuantity    int64          `json:"max_quantity,omitempty"`
	MinValue       float64        `json:"min_value,omitempty"`
	MaxValue       float64        `json:"max_value,omitempty"`
	Sides          []TradeSide    `json:"sides,omitempty"`
	OrderTypes     []OrderType    `json:"order_types,omitempty"`
	MarketTypes    []MarketType   `json:"market_types,omitempty"`
	BrokerCodes    []string       `json:"broker_codes,omitempty"`
}
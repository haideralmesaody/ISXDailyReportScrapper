package contracts

import "time"

// AlertRuleRequest represents a request to create/update an alert rule
type AlertRuleRequest struct {
	Symbol    string  `json:"symbol" validate:"required"`
	RuleType  string  `json:"rule_type" validate:"required,oneof=PRICE RSI"`
	Condition string  `json:"condition" validate:"required"`
	Value     float64 `json:"value" validate:"required,gt=0"`
	Active    *bool   `json:"active,omitempty"`
}

// AlertRuleResponse represents a response with alert rule details
type AlertRuleResponse struct {
	ID            string                 `json:"id"`
	Symbol        string                 `json:"symbol"`
	RuleType      string                 `json:"rule_type"`
	Condition     string                 `json:"condition"`
	Value         float64                `json:"value"`
	Active        bool                   `json:"active"`
	CreatedAt     time.Time              `json:"created_at"`
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
}

// PriceAlertResponse represents a price alert response
type PriceAlertResponse struct {
	ID            string    `json:"id"`
	Symbol        string    `json:"symbol"`
	AlertType     string    `json:"alert_type"`
	TargetPrice   float64   `json:"target_price"`
	CurrentPrice  float64   `json:"current_price"`
	PriceChange   float64   `json:"price_change"`
	PercentChange float64   `json:"percent_change"`
	Volume        int64     `json:"volume"`
	Direction     string    `json:"direction"`
	TriggeredAt   time.Time `json:"triggered_at"`
	Message       string    `json:"message"`
}

// RSIAlertResponse represents an RSI alert response
type RSIAlertResponse struct {
	ID            string    `json:"id"`
	Symbol        string    `json:"symbol"`
	AlertType     string    `json:"alert_type"`
	RSIValue      float64   `json:"rsi_value"`
	RSIPeriod     int       `json:"rsi_period"`
	Threshold     float64   `json:"threshold"`
	SignalStrength string  `json:"signal_strength"`
	TrendDirection string  `json:"trend_direction"`
	TriggeredAt   time.Time `json:"triggered_at"`
	Message       string    `json:"message"`
}

// AlertSummary represents a summary of alerts for a specific date
type AlertSummary struct {
	Date            string                `json:"date"`
	TotalAlerts     int                   `json:"total_alerts"`
	PriceAlerts     int                   `json:"price_alerts"`
	RSIAlerts       int                   `json:"rsi_alerts"`
	PriceAlertsList []PriceAlertResponse  `json:"price_alerts_list"`
	RSIAlertsList   []RSIAlertResponse    `json:"rsi_alerts_list"`
}

// AlertConfigRequest represents a request to update alert configuration
type AlertConfigRequest struct {
	PriceChangeThreshold *float64 `json:"price_change_threshold,omitempty" validate:"omitempty,gt=0,lte=100"`
	MinAlertVolume       *int64   `json:"min_alert_volume,omitempty" validate:"omitempty,gt=0"`
	RSIPeriod            *int     `json:"rsi_period,omitempty" validate:"omitempty,gte=1,lte=100"`
	RSIOverboughtLevel   *int     `json:"rsi_overbought_level,omitempty" validate:"omitempty,gte=50,lte=100"`
	RSIOversoldLevel     *int     `json:"rsi_oversold_level,omitempty" validate:"omitempty,gte=1,lte=50"`
	MaxAlertsPerDay      *int     `json:"max_alerts_per_day,omitempty" validate:"omitempty,gte=1,lte=1000"`
	CooldownMinutes      *int     `json:"cooldown_minutes,omitempty" validate:"omitempty,gte=1,lte=1440"`
}

// AlertConfigResponse represents the current alert configuration
type AlertConfigResponse struct {
	PriceChangeThreshold float64 `json:"price_change_threshold"`
	MinAlertVolume       int64   `json:"min_alert_volume"`
	RSIPeriod            int     `json:"rsi_period"`
	RSIOverboughtLevel   int     `json:"rsi_overbought_level"`
	RSIOversoldLevel     int     `json:"rsi_oversold_level"`
	MaxAlertsPerDay      int     `json:"max_alerts_per_day"`
	CooldownMinutes      int     `json:"cooldown_minutes"`
}

// AlertHistoryRequest represents a request for alert history
type AlertHistoryRequest struct {
	FromDate   string `json:"from_date,omitempty" validate:"omitempty,date"`
	ToDate     string `json:"to_date,omitempty" validate:"omitempty,date"`
	Symbol     string `json:"symbol,omitempty"`
	AlertType  string `json:"alert_type,omitempty" validate:"omitempty,oneof=PRICE RSI"`
	Limit      int    `json:"limit,omitempty" validate:"omitempty,gte=1,lte=1000"`
	Offset     int    `json:"offset,omitempty" validate:"omitempty,gte=0"`
}

// AlertHistoryResponse represents paginated alert history
type AlertHistoryResponse struct {
	PriceAlerts   []PriceAlertResponse `json:"price_alerts"`
	RSIAlerts     []RSIAlertResponse   `json:"rsi_alerts"`
	TotalCount    int                  `json:"total_count"`
	HasMore       bool                 `json:"has_more"`
	NextOffset    int                  `json:"next_offset,omitempty"`
}

// AlertStatsResponse represents alert statistics
type AlertStatsResponse struct {
	TotalAlertsGenerated int                    `json:"total_alerts_generated"`
	PriceAlertsGenerated  int                    `json:"price_alerts_generated"`
	RSIAlertsGenerated    int                    `json:"rsi_alerts_generated"`
	TopSymbols           []AlertSymbolStats      `json:"top_symbols"`
	DailyBreakdown       []AlertDailyBreakdown   `json:"daily_breakdown"`
	AlertTypeBreakdown   AlertTypeBreakdown       `json:"alert_type_breakdown"`
}

// AlertSymbolStats represents alert statistics per symbol
type AlertSymbolStats struct {
	Symbol       string `json:"symbol"`
	AlertCount   int    `json:"alert_count"`
	LastAlertAt  string `json:"last_alert_at,omitempty"`
}

// AlertDailyBreakdown represents daily alert breakdown
type AlertDailyBreakdown struct {
	Date        string `json:"date"`
	PriceAlerts int    `json:"price_alerts"`
	RSIAlerts   int    `json:"rsi_alerts"`
	TotalAlerts int    `json:"total_alerts"`
}

// AlertTypeBreakdown represents alert type breakdown
type AlertTypeBreakdown struct {
	PriceAlerts int `json:"price_alerts"`
	RSIAlerts   int `json:"rsi_alerts"`
}

// WebSocketAlertMessage represents a WebSocket message for real-time alerts
type WebSocketAlertMessage struct {
	Type      string      `json:"type"` // "price_alert", "rsi_alert"
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// WebSocketAlertData represents the data structure for WebSocket alerts
type WebSocketAlertData struct {
	ID          string                 `json:"id"`
	Symbol      string                 `json:"symbol"`
	AlertType   string                 `json:"alert_type"`
	TriggerValue float64              `json:"trigger_value"`
	Message     string                 `json:"message"`
	TriggeredAt time.Time             `json:"triggered_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationError represents a validation error response
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string          `json:"error"`
	Message string          `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
	Code    string          `json:"code,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Alert rule condition constants
const (
	AlertConditionAbove       = "ABOVE"
	AlertConditionBelow       = "BELOW"
	AlertConditionPercentUp   = "PERCENT_UP"
	AlertConditionPercentDown = "PERCENT_DOWN"
	AlertConditionOverbought  = "OVERBOUGHT"
	AlertConditionOversold    = "OVERSOLD"
)

// Alert type constants
const (
	AlertTypePrice = "PRICE"
	AlertTypeRSI   = "RSI"
)

// Alert signal strength constants
const (
	SignalStrengthWeak     = "WEAK"
	SignalStrengthModerate = "MODERATE"
	SignalStrengthStrong   = "STRONG"
)

// Alert trend direction constants
const (
	TrendDirectionImproving    = "IMPROVING"
	TrendDirectionDeteriorating = "DETERIORATING"
	TrendDirectionStable       = "STABLE"
)

// Alert direction constants
const (
	DirectionUp   = "UP"
	DirectionDown = "DOWN"
)
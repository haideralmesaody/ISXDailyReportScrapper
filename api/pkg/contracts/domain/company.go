package domain

import (
	"time"
)

// Company represents a listed company
type Company struct {
	ID                string                 `json:"id" db:"id" validate:"required,uuid"`
	Name              string                 `json:"name" db:"name" validate:"required,min=2,max=200"`
	NameAr            string                 `json:"name_ar,omitempty" db:"name_ar"`
	Symbol            string                 `json:"symbol" db:"symbol" validate:"required,min=1,max=10"`
	ISINCode          string                 `json:"isin_code" db:"isin_code" validate:"required,len=12"`
	Sector            string                 `json:"sector" db:"sector" validate:"required"`
	SubSector         string                 `json:"sub_sector,omitempty" db:"sub_sector"`
	Industry          string                 `json:"industry,omitempty" db:"industry"`
	Description       string                 `json:"description,omitempty" db:"description"`
	DescriptionAr     string                 `json:"description_ar,omitempty" db:"description_ar"`
	Website           string                 `json:"website,omitempty" db:"website" validate:"omitempty,url"`
	Email             string                 `json:"email,omitempty" db:"email" validate:"omitempty,email"`
	Phone             string                 `json:"phone,omitempty" db:"phone"`
	Fax               string                 `json:"fax,omitempty" db:"fax"`
	Address           Address                `json:"address,omitempty" db:"address"`
	IncorporationDate time.Time              `json:"incorporation_date" db:"incorporation_date"`
	ListingDate       time.Time              `json:"listing_date" db:"listing_date"`
	FiscalYearEnd     string                 `json:"fiscal_year_end,omitempty" db:"fiscal_year_end"` // MM-DD format
	Employees         int                    `json:"employees,omitempty" db:"employees"`
	SharesOutstanding int64                  `json:"shares_outstanding" db:"shares_outstanding"`
	FreeFloat         float64                `json:"free_float" db:"free_float"` // Percentage
	Status            CompanyStatus          `json:"status" db:"status"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// Address represents a company address
type Address struct {
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

// CompanyStatus represents the status of a company
type CompanyStatus string

const (
	CompanyStatusActive    CompanyStatus = "active"
	CompanyStatusSuspended CompanyStatus = "suspended"
	CompanyStatusDelisted  CompanyStatus = "delisted"
	CompanyStatusMerged    CompanyStatus = "merged"
)

// CompanyFinancials represents financial data for a company
type CompanyFinancials struct {
	ID              string                 `json:"id" db:"id" validate:"required,uuid"`
	CompanyID       string                 `json:"company_id" db:"company_id" validate:"required,uuid"`
	Period          string                 `json:"period" db:"period" validate:"required"` // e.g., "Q1 2024", "FY 2023"
	PeriodType      string                 `json:"period_type" db:"period_type" validate:"required,oneof=quarterly annual"`
	StartDate       time.Time              `json:"start_date" db:"start_date"`
	EndDate         time.Time              `json:"end_date" db:"end_date"`
	Currency        string                 `json:"currency" db:"currency" validate:"required,len=3"`
	Revenue         float64                `json:"revenue" db:"revenue"`
	GrossProfit     float64                `json:"gross_profit" db:"gross_profit"`
	OperatingIncome float64                `json:"operating_income" db:"operating_income"`
	NetIncome       float64                `json:"net_income" db:"net_income"`
	EPS             float64                `json:"eps" db:"eps"` // Earnings Per Share
	DilutedEPS      float64                `json:"diluted_eps" db:"diluted_eps"`
	TotalAssets     float64                `json:"total_assets" db:"total_assets"`
	TotalLiabilities float64               `json:"total_liabilities" db:"total_liabilities"`
	ShareholderEquity float64              `json:"shareholder_equity" db:"shareholder_equity"`
	CashFlow        float64                `json:"cash_flow" db:"cash_flow"`
	Dividends       float64                `json:"dividends" db:"dividends"`
	ReportDate      time.Time              `json:"report_date" db:"report_date"`
	AuditStatus     string                 `json:"audit_status" db:"audit_status" validate:"omitempty,oneof=audited unaudited reviewed"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// CompanyOfficer represents a company officer or board member
type CompanyOfficer struct {
	ID          string    `json:"id" db:"id" validate:"required,uuid"`
	CompanyID   string    `json:"company_id" db:"company_id" validate:"required,uuid"`
	Name        string    `json:"name" db:"name" validate:"required"`
	NameAr      string    `json:"name_ar,omitempty" db:"name_ar"`
	Position    string    `json:"position" db:"position" validate:"required"`
	PositionAr  string    `json:"position_ar,omitempty" db:"position_ar"`
	Type        string    `json:"type" db:"type" validate:"required,oneof=executive board"`
	StartDate   time.Time `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	Biography   string    `json:"biography,omitempty" db:"biography"`
	BiographyAr string    `json:"biography_ar,omitempty" db:"biography_ar"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CompanyEvent represents a significant company event
type CompanyEvent struct {
	ID          string                 `json:"id" db:"id" validate:"required,uuid"`
	CompanyID   string                 `json:"company_id" db:"company_id" validate:"required,uuid"`
	Type        CompanyEventType       `json:"type" db:"type" validate:"required"`
	Title       string                 `json:"title" db:"title" validate:"required"`
	TitleAr     string                 `json:"title_ar,omitempty" db:"title_ar"`
	Description string                 `json:"description" db:"description"`
	DescriptionAr string               `json:"description_ar,omitempty" db:"description_ar"`
	EventDate   time.Time              `json:"event_date" db:"event_date"`
	AnnouncedDate time.Time            `json:"announced_date" db:"announced_date"`
	Impact      string                 `json:"impact,omitempty" db:"impact" validate:"omitempty,oneof=positive negative neutral"`
	Documents   []string               `json:"documents,omitempty" db:"documents"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// CompanyEventType defines types of company events
type CompanyEventType string

const (
	CompanyEventTypeEarnings     CompanyEventType = "earnings"
	CompanyEventTypeDividend     CompanyEventType = "dividend"
	CompanyEventTypeMerger       CompanyEventType = "merger"
	CompanyEventTypeAcquisition  CompanyEventType = "acquisition"
	CompanyEventTypeSplit        CompanyEventType = "split"
	CompanyEventTypeIPO          CompanyEventType = "ipo"
	CompanyEventTypeRightsIssue  CompanyEventType = "rights_issue"
	CompanyEventTypeBonusShares  CompanyEventType = "bonus_shares"
	CompanyEventTypeAGM          CompanyEventType = "agm" // Annual General Meeting
	CompanyEventTypeEGM          CompanyEventType = "egm" // Extraordinary General Meeting
	CompanyEventTypeManagement   CompanyEventType = "management_change"
	CompanyEventTypeRegulatory   CompanyEventType = "regulatory"
	CompanyEventTypeOther        CompanyEventType = "other"
)

// CompanyFilter represents filters for company queries
type CompanyFilter struct {
	Sectors           []string         `json:"sectors,omitempty"`
	SubSectors        []string         `json:"sub_sectors,omitempty"`
	Industries        []string         `json:"industries,omitempty"`
	Statuses          []CompanyStatus  `json:"statuses,omitempty"`
	MinMarketCap      float64          `json:"min_market_cap,omitempty"`
	MaxMarketCap      float64          `json:"max_market_cap,omitempty"`
	MinEmployees      int              `json:"min_employees,omitempty"`
	MaxEmployees      int              `json:"max_employees,omitempty"`
	ListedAfter       *time.Time       `json:"listed_after,omitempty"`
	ListedBefore      *time.Time       `json:"listed_before,omitempty"`
	SearchTerm        string           `json:"search_term,omitempty"`
}
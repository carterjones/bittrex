package bittrex

// Market represents a market on Bittrex for trading between two currencies: the
// market currency and the base currency.
type Market struct {
	MarketCurrency     string
	BaseCurrency       string
	MarketCurrencyLong string
	BaseCurrencyLong   string
	MinTradeSize       float64
	Name               string `json:"MarketName"`
	IsActive           bool
	Created            string
	Notice             *string
	IsSponsored        *bool
	LogoURL            string `json:"LogoUrl"`
}

func (m Market) String() string {
	return m.Name
}

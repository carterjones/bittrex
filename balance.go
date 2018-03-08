package bittrex

// Balance represents an account balance.
type Balance struct {
	Currency      string
	Balance       float64
	Available     float64
	Pending       float64
	CryptoAddress string
}

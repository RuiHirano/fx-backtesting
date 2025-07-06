package models

// Config holds the configuration for backtesting
type Config struct {
	InitialBalance float64 // Initial balance in base currency
	Spread         float64 // Bid-ask spread
	Commission     float64 // Commission rate (percentage)
	Slippage       float64 // Slippage in pips
	Leverage       float64 // Leverage ratio
}

// NewConfig creates a new Config instance
func NewConfig(initialBalance, spread, commission, slippage, leverage float64) Config {
	return Config{
		InitialBalance: initialBalance,
		Spread:         spread,
		Commission:     commission,
		Slippage:       slippage,
		Leverage:       leverage,
	}
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() Config {
	return Config{
		InitialBalance: 10000.0, // $10,000
		Spread:         0.0001,  // 1 pip
		Commission:     0.0,     // No commission
		Slippage:       0.0,     // No slippage
		Leverage:       100.0,   // 1:100 leverage
	}
}

// IsValid checks if the configuration is valid
func (c Config) IsValid() bool {
	if c.InitialBalance <= 0 {
		return false
	}
	
	if c.Spread < 0 {
		return false
	}
	
	if c.Commission < 0 {
		return false
	}
	
	if c.Slippage < 0 {
		return false
	}
	
	if c.Leverage <= 0 {
		return false
	}
	
	return true
}

// CalculateMarginRequired calculates the margin required for a position
func (c Config) CalculateMarginRequired(positionSize, entryPrice float64) float64 {
	return (positionSize * entryPrice) / c.Leverage
}

// CalculateCommission calculates the commission for a trade
func (c Config) CalculateCommission(positionSize, entryPrice float64) float64 {
	return positionSize * entryPrice * c.Commission
}

// ApplySpread applies the spread to get the actual execution price
func (c Config) ApplySpread(midPrice float64, side OrderSide) float64 {
	halfSpread := c.Spread / 2.0
	
	if side == OrderSideBuy {
		// Buy at ask price (higher)
		return midPrice + halfSpread
	} else {
		// Sell at bid price (lower)
		return midPrice - halfSpread
	}
}
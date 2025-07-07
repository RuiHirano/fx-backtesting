# FX Backtesting Library - Examples

This directory contains practical examples demonstrating how to use the FX backtesting library.

## Examples Overview

### 1. Basic Example (`basic_example.go`)
A simple introduction to the library showing:
- Basic configuration setup
- Loading CSV data
- Running a moving average crossover strategy
- Generating text reports
- Interpreting results

**Run:**
```bash
cd examples
go run basic_example.go
```

### 2. Advanced Example (`advanced_example.go`)
Demonstrates advanced features including:
- Strategy optimization across multiple parameter sets
- Performance comparison between strategies
- Automated report generation (text, JSON, CSV)
- Risk assessment and recommendations

**Run:**
```bash
cd examples
go run advanced_example.go
```

### 3. Configuration File (`config.json`)
Sample JSON configuration file showing custom settings:
- Initial balance: $5,000
- Spread: 2 pips
- Commission: $1 per trade
- Leverage: 1:50

## CLI Usage Examples

### Basic CLI Usage
```bash
# Simple backtest with default settings
go run ../cmd/backtester/main.go --data ../testdata/sample.csv

# Custom strategy parameters
go run ../cmd/backtester/main.go \
  --data ../testdata/sample.csv \
  --fast-period 5 \
  --slow-period 20 \
  --position-size 2000

# Generate JSON report
go run ../cmd/backtester/main.go \
  --data ../testdata/sample.csv \
  --format json \
  --output results.json

# Use custom configuration
go run ../cmd/backtester/main.go \
  --data ../testdata/sample.csv \
  --config config.json \
  --output detailed_report.txt
```

### Advanced CLI Usage
```bash
# High-frequency scalping strategy
go run ../cmd/backtester/main.go \
  --data ../testdata/sample.csv \
  --fast-period 2 \
  --slow-period 5 \
  --position-size 500 \
  --format json

# Conservative long-term strategy
go run ../cmd/backtester/main.go \
  --data ../testdata/sample.csv \
  --fast-period 20 \
  --slow-period 50 \
  --position-size 1000 \
  --output conservative_results.txt

# Export trade details to CSV
go run ../cmd/backtester/main.go \
  --data ../testdata/sample.csv \
  --format csv \
  --output trades.csv
```

## Data Format

The library expects CSV files with the following format:
```csv
timestamp,open,high,low,close,volume
2024-01-01 09:00:00,1.0500,1.0520,1.0490,1.0510,1000
2024-01-01 09:01:00,1.0510,1.0530,1.0500,1.0520,1200
...
```

## Strategy Customization

You can customize the moving average strategy by adjusting:
- **Fast Period**: Shorter period for quick signal detection
- **Slow Period**: Longer period for trend confirmation
- **Position Size**: Units to trade per signal

## Output Formats

### Text Report
Human-readable summary with:
- Performance metrics
- Trade statistics
- Risk analysis
- Executive summary

### JSON Report
Machine-readable format for:
- Programmatic analysis
- Data visualization
- Integration with other tools

### CSV Report
Trade-by-trade details for:
- Detailed analysis
- Spreadsheet import
- Custom reporting

## Performance Tips

1. **Larger Datasets**: Use more historical data for robust results
2. **Parameter Optimization**: Test multiple strategy configurations
3. **Risk Management**: Monitor drawdown and Sharpe ratios
4. **Market Conditions**: Analyze performance across different periods
5. **Transaction Costs**: Include realistic spreads and commissions

## Next Steps

1. Try the basic example first to understand the workflow
2. Experiment with different strategy parameters
3. Use the advanced example for strategy optimization
4. Implement your own custom strategies using the framework
5. Backtest on longer historical datasets for production use
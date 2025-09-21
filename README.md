# DazedTrader

A TUI for Robinhood Crypto trading.

## ✨ Features

### 🎯 Core Crypto Trading Functionality
- **Real-time Crypto Portfolio** - Live portfolio with current prices and day changes
- **Interactive Crypto Trading** - 6-step trading interface with live price estimates
- **🤖 Algorithmic Trading** - Automated trading with configurable strategies and risk management
- **Order History** - Complete order tracking with status indicators
- **Market Data** - Top gaining/losing cryptocurrencies with real-time data
- **Real-Time Crypto News** - Live news feed from CryptoCompare API with impact analysis
- **Live Price Updates** - Real-time pricing throughout the trading experience

### 🤖 Algorithmic Trading Features
- **Multiple Trading Strategies** - Moving Average, Momentum, Mean Reversion algorithms
- **Real-time Price Monitoring** - High-frequency price feeds for fast decision making
- **Advanced Risk Management** - Position sizing, stop-losses, daily limits, max drawdown protection
- **Portfolio Performance Analytics** - Sharpe ratio, win rate, profit factor, drawdown analysis
- **Emergency Controls** - Instant stop, pause/resume, per-strategy controls
- **Strategy Configuration** - JSON-based configs with hot-reloading capabilities
- **Backtesting Support** - Test strategies on historical data before going live


## 🚀 Installation & Setup

### Prerequisites
- **Go 1.24 or higher**
- **Terminal with 256 color support**
- **Robinhood account with Crypto API access**
- **Robinhood Crypto API credentials** (API key + Ed25519 private key)

### Getting Robinhood Crypto API Credentials

1. **Visit Robinhood Crypto API Documentation**: https://docs.robinhood.com/crypto/trading/
2. **Create API credentials** in your Robinhood account
3. **Generate Ed25519 key pair** as instructed
4. **Save your API key and private key** (you'll need both)

**Format required**: `apikey:privatekey` where privatekey is base64-encoded

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/DazedTrader.git
cd DazedTrader

# Initialize Go modules and download dependencies
go mod tidy

# Build the main application
go build -o dazedtrader .

# Build the algorithmic trading configuration utility
go build -o configure_algo cmd/configure_algo/main.go

# Run the application
./dazedtrader
```

### Security Check (Optional)

```bash
# Run security check to ensure no credentials in source
./check_security.sh
```

### 🤖 Algorithmic Trading Setup

Before using algorithmic trading features, you need to configure financial limits and risk parameters for safety.

#### Quick Setup (Recommended for beginners)

```bash
# 1. Configure ultra-conservative limits ($100 budget, $5 max per trade)
./configure_algo -preset=ultra_conservative

# 2. View your configuration
./configure_algo -show

# 3. Start the application
./dazedtrader

# 4. Navigate to "🤖 Algorithmic Trading" in the menu
# 5. Press '1' to enable the first strategy (it starts disabled)
# 6. Press 'S' to start the trading engine
# 7. Monitor closely and use 'E' for emergency stop if needed
```

#### Custom Budget Setup

```bash
# Configure with custom budget and daily loss limit
./configure_algo -budget=500 -daily=25

# This creates:
# - Max position size: $50 per trade (10% of budget)
# - Daily loss limit: $25 total across all strategies
# - Stop loss: 3% per trade
# - Conservative strategy parameters
```

#### Available Risk Presets

| Preset | Total Budget | Max Per Trade | Daily Loss Limit | Best For |
|--------|-------------|---------------|------------------|----------|
| `ultra_conservative` | $100 | $5 | $10 | Learning, testing |
| `conservative` | $500 | $50 | $25 | Steady growth |
| `moderate` | $1000 | $150 | $50 | Experienced users |

#### Configuration Files Location

After running `configure_algo`, files are created in:
```
~/.dazedtrader/algo/
├── engine.json           # Engine-wide limits
└── strategies/
    ├── btc_*.json        # BTC strategy configuration
    └── eth_*.json        # ETH strategy configuration
```

#### Manual Configuration (Advanced)

You can manually edit these JSON files for fine-tuned control:

**Engine Configuration** (`~/.dazedtrader/algo/engine.json`):
```json
{
  "max_concurrent_strategies": 2,
  "price_update_interval": "15s",
  "emergency_stop_loss": 5.0,      // Stop all trading if 5% total loss
  "daily_loss_limit": 50.0,        // Stop all trading if $50 daily loss
  "enable_backtesting": true
}
```

**Strategy Configuration** (e.g., `~/.dazedtrader/algo/strategies/btc_conservative.json`):
```json
{
  "name": "btc_conservative",
  "symbol": "BTC-USD",
  "enabled": false,                 // Always starts disabled for safety
  "parameters": {
    "short_period": 10.0,           // 10-period moving average
    "long_period": 30.0,            // 30-period moving average
    "min_volume": 2000.0            // Minimum volume threshold
  },
  "risk_limits": {
    "max_position_size": 50.0,      // Max $50 per trade
    "stop_loss_percent": 3.0,       // Auto-sell if 3% loss
    "take_profit_percent": 6.0,     // Auto-sell if 6% gain
    "max_daily_trades": 5,          // Max 5 trades per day
    "max_daily_loss": 25.0,         // Stop strategy if $25 daily loss
    "cooldown_period": "10m"        // 10 min wait between trades
  }
}
```

## 🎯 Usage

### 🤖 Algorithmic Trading Walkthrough

#### Step 1: Initial Setup and Configuration
```bash
# After building the application, configure risk limits first
./configure_algo -preset=ultra_conservative

# Verify configuration
./configure_algo -show
```

**Expected output:**
```
🤖 Current Algorithmic Trading Configuration
═══════════════════════════════════════════
💰 Engine-Wide Limits:
   Emergency Stop Loss: 10.0%
   Daily Loss Limit:    $10.00
   Max Strategies:      2
   Price Update:        15s

📊 Strategy Limits:
1. btc_conservative_ma (BTC-USD) ⚪ Disabled
   Max Position:     $5.00
   Stop Loss:        2.0%
   Daily Trades:     3
   Daily Loss Limit: $5.00
   Cooldown:         15m0s
```

#### Step 2: Start Application and Navigate
```bash
./dazedtrader
```

1. **Login first** with your Robinhood API credentials
2. Navigate to **"🤖 Algorithmic Trading"** (option 3)
3. You'll see the algo trading dashboard

#### Step 3: Understanding the Interface
```
🤖 ALGORITHMIC TRADING

🔧 TRADING ENGINE STATUS
══════════════════════════
Status:            🔴 Stopped          # Engine not running
Active Strategies:  0                   # No strategies active
Active Positions:   0                   # No open positions
Total Trades:       0                   # No trades executed

📊 STRATEGY STATUS
═══════════════════════
Strategy         Symbol    Status      Position    P&L         Trades
────────────────────────────────────────────────────────────────────
1. btc_conservative  BTC-USD   ⚪ Disabled None       $0.00       0

🎮 CONTROLS
═════════════
S - Start Engine    T - Stop Engine     P - Pause/Resume
E - Emergency Stop   1-5 - Toggle Strategy
R - Refresh          Esc - Back to Menu
```

#### Step 4: Enable Your First Strategy
1. **Press `1`** to toggle the BTC strategy
   - Status changes from "⚪ Disabled" to "🔴 Stopped"
   - Strategy is now enabled but engine isn't running

#### Step 5: Start the Trading Engine
1. **Press `S`** to start the engine
   - Status changes to "🟢 Running"
   - Strategy status changes to "🟢 Active"
   - Price monitoring begins

#### Step 6: Monitor Your Strategy
The interface will update showing:
```
📈 ACTIVE STRATEGY DETAILS
════════════════════════════

**btc_conservative (BTC-USD)**
  Last Signal:   None (waiting for signal)
  Position:      None
  Indicators:    MA(10): $43,100.25 MA(30): $42,950.75
```

#### Step 7: Understanding Signals and Trades
When the strategy generates a signal, you'll see:
```
**btc_conservative (BTC-USD)**
  Last Signal:   BUY at $43,250.50 (87.3% confidence)
  Signal Time:   14:23:45
  Position:      0.0015 @ $43,250.50
  Market Value:  $4.85
  Unrealized:    +$0.05
```

#### Step 8: Emergency Controls
- **Press `E`** for immediate emergency stop (cancels all orders, stops all trading)
- **Press `P`** to pause (stops new trades but keeps positions)
- **Press `T`** to gracefully stop the engine
- **Press `1`** again to disable the strategy

### Safety Checklist Before Going Live

#### ✅ Pre-Flight Checklist
- [ ] **API credentials tested** in manual trading first
- [ ] **Risk limits configured** with amounts you can afford to lose
- [ ] **Emergency controls practiced** (know how to press 'E')
- [ ] **Configuration verified** with `./configure_algo -show`
- [ ] **Strategy understood** (read how moving averages work)
- [ ] **Monitoring plan** (check every 30 minutes initially)
- [ ] **Stop-loss understood** (automatic 2-3% loss limit per trade)
- [ ] **Daily limits set** (maximum daily loss you're comfortable with)

#### 🚨 First Day Protocol
1. **Start with ONE strategy only**
2. **Monitor constantly for first 2 hours**
3. **Use smallest preset** (`ultra_conservative`)
4. **Check positions every 30 minutes**
5. **Don't leave unattended** for first few sessions
6. **Practice emergency stop** multiple times
7. **End session with** engine stopped (`T` key)

#### 🎯 What to Expect
- **Few trades initially**: Conservative settings mean fewer signals
- **Small position sizes**: $5 max with ultra_conservative preset
- **Automatic stop losses**: Positions auto-sell at 2% loss
- **Cooldown periods**: 15 minutes between trades minimum
- **Daily limits**: Trading stops after $5 daily loss

### First Launch

When you first run DazedTrader, you'll see the main menu:

```
🚀 DAZED TRADER 🚀
Crypto Trading Terminal

> 🔐 Setup API Key
  ₿ Crypto Portfolio      (Login Required)
  📈 Crypto Trading       (Login Required)
  🤖 Algorithmic Trading  (Login Required)
  📊 Market Data
  📋 Order History        (Login Required)
  📰 Crypto News
  ❓ Help
  🔓 Logout
  🚪 Exit

Status: 🔴 Not Authenticated
Press 'q' to quit • Use ↑↓ to navigate • Enter to select
```

### Authentication Setup

1. **Select "🔐 Setup API Key"**
2. **Enter credentials** in format: `apikey:privatekey`
   - API key from Robinhood
   - Private key in base64 format
3. **Press Enter** to verify credentials
4. **Credentials saved securely** to ~/.config/dazedtrader/

### Main Features

#### 📊 Crypto Portfolio
```
📊 PORTFOLIO DASHBOARD
═════════════════════════════════════════════

📈 PORTFOLIO SUMMARY
═════════════════════
Total Value:     $15,847.32
Day Change:      +$1,247.82 (8.5%)
Buying Power:    $3,250.00

🏆 CRYPTO HOLDINGS
══════════════════
Asset     Quantity    Price      Market Value   Day Change
──────────────────────────────────────────────────────────
BTC       0.1950   $43,250.50     $8,433.85    +$335.20
ETH       5.6800   $2,642.30      $15,008.26   +$891.15
SOL       102.45   $102.45        $10,490.23   +$965.87

📋 RECENT ORDERS
═══════════════
Symbol    Side  Quantity     Avg Price    State
───────────────────────────────────────────────
BTC-USD   BUY   0.0500      $42,890.00   filled
ETH-USD   SELL  0.2500      $2,580.45    filled

Last updated: 2:34 PM
```

#### 📈 Crypto Trading Interface
```
💹 CRYPTO TRADING

📝 **STEP 1: SELECT SYMBOL**

Enter crypto symbol (e.g., BTC-USD, ETH-USD):
BTC-USD│

💰 Current Price: $43,250.50

Popular symbols: BTC-USD, ETH-USD, DOGE-USD, ADA-USD

🔢 **STEP 4: QUANTITY**

Symbol: BTC-USD | Side: BUY | Type: MARKET | Price: $43,250.50

Enter quantity:
0.025│

💰 Estimated Cost: $1,081.26
Available buying power: $3,250.00
```

#### 📊 Market Data
```
📊 CRYPTO MARKET DATA

🚀 TOP GAINERS (24H)
═══════════════════════
Symbol    Price        Change      Volume         Market Cap
──────────────────────────────────────────────────────────────
SOL       $102.45      +9.5%      $2.4B          $45.0B
ETH       $2,642.30    +7.5%      $15.8B         $317.0B
ADA       $0.485       +7.1%      $520.0M        $17.0B

📉 TOP LOSERS (24H)
══════════════════════
Symbol    Price        Change      Volume         Market Cap
──────────────────────────────────────────────────────────────
SHIB      $0.0000095   -7.8%      $245.0M        $5.6B
LINK      $14.82       -7.2%      $890.0M        $8.7B
BCH       $245.60      -7.0%      $320.0M        $4.8B
```

#### 🤖 Algorithmic Trading Interface
```
🤖 ALGORITHMIC TRADING

🔧 TRADING ENGINE STATUS
══════════════════════════
Status:            🟢 Running
Active Strategies:  2
Active Positions:   1
Total Trades:       47

💰 PERFORMANCE SUMMARY
═════════════════════════
Total P&L:         +$247.82
Daily P&L:          +$47.30
Win Rate:           68.1%
Profit Factor:      1.42
Max Drawdown:       3.2%

📊 STRATEGY STATUS
═══════════════════════
Strategy         Symbol    Status      Position    P&L         Trades
────────────────────────────────────────────────────────────────────
1. btc_ma        BTC-USD   🟢 Active   0.0150     +$125.40    12
2. eth_momentum  ETH-USD   🟢 Active   None       +$89.25     8
3. doge_scalping DOGE-USD  ⚪ Disabled None       +$0.00      0

📈 ACTIVE STRATEGY DETAILS
════════════════════════════

**btc_ma (BTC-USD)**
  Last Signal:   BUY at $43,250.50 (87.3% confidence)
  Signal Time:   14:23:45
  Position:      0.0150 @ $42,890.00
  Market Value:  $648.76
  Unrealized:    +$5.41
  Indicators:    MA(5): $43,100.25 MA(20): $42,950.75

**eth_momentum (ETH-USD)**
  Last Signal:   SELL at $2,642.30 (72.1% confidence)
  Signal Time:   14:18:22
  Position:      None
  Indicators:    Mom: +3.45%

🎮 CONTROLS
═════════════
S - Start Engine    T - Stop Engine     P - Pause/Resume
E - Emergency Stop   1-5 - Toggle Strategy
R - Refresh          Esc - Back to Menu

⚠️  CAUTION: Algorithmic trading involves significant risk. Monitor closely!
```

#### Available Trading Strategies

##### 📈 Moving Average Crossover
- **Logic**: Buy when short MA crosses above long MA, sell when below
- **Parameters**: Short period (5), Long period (20), Min volume threshold
- **Best for**: Trending markets, medium-term position holding
- **Risk**: Medium - False signals in sideways markets

##### 🚀 Momentum Trading
- **Logic**: Buy on strong upward momentum with volume confirmation
- **Parameters**: Lookback period (10), Momentum threshold (2%), Volume multiplier (1.5x)
- **Best for**: Volatile markets, quick profits from price momentum
- **Risk**: High - Can get caught in sudden reversals

##### 🔄 Mean Reversion (Template)
- **Logic**: Buy oversold conditions, sell overbought conditions
- **Parameters**: RSI period (14), Oversold threshold (30), Overbought threshold (70)
- **Best for**: Range-bound markets, contrarian trades
- **Risk**: Medium - Can struggle in strong trends

#### 📰 Real-Time Crypto News Feed
```
📰 CRYPTO NEWS

🌐 LATEST CRYPTO NEWS (Live from CryptoCompare API)
═══════════════════════

📈 Bitcoin Surge Continues as Institutional Interest Grows [BTC]
   Major financial institutions are showing increased interest in Bitcoin
   as a store of value amid economic uncertainty and inflation concerns.
   CoinDesk • 45 minutes ago

📊 Ethereum Network Upgrade Improves Transaction Efficiency [ETH]
   The latest Ethereum improvement proposal has been successfully implemented,
   reducing gas costs and improving overall network performance.
   CoinTelegraph • 2 hours ago

📉 Regulatory Concerns Impact Smaller Altcoins [Various]
   New regulatory guidance from major jurisdictions is creating uncertainty
   for several altcoin projects and their compliance strategies.
   CryptoNews • 3 hours ago

💡 Live news updates every 15 minutes from CryptoCompare API
```

#### 📋 Order History
```
📋 ORDER HISTORY

📋 RECENT ORDERS
═══════════════════
Date/Time         Symbol      Side   Type    Quantity      Avg Price    State
─────────────────────────────────────────────────────────────────────────────
2024-01-15 14:23  BTC-USD     BUY    market  0.0500       $42,890.00   filled
2024-01-15 11:45  ETH-USD     SELL   limit   0.2500       $2,580.45    filled
2024-01-14 16:30  DOGE-USD    BUY    market  1000.0000    $0.0825      cancelled

Showing 10 most recent orders
Last updated: 2:34 PM
```

### Keyboard Controls

#### General Controls
| Key | Action |
|-----|--------|
| `↑↓` or `jk` | Navigate menu options |
| `Enter` or `Space` | Select option |
| `Esc` | Go back / Return to main menu |
| `q` or `Ctrl+C` | Quit application |
| `1-9` | Quick menu navigation |
| `r` or `F5` | Refresh current view |

#### Algorithmic Trading Controls
| Key | Action |
|-----|--------|
| `s` | Start trading engine |
| `t` | Stop trading engine |
| `p` | Pause/Resume trading engine |
| `e` | Emergency stop (immediate halt) |
| `1-5` | Toggle individual strategies on/off |
| `r` | Refresh algo trading status |

### Auto-refresh Schedule

- **Portfolio/Orders**: Every 5 seconds
- **Market Data**: Every 30 seconds
- **News Feed**: Every 15 minutes
- **Trading Prices**: Real-time during order placement



### Project Structure
```
DazedTrader/
├── main.go                 # Application entry point
├── cmd/
│   └── configure_algo/
│       └── main.go         # Algorithmic trading configuration utility
├── api/
│   └── crypto_client.go    # Robinhood Crypto API client
├── auth/
│   └── storage.go          # Secure credential storage
├── models/
│   ├── app.go              # Main application model
│   ├── handlers.go         # Input handling and navigation
│   └── views.go            # UI view rendering
├── ui/
│   └── styles.go           # UI styling and formatting
├── algo/                   # Algorithmic trading system
│   ├── engine.go           # Trading engine core
│   ├── strategy_runner.go  # Strategy execution management
│   ├── risk_manager.go     # Risk management system
│   ├── order_manager.go    # Order execution and tracking
│   ├── price_feed.go       # Real-time price monitoring
│   ├── performance_metrics.go # Performance analytics
│   ├── config.go           # Configuration management
│   ├── config_helper.go    # Configuration utilities
│   └── strategies/         # Trading strategy implementations
│       ├── moving_average.go
│       └── momentum.go
├── ALGO_QUICK_START.md     # Quick reference for algorithmic trading
├── .gitignore              # Comprehensive credential protection
├── check_security.sh       # Security audit script
├── go.mod                  # Go module definition
└── README.md
```



## ⚠️ Important Disclaimers

### Security Notice
- This application uses **official Robinhood Crypto APIs**
- All credentials are handled securely and stored locally only
- Review the source code before using with real accounts
- Enable 2FA on your Robinhood account for additional security
- User assumes all risks using this software

### Trading Risks
- **All crypto trading involves significant financial risk**
- **Cryptocurrency markets are highly volatile**
- This software is provided "as-is" without warranties
- Always verify trades and double-check order details
- The developers are not responsible for trading losses

### Algorithmic Trading Risks
- **⚠️ EXTRA HIGH RISK: Algorithmic trading can amplify losses**
- **Strategies may fail in unexpected market conditions**
- **Always start with small position sizes and monitor closely**
- **Never run unattended for extended periods**
- **Use emergency stop controls if markets move against you**
- **Past performance does not guarantee future results**
- **Ensure you understand each strategy before enabling it**
- **Test strategies extensively before using significant capital**

## 🔧 Troubleshooting

### Algorithmic Trading Issues

#### "No strategies configured" or "Configuration files missing"
```bash
# Solution: Run the configuration utility first
./configure_algo -preset=ultra_conservative
./configure_algo -show  # Verify files were created
```

#### "Failed to start trading engine" or "API error"
1. **Check API credentials**: Test manual trading first
2. **Check network connection**: Ensure internet connectivity
3. **Check account permissions**: Verify Robinhood account supports crypto trading
4. **Check buying power**: Ensure sufficient funds in account

#### "Strategy not generating signals"
This is normal! Conservative strategies may wait hours or days for good signals:
- **Moving Average strategy**: Waits for MA crossovers (can take hours)
- **Conservative settings**: Fewer signals = safer trading
- **Market conditions**: Sideways markets generate fewer signals
- **Check indicators**: View MA values in strategy details

#### "Emergency stop triggered unexpectedly"
- **Check daily loss limit**: May have hit $10 limit with ultra_conservative
- **Check total portfolio loss**: May have hit 10% emergency stop
- **Review trade history**: Look for pattern of losses
- **Consider looser limits**: If too restrictive for market conditions

#### "Orders failing to execute"
1. **Check position size**: May be below minimum order size ($1)
2. **Check account status**: Ensure account is in good standing
3. **Check symbol**: Ensure crypto pair is supported (BTC-USD, ETH-USD, etc.)
4. **Check market hours**: Crypto trades 24/7 but API may have brief outages

#### Configuration Help
```bash
# View current settings
./configure_algo -show

# Reset to defaults
./configure_algo -preset=ultra_conservative

# Custom budget
./configure_algo -budget=200 -daily=10

# Help
./configure_algo -help
```

### General Application Issues

#### "Error loading API key" or "Authentication failed"
1. **Check credentials format**: Should be `apikey:privatekey`
2. **Check file permissions**: Credentials stored in `~/.config/dazedtrader/`
3. **Re-enter credentials**: Use "Setup API Key" menu option
4. **Check Robinhood account**: Ensure 2FA is properly configured

#### Application crashes or freezes
1. **Check terminal compatibility**: Requires 256-color terminal support
2. **Check Go version**: Requires Go 1.19 or higher
3. **Check dependencies**: Run `go mod tidy` to update modules
4. **Check system resources**: Ensure sufficient RAM/CPU

#### Build errors
```bash
# Clean and rebuild
rm dazedtrader configure_algo
go clean
go mod tidy
go build -o dazedtrader .
go build -o configure_algo configure_algo.go
```

### Getting Help

1. **Check configuration**: `./configure_algo -show`
2. **Test manual trading**: Verify API credentials work manually first
3. **Start conservatively**: Use `ultra_conservative` preset initially
4. **Read strategy docs**: Understand how moving averages work
5. **Monitor closely**: Don't leave algorithmic trading unattended initially

### Safe Recovery from Issues

If algorithmic trading goes wrong:

1. **Emergency stop**: Press `E` key immediately
2. **Check positions**: Navigate to portfolio to see open positions
3. **Manual intervention**: Use manual trading to close positions if needed
4. **Review logs**: Check what happened before restarting
5. **Adjust limits**: Tighten risk limits before trying again

### API Usage
- Official Robinhood Crypto API with proper authentication
- Rate limiting respects API guidelines
- Use in accordance with Robinhood's Terms of Service
- Real money transactions - use responsibly


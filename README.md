# DazedTrader

A TUI for Robinhood Crypto trading.

## ✨ Features

### 🎯 Core Crypto Trading Functionality
- **Real-time Crypto Portfolio** - Live portfolio with current prices and day changes
- **Interactive Crypto Trading** - 6-step trading interface with live price estimates
- **Order History** - Complete order tracking with status indicators
- **Market Data** - Top gaining/losing cryptocurrencies with real-time data
- **Real-Time Crypto News** - Live news feed from CryptoCompare API with impact analysis
- **Live Price Updates** - Real-time pricing throughout the trading experience


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

# Build the application
go build -o dazedtrader .

# Run the application
./dazedtrader
```

### Security Check (Optional)

```bash
# Run security check to ensure no credentials in source
./check_security.sh
```

## 🎯 Usage

### First Launch

When you first run DazedTrader, you'll see the main menu:

```
🚀 DAZED TRADER 🚀
Crypto Trading Terminal

> 🔐 Setup API Key
  ₿ Crypto Portfolio      (Login Required)
  📈 Crypto Trading       (Login Required)
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

| Key | Action |
|-----|--------|
| `↑↓` or `jk` | Navigate menu options |
| `Enter` or `Space` | Select option |
| `Esc` | Go back / Return to main menu |
| `q` or `Ctrl+C` | Quit application |
| `1-8` | Quick menu navigation |
| `r` or `F5` | Refresh current view |

### Auto-refresh Schedule

- **Portfolio/Orders**: Every 5 seconds
- **Market Data**: Every 30 seconds
- **News Feed**: Every 15 minutes
- **Trading Prices**: Real-time during order placement



### Project Structure
```
DazedTrader/
├── main.go                 # Application entry point
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

### API Usage
- Official Robinhood Crypto API with proper authentication
- Rate limiting respects API guidelines
- Use in accordance with Robinhood's Terms of Service
- Real money transactions - use responsibly


# DazedTrader

> **Advanced Crypto Trading Terminal for Robinhood**

A powerful Terminal User Interface (TUI) application built with Go and Bubble Tea that connects to your Robinhood Crypto account, providing real-time portfolio management, trading capabilities, market data, and news feeds entirely within your terminal.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)
![Bubble Tea](https://img.shields.io/badge/Bubble%20Tea-TUI-FF69B4?style=flat)
![Crypto](https://img.shields.io/badge/Crypto-Trading-F7931E?style=flat&logo=bitcoin)
![License](https://img.shields.io/badge/License-MIT-blue)

## ✨ Features

### 🎯 Core Crypto Trading Functionality
- **Real-time Crypto Portfolio** - Live portfolio with current prices and day changes
- **Interactive Crypto Trading** - 6-step trading interface with live price estimates
- **Order History** - Complete order tracking with status indicators
- **Market Data** - Top gaining/losing cryptocurrencies with real-time data
- **Real-Time Crypto News** - Live news feed from CryptoCompare API with impact analysis
- **Live Price Updates** - Real-time pricing throughout the trading experience

### 🎮 Beautiful Terminal Interface
- **Bubble Tea Framework** - Modern, reactive TUI built with Go
- **Lipgloss Styling** - Professional colors, layouts, and visual indicators
- **Keyboard Navigation** - Intuitive menu navigation and shortcuts
- **Responsive Design** - Adapts to any terminal size
- **Real-time Updates** - Auto-refresh without screen flicker
- **Color-coded Data** - Green/red indicators for gains/losses and buy/sell

### 🔐 Security & Performance
- **Ed25519 Authentication** - Secure API key and private key authentication
- **Local Credential Storage** - Encrypted storage in user's home directory (~/.config/dazedtrader/)
- **No Credential Commits** - Comprehensive .gitignore and security checks
- **HTTPS Only** - All API communication encrypted
- **Secure by Design** - Credentials never stored in source code

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

## 🛠️ Technical Architecture

### Built With
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Modern TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Terminal styling and layout
- **Go 1.24+** - High-performance compiled language
- **Ed25519** - Cryptographic signatures for API authentication

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

### API Integration
- **Authentication**: Ed25519 signature-based authentication
- **Portfolio Data**: `/api/v1/crypto/trading/accounts/`, `/holdings/`
- **Market Data**: `/api/v1/crypto/marketdata/best_bid_ask/`
- **Trading**: `/api/v1/crypto/trading/orders/` for buy/sell operations
- **Rate Limiting**: Efficient polling with appropriate intervals

## 🔒 Security Features

### Credential Protection
- **Local Storage Only**: Credentials stored in `~/.config/dazedtrader/`
- **File Permissions**: 0600 (user read/write only)
- **No Source Code Storage**: Zero risk of committing credentials
- **Comprehensive .gitignore**: Protects all credential patterns
- **Security Audit**: Built-in `check_security.sh` script

### Authentication Security
- **Ed25519 Signatures**: Industry-standard cryptographic authentication
- **API Key Validation**: Real-time credential verification
- **Session Management**: Secure token handling and expiration
- **HTTPS Only**: All communication encrypted in transit

## ⚠️ Important Disclaimers

### Security Notice
- This application uses **official Robinhood Crypto APIs**
- All credentials are handled securely and stored locally only
- Review the source code before using with real accounts
- Enable 2FA on your Robinhood account for additional security

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

## 🤝 Contributing

Contributions are welcome! This project follows standard Go development practices:

```bash
# Fork and clone the repository
git clone https://github.com/yourusername/DazedTrader.git
cd DazedTrader

# Create a feature branch
git checkout -b feature/your-feature-name

# Make changes and test
go build && ./dazedtrader

# Run security check
./check_security.sh

# Run tests
go test ./...

# Submit a pull request
```

### Development Guidelines
- Follow Go conventions and use `go fmt`
- Add tests for new functionality
- Update documentation for user-facing changes
- Test TUI components thoroughly
- Never commit real credentials or API keys
- Run security checks before submitting

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Robinhood Crypto API](https://docs.robinhood.com/crypto/trading/) - Official API documentation

---

**⚡ Built for crypto traders who value speed, security, and beautiful terminal interfaces ⚡**

*Remember: Cryptocurrency trading is highly speculative and involves substantial risk of loss. Only trade with funds you can afford to lose completely.*
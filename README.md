# DazedTrader

> **Advanced Terminal User Interface for Robinhood Trading**

A powerful TUI (Terminal User Interface) application built with Go and Bubble Tea that connects to your Robinhood account, providing real-time portfolio management and trading capabilities entirely within your terminal.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Bubble Tea](https://img.shields.io/badge/Bubble%20Tea-TUI-FF69B4?style=flat)
![License](https://img.shields.io/badge/License-MIT-blue)

## ✨ Features

### 🎯 Core Functionality
- **Real-time Portfolio Dashboard** - Live terminal interface with portfolio data
- **Interactive Trading** - Buy/sell orders directly in the terminal
- **Position Management** - View and manage your stock holdings
- **Order Tracking** - Monitor and cancel pending orders
- **Stock Quotes** - Real-time market data lookup
- **Secure Authentication** - 2FA/MFA support with token management

### 🎮 Terminal User Interface
- **Bubble Tea Framework** - Modern, reactive TUI built with Go
- **Beautiful Styling** - Lipgloss-powered colors and layouts
- **Keyboard Navigation** - Vim-style keybindings (hjkl) and arrow keys
- **Responsive Design** - Adapts to any terminal size
- **Real-time Updates** - Live data refresh without screen flicker
- **Full-screen Mode** - Immersive terminal experience

### 🔐 Security & Performance
- **Go-powered** - Fast, compiled binary with minimal dependencies
- **Local Authentication** - Secure token storage and management
- **HTTPS Only** - All API communication encrypted
- **No Third-party Tracking** - Your data stays between you and Robinhood

## 🚀 Installation & Setup

### Prerequisites
- Go 1.21 or higher
- Terminal with 256 color support
- Robinhood account

### Build from Source
```bash
# Clone the repository
git clone https://github.com/N-Erickson/DazedTrader.git
cd DazedTrader

# Initialize Go modules and download dependencies
go mod tidy

# Build the application
go build -o dazedtrader .

# Run the TUI application
./dazedtrader
```

### Quick Start
```bash
# Launch the main TUI interface
./dazedtrader

# Or test with demo data
./dazedtrader -demo
```

## 🎯 Usage

### TUI Interface Navigation

The application launches in full-screen terminal mode with keyboard navigation:

```
🚀 DAZED TRADER 🚀
Robinhood Terminal Interface

> 🔐 Login to Robinhood
  📊 Portfolio Dashboard  (Login Required)
  💹 Trading Interface    (Login Required)
  📈 View Positions       (Login Required)
  📋 Recent Orders        (Login Required)
  ❓ Help
  🚪 Exit

Status: 🔴 Not Authenticated
Press 'q' to quit • Use ↑↓ to navigate • Enter to select
```

### Keyboard Controls

| Key | Action |
|-----|--------|
| `↑↓` or `jk` | Navigate menu options |
| `Enter` or `Space` | Select option |
| `Esc` | Go back / Return to main menu |
| `q` or `Ctrl+C` | Quit application |
| `1-6` | Quick navigation shortcuts |

### Dashboard Interface

Once authenticated, the dashboard provides:

```
📊 PORTFOLIO DASHBOARD
═════════════════════════════════════════════

📈 PORTFOLIO SUMMARY
═════════════════════
Total Value:     $52,847.32
Day Change:      +$1,247.82 (2.42%)
Buying Power:    $3,250.00

🏆 TOP POSITIONS
═══════════════
Symbol    Shares      Price    Market Value   Day Change
────────────────────────────────────────────────────────
AAPL      50.0000   $185.42      $9,271.00    +$124.50
TSLA      25.0000   $248.73      $6,218.25     -$87.25
MSFT      30.0000   $378.92     $11,367.60    +$215.40
```

## 🛠️ Technical Architecture

### Built With
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Modern TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Terminal styling and layout
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - TUI components
- **Go 1.21+** - High-performance compiled language

### Project Structure
```
DazedTrader/
├── main.go              # Main TUI application & state management
├── api/
│   └── robinhood.go     # Robinhood API client
├── models/
│   ├── auth.go          # Authentication models
│   ├── portfolio.go     # Portfolio data structures
│   └── trading.go       # Trading functionality
├── ui/
│   ├── dashboard.go     # Dashboard view
│   ├── trading.go       # Trading interface
│   └── styles.go        # UI styling definitions
├── utils/
│   ├── config.go        # Configuration management
│   └── storage.go       # Token storage utilities
├── go.mod               # Go module definition
└── README.md
```

### API Integration
- **Authentication**: `/api-token-auth/` with 2FA support
- **Portfolio Data**: `/accounts/`, `/portfolio/`, `/positions/`
- **Market Data**: `/quotes/`, `/instruments/`
- **Trading**: `/orders/` for buy/sell operations
- **Real-time Updates**: Efficient polling with rate limiting

## 🎮 Features Showcase

### 1. Main Menu
- Clean, navigable interface
- Authentication status display
- Keyboard shortcuts
- Context-sensitive options

### 2. Portfolio Dashboard
- Real-time portfolio value tracking
- Day change indicators with color coding
- Top positions display
- Recent order history

### 3. Trading Interface
- Quick buy/sell shortcuts
- Stock quote lookup
- Order management
- Risk indicators

### 4. Position Management
- Detailed holdings view
- Performance metrics
- Sorting and filtering
- Export capabilities

## ⚠️ Important Disclaimers

### Security Notice
- This application uses **unofficial Robinhood APIs**
- All credentials are handled securely and stored locally
- Review the source code before using with real accounts
- Enable 2FA on your Robinhood account for additional security

### Trading Risks
- **All trading involves financial risk**
- This software is provided "as-is" without warranties
- Always verify trades in the official Robinhood app
- The developers are not responsible for trading losses

### API Limitations
- Unofficial API may change without notice
- Rate limiting may apply to API requests
- Some advanced features may not be available
- Use in accordance with Robinhood's Terms of Service

## 🤝 Contributing

Contributions are welcome! This project uses standard Go development practices:

```bash
# Fork the repository and clone your fork
git clone https://github.com/yourusername/DazedTrader.git
cd DazedTrader

# Create a feature branch
git checkout -b feature/your-feature-name

# Make changes and test
go build && ./dazedtrader

# Run tests
go test ./...

# Submit a pull request
```

### Development Guidelines
- Follow Go conventions and use `go fmt`
- Add tests for new functionality
- Update documentation for user-facing changes
- Test TUI components thoroughly

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Charm](https://charm.sh/) - Terminal-focused development tools

---

**⚡ Built for terminal traders who value speed, efficiency, and beautiful interfaces ⚡**

*Remember: Trade responsibly and never invest more than you can afford to lose.*
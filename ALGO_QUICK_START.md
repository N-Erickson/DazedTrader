# 🤖 Algorithmic Trading Quick Start Guide

## 🚀 30-Second Setup

```bash
# 1. Build everything
go mod tidy
go build -o dazedtrader .
go build -o configure_algo cmd/configure_algo/main.go

# 2. Configure ultra-safe limits
./configure_algo -preset=ultra_conservative

# 3. Start and test
./dazedtrader
# → Go to "🤖 Algorithmic Trading"
# → Press '1' to enable BTC strategy
# → Press 'S' to start engine
# → Press 'E' for emergency stop
```

## 💰 Risk Presets

| Preset | Budget | Max/Trade | Daily Loss | Good For |
|--------|--------|-----------|------------|----------|
| `ultra_conservative` | $100 | $5 | $10 | **Learning** |
| `conservative` | $500 | $50 | $25 | **Steady growth** |
| `moderate` | $1000 | $150 | $50 | **Experienced** |

## 🎮 Controls

| Key | Action | When to Use |
|-----|--------|-------------|
| `S` | Start engine | Begin automated trading |
| `T` | Stop engine | End session gracefully |
| `P` | Pause/Resume | Temporary halt |
| `E` | **EMERGENCY STOP** | **Immediate danger** |
| `1-5` | Toggle strategies | Enable/disable individual bots |
| `R` | Refresh | Update display |

## ⚠️ Safety Rules

1. **Always start with `ultra_conservative`**
2. **Test manual trading first**
3. **Monitor for first 2 hours**
4. **Practice emergency stop**
5. **Never leave unattended initially**

## 🛠️ Configuration Commands

```bash
# View current setup
./configure_algo -show

# Custom budget ($300 budget, $15 daily loss limit)
./configure_algo -budget=300 -daily=15

# Reset to safe defaults
./configure_algo -preset=ultra_conservative

# Help
./configure_algo -help
```

## 🚨 Emergency Procedures

**If something goes wrong:**
1. Press `E` (Emergency Stop) immediately
2. Check portfolio for open positions
3. Use manual trading to close positions if needed
4. Review what happened before restarting

**Configuration location:** `~/.dazedtrader/algo/`

---
**⚠️ REMEMBER: Start small, monitor closely, understand the risks!**
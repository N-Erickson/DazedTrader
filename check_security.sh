#!/bin/bash

# Security check script for DazedTrader
# This script checks for potential credential leaks in the codebase

echo "🔍 Checking for potential credential leaks..."

# Check for common credential patterns
echo "Checking for API keys..."
if grep -r -i "sk-" . --exclude-dir=.git --exclude="*.log" --exclude="check_security.sh" 2>/dev/null; then
    echo "⚠️  Found potential OpenAI API key"
fi

if grep -r -i "pk_" . --exclude-dir=.git --exclude="*.log" --exclude="check_security.sh" 2>/dev/null; then
    echo "⚠️  Found potential Stripe API key"
fi

if grep -r -i "AKIA" . --exclude-dir=.git --exclude="*.log" --exclude="check_security.sh" 2>/dev/null; then
    echo "⚠️  Found potential AWS access key"
fi

echo "Checking for private keys..."
if grep -r -i "BEGIN PRIVATE KEY" . --exclude-dir=.git --exclude="*.log" --exclude="check_security.sh" 2>/dev/null; then
    echo "⚠️  Found potential private key"
fi

if grep -r -i "BEGIN RSA PRIVATE KEY" . --exclude-dir=.git --exclude="*.log" --exclude="check_security.sh" 2>/dev/null; then
    echo "⚠️  Found potential RSA private key"
fi

echo "Checking for Ed25519 keys..."
if grep -r -E "[A-Za-z0-9+/]{88}=" . --exclude-dir=.git --exclude="*.log" --exclude="check_security.sh" 2>/dev/null; then
    echo "⚠️  Found potential base64-encoded Ed25519 key"
fi

echo "Checking for hardcoded credentials in Go files..."
if grep -r -i "apikey.*:" *.go **/*.go 2>/dev/null | grep -v "APIKey string" | grep -v "api_key" | grep -v "APIKeyData"; then
    echo "⚠️  Found potential hardcoded API key in Go code"
fi

if grep -r -i "password.*:" *.go **/*.go 2>/dev/null | grep -v "Password string" | grep -v "password field"; then
    echo "⚠️  Found potential hardcoded password in Go code"
fi

if grep -r -i "secret.*:" *.go **/*.go 2>/dev/null | grep -v "Secret string" | grep -v "secret field"; then
    echo "⚠️  Found potential hardcoded secret in Go code"
fi

echo "Checking for TODO comments with credentials..."
if grep -r -i "TODO.*password\|TODO.*api.*key\|TODO.*secret" . --exclude-dir=.git --exclude="*.log" --exclude="check_security.sh" 2>/dev/null; then
    echo "⚠️  Found TODO comments mentioning credentials"
fi

echo "✅ Security check completed."
echo ""
echo "📝 Remember:"
echo "   - Never commit real API keys or private keys"
echo "   - Use environment variables or secure config files"
echo "   - Credentials are stored in ~/.config/dazedtrader/ (outside repo)"
echo "   - Check .gitignore covers all sensitive file patterns"
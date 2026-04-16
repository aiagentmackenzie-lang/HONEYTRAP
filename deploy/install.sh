#!/bin/bash
# HONEYTRAP — Install systemd services
# Usage: sudo bash install.sh

set -euo pipefail

if [[ $EUID -ne 0 ]]; then
  echo "❌ Run as root: sudo bash install.sh"
  exit 1
fi

HONEYTRAP_DIR="${HONEYTRAP_DIR:-/opt/honeytrap}"
HONEYTRAP_USER="honeytrap"

# Create user if not exists
if ! id "$HONEYTRAP_USER" &>/dev/null; then
  useradd -r -s /bin/false -d "$HONEYTRAP_DIR" "$HONEYTRAP_USER"
  echo "✅ Created user: $HONEYTRAP_USER"
fi

# Create directories
mkdir -p "$HONEYTRAP_DIR"/{pcap,exports,logs}
chown -R "$HONEYTRAP_USER:$HONEYTRAP_USER" "$HONEYTRAP_DIR"

# Copy service files
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
for svc in honeytrap honeytrap-api honeytrap-ai; do
  cp "$SCRIPT_DIR/${svc}.service" /etc/systemd/system/
  echo "✅ Installed: ${svc}.service"
done

# Reload and enable
systemctl daemon-reload
for svc in honeytrap honeytrap-api honeytrap-ai; do
  systemctl enable "$svc"
  echo "✅ Enabled: $svc"
done

echo ""
echo "🚀 HONEYTRAP installed. Start with:"
echo "   sudo systemctl start honeytrap honeytrap-api honeytrap-ai"
echo ""
echo "📊 Check status:"
echo "   sudo systemctl status honeytrap"
echo ""
echo "📋 View logs:"
echo "   journalctl -u honeytrap -f"
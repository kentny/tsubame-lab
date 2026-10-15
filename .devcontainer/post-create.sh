#!/usr/bin/env bash
# Runs once when a codespace is created: install the observation tools the course uses
# and build tsubame-lab onto PATH.
set -euo pipefail

sudo apt-get update -qq
sudo apt-get install -y -qq --no-install-recommends \
  curl openssl iproute2 iputils-ping dnsutils tcpdump traceroute netcat-openbsd >/dev/null

go build -o /tmp/tsubame-lab .
sudo install -m 0755 /tmp/tsubame-lab /usr/local/bin/tsubame-lab
rm /tmp/tsubame-lab

tsubame-lab version
echo "tsubame-lab installed: $(command -v tsubame-lab)"

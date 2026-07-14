#!/usr/bin/env bash
set -euo pipefail

echo "Downloading TrashNeurons installer..."
curl -fsSL -o /tmp/trashneurons-installer https://github.com/velez1337fn/trashtalk-neurons/releases/download/v1.0.0/installer
chmod +x /tmp/trashneurons-installer
/tmp/trashneurons-installer
rm -f /tmp/trashneurons-installer

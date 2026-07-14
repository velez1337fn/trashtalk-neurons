#!/usr/bin/env bash
set -e
curl -fsSL https://github.com/velez1337fn/trashtalk-neurons/releases/download/v1.0.0/installer -o /tmp/trashneurons-installer
chmod +x /tmp/trashneurons-installer
/tmp/trashneurons-installer
rm -f /tmp/trashneurons-installer

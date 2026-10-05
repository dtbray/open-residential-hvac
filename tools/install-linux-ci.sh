#!/bin/sh
# SPDX-License-Identifier: AGPL-3.0-only
set -eu
if [ "${HVAC_INSTALL_LOCKED:-}" != 1 ]; then
  exec flock /tmp/hvac-ci-packages.lock env HVAC_INSTALL_LOCKED=1 sh "$0"
fi
if pkg-config --exists gtk+-3.0 webkit2gtk-4.1 2>/dev/null && command -v Xvfb >/dev/null && command -v xdotool >/dev/null && python3 -c 'import pyatspi' 2>/dev/null; then
  exit 0
fi
if [ "$(id -u)" = 0 ]; then SUDO=; else SUDO='sudo -n'; fi
$SUDO apt-get update
DEBIAN_FRONTEND=noninteractive $SUDO apt-get install -y --no-install-recommends \
  build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev \
  xvfb xauth dbus-x11 xdotool python3 python3-pyatspi at-spi2-core unzip

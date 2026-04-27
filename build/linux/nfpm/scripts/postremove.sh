#!/bin/bash
set -e

# Refresh desktop database after removal
if command -v update-desktop-database > /dev/null 2>&1; then
    update-desktop-database /usr/share/applications || true
fi

# Refresh icon cache after removal
if command -v gtk-update-icon-cache > /dev/null 2>&1; then
    gtk-update-icon-cache -f -t /usr/share/icons/hicolor || true
fi

# Refresh MIME database after removal
if command -v update-mime-database > /dev/null 2>&1; then
    update-mime-database /usr/share/mime || true
fi

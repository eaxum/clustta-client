#!/bin/bash
set -e

# Update desktop database so the .desktop file is picked up
if command -v update-desktop-database > /dev/null 2>&1; then
    update-desktop-database /usr/share/applications || true
fi

# Update icon cache
if command -v gtk-update-icon-cache > /dev/null 2>&1; then
    gtk-update-icon-cache -f -t /usr/share/icons/hicolor || true
fi

# Update MIME database for .clst file association
if command -v update-mime-database > /dev/null 2>&1; then
    update-mime-database /usr/share/mime || true
fi

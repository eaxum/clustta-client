#!/bin/bash

# APP_CERTIFICATE="3rd Party Mac Developer Application: Eaxum Creative LTD (XZHBM768JV)"
APP_CERTIFICATE="Apple Distribution: Eaxum Creative LTD (XZHBM768JV)"
PKG_CERTIFICATE="3rd Party Mac Developer Installer: Eaxum Creative LTD (XZHBM768JV)"
APP_NAME="clustta"

cp ./embedded.provisionprofile "./bin/$APP_NAME.app/Contents"

codesign --timestamp --options=runtime -s "$APP_CERTIFICATE" -v --entitlements ./build/darwin/clustta.entitlements "./bin/$APP_NAME.app"

productbuild --sign "$PKG_CERTIFICATE" --component "./bin/$APP_NAME.app" /Applications "./$APP_NAME.pkg"

# codesign --deep --force --sign "-" "./bin/$APP_NAME.app"
#!/usr/bin/env bash
set -euo pipefail

max_sdk_version="${1:-15.4}"
sdk_name_or_path="${2:-/Library/Developer/CommandLineTools/SDKs/MacOSX15.sdk}"

if ! command -v xcrun >/dev/null 2>&1; then
  echo "ERROR: xcrun is required to verify the macOS SDK version." >&2
  exit 1
fi

sdk_version="$(xcrun --sdk "$sdk_name_or_path" --show-sdk-version)"
sdk_path="$(xcrun --sdk "$sdk_name_or_path" --show-sdk-path)"

version_gt() {
  local left="$1"
  local right="$2"
  local left_major left_minor right_major right_minor

  IFS=. read -r left_major left_minor _ <<<"$left"
  IFS=. read -r right_major right_minor _ <<<"$right"
  left_minor="${left_minor:-0}"
  right_minor="${right_minor:-0}"

  if (( left_major > right_major )); then
    return 0
  fi

  if (( left_major == right_major && left_minor > right_minor )); then
    return 0
  fi

  return 1
}

echo "Using macOS SDK $sdk_version at $sdk_path"
echo "Maximum allowed macOS SDK is $max_sdk_version"

if version_gt "$sdk_version" "$max_sdk_version"; then
  cat >&2 <<EOF
ERROR: Selected Xcode macOS SDK $sdk_version is newer than the allowed SDK $max_sdk_version.

Release builds stamped with a newer SDK may be killed before startup on older macOS
versions. Select an older Xcode with xcode-select or raise MAX_MACOS_SDK_VERSION
only after updating the supported macOS baseline.
EOF
  exit 1
fi

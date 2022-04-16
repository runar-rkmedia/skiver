#!/bin/sh

set -e

# Script has been tested on:
# Arch Linux amd64
# Intel Mac
# TODO: Windows

url="http://192.168.10.125:8000/skiver"
platform="$(uname)"
savePath="${HOME}/.local/bin/skiver"
ext=""
where=""

case "$platform" in
  Linux) 
    platform="linux"
  ;;
  Darwin) 
    platform="darwin"
  ;;
  Windows|MINGW*) 
    echo "This script has only partial support on windows."
    platform="windows"
    ext=".exe"
  ;;
  *) echo "unknown platform: ${platform}. Sorry about this. Please see the readme for manuall install-instructions"
    exit 1
  ;;
esac


arch="$(uname -m)"

case "$arch" in
  x86_64) 
    arch="amd64"
  ;;
  arm*) 
    arch="arm64"
  ;;
  *386*) 
    arch="386"
  ;;
  *) echo "unknown architecture: ${arch}. Sorry about this. Please see the readme for manuall install-instructions"
    exit 1
  ;;
esac

if command -v which &> /dev/null;then
    where="$(which skiver 2>/dev/null || true)"
elif command -v whereis &> /dev/null;then
    where="$(whereis skiver | head -1 | sed 's/skiver: *//')"
fi

if [ ! -z "${where}" ];then
  echo "Looks like skiver is already installed at '${where}'"
  savePath="${where}"
fi

url="${url}_${platform}_${arch}${ext}"
echo "Detected you are on   ${platform} (${arch})." 
echo "Downloading from      ${url}"
echo "Saving to             ${savePath}"

tmpFile="skiver${ext}"

curl --fail -Lo "${tmpFile}" "$url"

echo "Download successful"

case "$platform" in
  windows) 
    if [ -z "${where}" ];then
      # I don't really know where the file should be put on windows.
      echo "The executable is downloaded to ${tmpFile}"
      echo "Please move the file to a directory inside your PATH"
      exit 1
    fi
  ;;
esac


echo "Making file executable"

chmod +x "${tmpFile}"

echo "Moving file to ${savePath}"

mv "${tmpFile}" "${savePath}"

echo "Skiver installed successfully"

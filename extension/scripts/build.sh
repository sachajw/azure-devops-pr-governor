#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "Installing dependencies..."
npm ci

echo "Building extension..."
npm run build

echo "Packaging VSIX..."
npx tfx-cli extension create --rev-version

echo "Done. Upload the .vsix to https://marketplace.visualstudio.com/manage/publishers/pangarabbit"

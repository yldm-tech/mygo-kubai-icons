#!/usr/bin/env bash
# Renders icons at points of their hover motion with upstream's code (through rsvg-convert, which draws SVG filters) and with this port (through MyGo), and lays them side by side in out/sheet.png: upstream on the left, the port on the right, one icon per row.
#
#   internal/compare/compare.sh ../kubai-icons folder,bell-x,bookmark 0,0.25,0.5,0.75,1
#
# Needs node, rsvg-convert and ImageMagick.
set -euo pipefail
upstream=$1 icons=$2 points=${3:-0,0.25,0.5,0.75,1}
cd "$(dirname "$0")/../.."
rm -rf out && mkdir -p out
timeout 300 node internal/compare/reference.mjs --upstream "$upstream" --out out/ref "$icons" "$points"
for f in out/ref/*.svg; do rsvg-convert -w 96 -h 96 "$f" -o "${f%.svg}.png"; done
go test -run TestWriteFrames -count=1 ./kubai -args -compare.out "$PWD/out/port" -compare.icons "$icons" -compare.points "$points" >/dev/null
IFS=, read -ra I <<<"$icons"
IFS=, read -ra P <<<"$points"
rows=()
for i in "${I[@]}"; do
	ref=() port=()
	for p in "${P[@]}"; do ref+=("out/ref/$i@$p.png"); port+=("out/port/$i@$p.png"); done
	magick "${ref[@]}" +append +repage out/a.png
	magick "${port[@]}" +append +repage out/b.png
	magick out/a.png -size 32x96 xc:none out/b.png +append +repage "out/row-$i.png"
	rows+=("out/row-$i.png")
done
magick "${rows[@]}" -append +repage -background white -flatten out/sheet.png
rm -f out/a.png out/b.png out/row-*.png
echo out/sheet.png

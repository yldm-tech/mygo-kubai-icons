#!/usr/bin/env bash
# Builds out/motion.gif, this port's icons playing their hover forward and back, from frames TestWriteFrames renders through MyGo.
#
#   internal/compare/motion.sh folder,mail,lock,trash,calendar,bell-x
#
# Needs ImageMagick.
set -euo pipefail
icons=$1
cd "$(dirname "$0")/../.."
points=$(seq -f '%g' 0 0.05 1 | paste -sd, -)
rm -rf out/port && mkdir -p out
go test -run TestWriteFrames -count=1 . -args -compare.out "$PWD/out/port" -compare.icons "$icons" -compare.points "$points" >/dev/null
IFS=, read -ra I <<<"$icons"
IFS=, read -ra P <<<"$points"
frames=()
for p in "${P[@]}"; do
	row=()
	for i in "${I[@]}"; do row+=("out/port/$i@$p.png"); done
	half=$(((${#row[@]} + 1) / 2))
	magick "${row[@]:0:$half}" +append +repage out/top.png
	magick "${row[@]:$half}" +append +repage out/bot.png
	magick out/top.png out/bot.png -append +repage -background white -flatten "out/f-$p.png"
	frames+=("out/f-$p.png")
done
last=${frames[${#frames[@]}-1]}
rev=()
for ((k = ${#frames[@]} - 1; k >= 0; k--)); do rev+=("${frames[$k]}"); done
magick -delay 3 "${frames[@]}" -delay 60 "$last" -delay 3 "${rev[@]}" -delay 60 "${frames[0]}" -loop 0 -layers Optimize out/motion.gif
rm -f out/top.png out/bot.png out/f-*.png
echo out/motion.gif

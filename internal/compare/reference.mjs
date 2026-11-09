// Writes upstream's own rendering of icons at points of their hover motion, as standalone SVGs, for comparing this port against (see compare.sh).
//
//   node internal/compare/reference.mjs --upstream ../kubai-icons --out out/ref folder,bell-x 0,0.5,1
import fs from 'node:fs';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

const args = process.argv.slice(2);
const opt = (n) => args[args.indexOf(`--${n}`) + 1];
const upstream = path.resolve(opt('upstream'));
const out = path.resolve(opt('out'));
const [names, points] = args.filter((a, i) => !a.startsWith('--') && !args[i - 1]?.startsWith('--'));
const load = async (rel) => import(pathToFileURL(path.join(upstream, rel)).href);
const { renderSVG } = await load('packages/core/src/render.js');
const { withBadge, withSlash } = await load('packages/core/src/compose.js');

async function def(name) {
  const file = (n) => path.join(upstream, 'icons', `${n}.js`);
  if (fs.existsSync(file(name))) return (await load(`icons/${name}.js`)).default;
  const cut = name.lastIndexOf('-');
  const base = (await load(`icons/${name.slice(0, cut)}.js`)).default, badge = name.slice(cut + 1);
  if (badge === 'off') return withSlash(base, (await load('badges/_slash.js')).default);
  return withBadge(base, (await load(`badges/${badge}.js`)).default);
}

fs.mkdirSync(out, { recursive: true });
for (const name of names.split(',')) {
  const d = await def(name);
  for (const p of points.split(',').map(Number)) fs.writeFileSync(path.join(out, `${name}@${p}.svg`), renderSVG(d, { progress: p, size: 96 }));
}

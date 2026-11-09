// Generates the Go icon data of this module from a checkout of kubai087/kubai-icons.
//
//   node gen/generate.mjs --upstream ../kubai-icons [--only folder,abacus] [--check]
//
// Each icon is rendered by upstream's own renderInner() at keyframes along its hover motion, so the geometry every icon's draw() computes is reproduced exactly at those keyframes. Keyframes are placed adaptively: an interval is split until straight interpolation of its numbers stays within TOL canvas units of what draw() gives. The Go runtime interpolates between them, fills in the colours and hands the markup to MyGo's SVG renderer.
//
// MyGo does not draw SVG filters, so the drop shadow is written as a second layer, the silhouette of everything upstream shadows, which the Go side blurs. The inner highlight filter is left out.

import fs from 'node:fs';
import zlib from 'node:zlib';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

const TOL = 0.04;
const MIN_DEPTH = 2;
const MAX_DEPTH = 7;
const SEP = '\u0001';
const PHASE = Symbol.for('kubai-icons.phase');

const args = process.argv.slice(2);
const opt = (name, fallback) => {
  const i = args.indexOf(`--${name}`);
  return i >= 0 ? args[i + 1] : fallback;
};
const upstream = path.resolve(opt('upstream', '../kubai-icons'));
const outDir = path.resolve(opt('out', path.join(path.dirname(new URL(import.meta.url).pathname), '..')));
const only = opt('only') ? new Set(opt('only').split(',')) : null;
const check = args.includes('--check');

const load = async (rel) => (await import(pathToFileURL(path.join(upstream, rel)).href));
const { renderInner, parse, serialize } = await load('packages/core/src/render.js');
const { withBadge, withSlash, BADGE_TONE } = await load('packages/core/src/compose.js');
const pkg = JSON.parse(fs.readFileSync(path.join(upstream, 'packages/core/package.json'), 'utf8'));
const pinned = JSON.parse(fs.readFileSync(new URL('./upstream.json', import.meta.url), 'utf8'));
if (pinned.tag !== `kubai-icons@${pkg.version}`) {
  const msg = `${upstream} holds kubai-icons ${pkg.version}, but gen/upstream.json pins ${pinned.tag}`;
  if (check) throw new Error(msg);
  console.warn(`${msg}: update gen/upstream.json to the release generated from.`);
}
const categories = JSON.parse(fs.readFileSync(path.join(upstream, 'icons/categories.json'), 'utf8'));

// ---------- The icon list: bases, then each base's badged and slashed composites ----------
const list = (dir) => fs.readdirSync(path.join(upstream, dir)).filter((n) => n.endsWith('.js') && !n.startsWith('_')).sort();
const bases = [];
for (const n of list('icons')) bases.push((await load(`icons/${n}`)).default);
const badges = new Map();
for (const n of list('badges')) {
  const b = (await load(`badges/${n}`)).default;
  badges.set(b.name, b);
}
const slash = (await load('badges/_slash.js')).default;
const icons = [];
for (const base of bases) {
  icons.push({ def: base, base });
  for (const v of base.variants || []) {
    const badge = badges.get(v);
    if (!badge) throw new Error(`${base.name}: unknown badge "${v}"`);
    icons.push({ def: withBadge(base, badge), base, badge });
  }
  if (base.off) icons.push({ def: withSlash(base, slash), base, badge: slash });
}

// ---------- Markup for one moment of the motion ----------
const walk = (node, fn) => {
  for (const c of node.children) if (typeof c !== 'string') { fn(c); walk(c, fn); }
};
const isShadow = (el) => el.getAttribute('filter') === 'url(#sh-K)';
const hasShadow = (el) => isShadow(el) || el.children.some((c) => typeof c !== 'string' && hasShadow(c));

// Paints everything inside a shadow group in the silhouette colour, as SourceAlpha does.
function silhouette(el) {
  if (typeof el === 'string') return;
  el.removeAttribute('filter');
  for (const a of ['fill', 'stroke']) {
    const v = el.getAttribute(a);
    if (v && v !== 'none') el.setAttribute(a, '{S}');
  }
  const style = el.getAttribute('style');
  if (style) el.setAttribute('style', style.replace(/(fill|stroke):(?!none)[^;]+/g, '$1:{S}'));
  el.children.forEach(silhouette);
}
// Keeps the shadow groups and the groups and transforms around them; drops everything else.
function shadowOnly(node) {
  node.children = node.children.filter((c) => typeof c !== 'string' && (c.tag === 'defs' || hasShadow(c)));
  for (const c of node.children) {
    if (c.tag === 'defs') c.children = c.children.filter((d) => typeof d !== 'string' && (d.tag === 'mask' || d.tag === 'clipPath'));
    else if (isShadow(c)) silhouette(c);
    else shadowOnly(c);
  }
}

// MyGo does not draw SVG text: each <text> becomes a path of glyph outlines from gen/glyphs.json, in its place and with its other attributes.
const glyphs = JSON.parse(fs.readFileSync(new URL('./glyphs.json', import.meta.url), 'utf8'));
const TEXT_ONLY = new Set(['x', 'y', 'text-anchor', 'font-size', 'font-weight', 'font-family']);
function textToPath(node) {
  node.children = node.children.map((el) => {
    if (typeof el === 'string') return el;
    if (el.tag !== 'text') { textToPath(el); return el; }
    const str = el.children.filter((c) => typeof c === 'string').join('');
    const size = +(el.getAttribute('font-size') || 16), k = size / 1000;
    const chars = [...str].map((ch) => {
      const g = glyphs[ch];
      if (!g) throw new Error(`no glyph for "${ch}" in gen/glyphs.json`);
      return g;
    });
    const width = chars.reduce((s, g) => s + g.advance * k, 0);
    const anchor = el.getAttribute('text-anchor');
    let pen = +(el.getAttribute('x') || 0) - (anchor === 'middle' ? width / 2 : anchor === 'end' ? width : 0);
    const y = +(el.getAttribute('y') || 0);
    let d = '';
    for (const g of chars) {
      const x0 = pen;
      d += g.d.replace(/(-?[\d.]+) (-?[\d.]+)/g, (_, gx, gy) => `${f3(x0 + gx * k)} ${f3(y + gy * k)}`);
      pen += g.advance * k;
    }
    el.tag = 'path';
    el.children = [];
    for (const a of TEXT_ONLY) el.removeAttribute(a);
    el.attrs = new Map([['d', d], ...el.attrs]);
    return el;
  });
}
const f3 = (n) => +n.toFixed(3);

function frame(def, v, active, out) {
  globalThis[PHASE] = out ? { out: true, e: 1 - v } : { out: false, e: v };
  try {
    const rendered = parse(renderInner(def, { k: 'K', progress: v, state: active ? { active: true } : {}, paint: 'template', names: false }));
    textToPath(rendered);
    const inner = rendered.children.map(serialize).join('');
    const main = parse(inner);
    walk(main, (el) => {
      el.removeAttribute('filter');
      el.removeAttribute('font-family');
      if (el.tag === 'defs') el.children = el.children.filter((d) => typeof d === 'string' || d.tag !== 'filter');
    });
    const shadow = parse(inner);
    shadowOnly(shadow);
    const s = shadow.children.some((c) => typeof c !== 'string' && c.tag !== 'defs') ? shadow.children.map(serialize).join('') : '';
    return main.children.map(serialize).join('') + SEP + s;
  } finally {
    globalThis[PHASE] = undefined;
  }
}

// ---------- Splitting markup into fixed text and moving numbers ----------
const NUMERIC = new Set(['d', 'transform', 'x', 'y', 'width', 'height', 'cx', 'cy', 'r', 'rx', 'ry', 'x1', 'y1', 'x2', 'y2', 'opacity', 'fill-opacity', 'stroke-opacity', 'stroke-width', 'stroke-dashoffset', 'stroke-dasharray', 'offset', 'points']);
const NUM = /-?(?:\d+\.?\d*|\.\d+)(?:e-?\d+)?/gy;

// Splits path data into text and numbers, keeping arc flags in the text: they cannot be interpolated.
function splitPath(d, text, nums) {
  let cmd = '', argi = 0, i = 0, lit = '';
  while (i < d.length) {
    const ch = d[i];
    if (/[A-Za-z]/.test(ch)) { cmd = ch.toUpperCase(); argi = 0; lit += ch; i++; continue; }
    if (/[-.\d]/.test(ch)) {
      const flag = cmd === 'A' && (argi % 7 === 3 || argi % 7 === 4);
      let tok = ch;
      if (!flag) { NUM.lastIndex = i; tok = NUM.exec(d)[0]; }
      if (flag) lit += tok;
      else { text.push(lit); lit = ''; nums.push(+tok); }
      argi++;
      i += tok.length;
      continue;
    }
    lit += ch;
    i++;
  }
  return lit;
}

function tokenize(markup) {
  const text = [], nums = [];
  let lit = '', last = 0;
  for (const m of markup.matchAll(/([:\w-]+)="([^"]*)"/g)) {
    lit += markup.slice(last, m.index);
    last = m.index + m[0].length;
    const [, name, value] = m;
    if (!NUMERIC.has(name)) { lit += m[0]; continue; }
    lit += `${name}="`;
    if (name === 'd') {
      const sub = [];
      const tail = splitPath(value, sub, nums);
      if (sub.length) { sub[0] = lit + sub[0]; text.push(...sub); lit = tail; } else lit += tail;
    } else {
      let j = 0;
      for (const n of value.matchAll(/-?(?:\d+\.?\d*|\.\d+)(?:e-?\d+)?/g)) {
        text.push(lit + value.slice(j, n.index));
        lit = '';
        nums.push(+n[0]);
        j = n.index + n[0].length;
      }
      lit += value.slice(j);
    }
    lit += '"';
  }
  text.push(lit + markup.slice(last));
  return { text, nums, key: text.join('\u0002') };
}

// ---------- Adaptive keyframes ----------
function track(def, active, out) {
  const cache = new Map();
  const at = (v) => {
    const k = +v.toFixed(6);
    if (!cache.has(k)) cache.set(k, tokenize(frame(def, k, active, out)));
    return cache.get(k);
  };
  const keys = new Set([0, 1]);
  const fits = (a, b) => {
    const A = at(a), B = at(b);
    if (A.key !== B.key) return false;
    for (const q of [0.25, 0.5, 0.75]) {
      const m = at(a + (b - a) * q);
      if (m.key !== A.key) return false;
      for (let i = 0; i < A.nums.length; i++) if (Math.abs(A.nums[i] + (B.nums[i] - A.nums[i]) * q - m.nums[i]) > TOL) return false;
    }
    return true;
  };
  const split = (a, b, depth) => {
    if (depth >= MAX_DEPTH || (depth >= MIN_DEPTH && fits(a, b))) return;
    const m = +((a + b) / 2).toFixed(6);
    keys.add(m);
    split(a, m, depth + 1);
    split(m, b, depth + 1);
  };
  split(0, 1, 0);
  return [...keys].sort((x, y) => x - y).map((v) => ({ v, ...at(v) }));
}

const same = (a, b) => a.length === b.length && a.every((k, i) => k.v === b[i].v && k.key === b[i].key && k.nums.every((n, j) => n === b[i].nums[j]));

// ---------- One icon's data ----------
function build({ def }) {
  const tracks = [];
  for (const active of def.tap ? [false, true] : [false]) {
    const inTrack = track(def, active, false);
    const outTrack = track(def, active, true);
    tracks.push({ active, out: false, keys: inTrack });
    if (!same(inTrack, outTrack)) tracks.push({ active, out: true, keys: outTrack });
  }
  // Templates shared by keyframes. Numbers that never move within a template fold into its text, and numbers that always move together, such as a path and the mask cut from it, share one slot.
  const tpls = new Map();
  for (const t of tracks) for (const k of t.keys) {
    if (!tpls.has(k.key)) tpls.set(k.key, { text: k.text, frames: [] });
    tpls.get(k.key).frames.push(k.nums);
  }
  const out = [];
  for (const tpl of tpls.values()) {
    const n = tpl.frames[0].length;
    const slots = [], bySeries = new Map(), refs = [];
    const text = [tpl.text[0]];
    for (let i = 0; i < n; i++) {
      const series = tpl.frames.map((f) => fmt(f[i]));
      if (series.every((v) => v === series[0])) { text[text.length - 1] += series[0] + tpl.text[i + 1]; continue; }
      const key = series.join(',');
      if (!bySeries.has(key)) { bySeries.set(key, slots.length); slots.push(i); }
      refs.push(bySeries.get(key));
      text.push(tpl.text[i + 1]);
    }
    tpl.index = out.length;
    tpl.slots = slots;
    out.push({ text, refs });
  }
  return {
    templates: out,
    tracks: tracks.map((t) => ({
      active: t.active,
      out: t.out,
      keys: t.keys.map((k) => {
        const tpl = tpls.get(k.key);
        return [+fmt(k.v), tpl.index, ...tpl.slots.map((i) => +fmt(k.nums[i]))];
      })
    }))
  };
}

const fmt = (n) => String(+n.toFixed(3));

// ---------- Go source ----------
const INITIALISMS = new Set(['api', 'ascii', 'cpu', 'css', 'dns', 'eof', 'faq', 'gps', 'gpu', 'html', 'http', 'https', 'id', 'ip', 'json', 'pdf', 'qr', 'ram', 'rss', 'sql', 'ssh', 'tcp', 'tv', 'ui', 'uri', 'url', 'usb', 'vpn', 'xml', 'ai', 'nfc', 'sms']);
function goName(name) {
  const parts = name.split('-');
  let s = '';
  for (let i = 0; i < parts.length; i++) {
    const p = parts[i];
    if (i && /^\d/.test(p) && /\d$/.test(parts[i - 1])) s += '_';
    s += INITIALISMS.has(p) ? p.toUpperCase() : p.charAt(0).toUpperCase() + p.slice(1);
  }
  return /^\d/.test(s) ? 'Icon' + s : s;
}
// Go source has no binary literals; base64 in a raw string is the most compact text form.
const goBlob = (buf) => {
  const b = buf.toString('base64');
  const lines = [];
  for (let i = 0; i < b.length; i += 120) lines.push(b.slice(i, i + 120));
  return '`\n' + lines.join('\n') + '`';
};

// ---------- Families: a base icon and its composites share one compressed blob ----------
const families = new Map();
const names = new Map();
const index = [];
let keyframes = 0, raw = 0, packed = 0;
const started = Date.now();
for (const ic of icons) {
  const { def, base } = ic;
  if (only && !only.has(def.name) && !only.has(base.name)) continue;
  const go = goName(def.name);
  if (names.has(go)) throw new Error(`Go name ${go} of ${def.name} collides with ${names.get(go)}`);
  names.set(go, def.name);
  const data = build(ic);
  keyframes += data.tracks.reduce((s, t) => s + t.keys.length, 0);
  if (!families.has(base.name)) families.set(base.name, { base, members: [] });
  families.get(base.name).members.push({ ic, go, data });
  index.push({ name: def.name, go, base: base.name, category: base.category, tone: def.tone || '', tap: Boolean(base.tap) });
}

const header = `// Code generated by gen/generate.mjs from kubai-icons ${pkg.version}; DO NOT EDIT.\n\npackage kubai\n\n`;
const files = new Map();
for (const [name, fam] of families) {
  const fv = 'f' + goName(name);
  const json = JSON.stringify(Object.fromEntries(fam.members.map(({ ic, data }) => [ic.def.name, { tone: ic.def.tone || '', tap: Boolean(ic.base.tap), templates: data.templates, tracks: data.tracks }])));
  const blob = zlib.deflateRawSync(Buffer.from(json), { level: 9 });
  raw += json.length;
  packed += blob.length;
  let src = '';
  for (const { ic, go } of fam.members) {
    const label = ic.badge ? `${name} icon with ${ic.badge === slash ? 'a slash across it' : `the ${ic.badge.name} badge`}` : `${name} icon`;
    src += `// ${go} is the ${label}: "${ic.def.name}", in the ${fam.base.category} category.${fam.base.tap ? ' A click toggles it; see Options.Active.' : ''}\nfunc ${go}() *Icon {\n\treturn ${fv}.icon(${fv}Data, ${JSON.stringify(ic.def.name)})\n}\n\n`;
  }
  src += `var ${fv} family\n\nconst ${fv}Data = ${goBlob(blob)}\n`;
  const file = /^[a-z]/.test(name) ? name[0] : '0';
  if (!files.has(file)) files.set(file, []);
  files.get(file).push(src);
}

const want = new Map();
for (const [file, srcs] of files) want.set(path.join(outDir, `icons_${file}_gen.go`), header + srcs.join('\n'));
const cats = Array.isArray(categories) ? categories : Object.keys(categories);
const catalog = header + `// Version is the kubai-icons release the icons were generated from.\nconst Version = ${JSON.stringify(pkg.version)}\n\n// Count is the number of icons, composites included.\nconst Count = ${index.length}\n\n// Categories lists the categories icons belong to, in upstream's order.\nvar Categories = []string{${cats.map((c) => JSON.stringify(typeof c === 'string' ? c : c.id)).join(', ')}}\n`;
if (!only) want.set(path.join(outDir, 'version_gen.go'), catalog);
if (!only) want.set(path.join(outDir, 'gen', 'icons.json'), JSON.stringify(index, null, '\t') + '\n');
const registry = `// Code generated by gen/generate.mjs from kubai-icons ${pkg.version}; DO NOT EDIT.\n\n// Package all lists every icon, for galleries and pickers that look icons up by name. Importing it keeps every icon in the executable; a program that names its icons should call their functions in package kubai instead.\npackage all\n\nimport kubai "github.com/yldm-tech/mygo-kubai-icons"\n\n// An Entry is one icon: its upstream name, the name of its function in package kubai, its category and the function.\ntype Entry struct {\n\tName     string\n\tGoName   string\n\tCategory string\n\tIcon     func() *kubai.Icon\n}\n\n// Entries lists every icon, bases in name order, each followed by its composites.\nvar Entries = []Entry{\n${index.map((e) => `\t{${JSON.stringify(e.name)}, ${JSON.stringify(e.go)}, ${JSON.stringify(e.category)}, kubai.${e.go}},\n`).join('')}}\n`;
if (!only) want.set(path.join(outDir, 'all', 'all_gen.go'), registry);

if (check) {
  const stale = [];
  for (const [file, src] of want) if (!fs.existsSync(file) || fs.readFileSync(file, 'utf8') !== src) stale.push(path.relative(outDir, file));
  if (!only) for (const n of fs.readdirSync(outDir)) if (/^icons_.*_gen\.go$/.test(n) && !want.has(path.join(outDir, n))) stale.push(`${n} (not generated any more)`);
  if (stale.length) { console.error(`Generated files are stale:\n  ${stale.join('\n  ')}`); process.exit(1); }
  console.log(`${want.size} generated files are up to date.`);
} else {
  if (!only) for (const n of fs.readdirSync(outDir)) if (/^icons_.*_gen\.go$/.test(n)) fs.rmSync(path.join(outDir, n));
  for (const [file, src] of want) { fs.mkdirSync(path.dirname(file), { recursive: true }); fs.writeFileSync(file, src); }
  console.log(`${index.length} icons in ${families.size} families, ${keyframes} keyframes, ${(raw / 1e6).toFixed(1)} MB of data packed into ${(packed / 1e6).toFixed(2)} MB, in ${((Date.now() - started) / 1000).toFixed(1)}s.`);
}

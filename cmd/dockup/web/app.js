(() => {
'use strict';

// =====================================================================
//  DOM helpers
// =====================================================================
// Element.append() stringifies arrays and null ("[object HTMLDivElement]",
// "null"). The UI composes optional and list children everywhere, so append
// is made to flatten arrays and skip null/undefined/false, like h() does.
(() => {
  const native = Element.prototype.append;
  Element.prototype.append = function (...kids) {
    const out = [];
    const walk = (k) => { if (k === null || k === undefined || k === false) return; if (Array.isArray(k)) k.forEach(walk); else out.push(k); };
    kids.forEach(walk);
    return native.apply(this, out);
  };
})();
const SVGNS = 'http://www.w3.org/2000/svg';
const ICONS = {
  logo: '<polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>',
  grid: '<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/>',
  folder: '<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>',
  terminal: '<polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>',
  server: '<rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/>',
  key: '<circle cx="7.5" cy="15.5" r="5.5"/><path d="m21 2-9.6 9.6"/><path d="m15.5 7.5 3 3L22 7l-3-3"/>',
  users: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
  sliders: '<line x1="4" y1="21" x2="4" y2="14"/><line x1="4" y1="10" x2="4" y2="3"/><line x1="12" y1="21" x2="12" y2="12"/><line x1="12" y1="8" x2="12" y2="3"/><line x1="20" y1="21" x2="20" y2="16"/><line x1="20" y1="12" x2="20" y2="3"/><line x1="1" y1="14" x2="7" y2="14"/><line x1="9" y1="8" x2="15" y2="8"/><line x1="17" y1="16" x2="23" y2="16"/>',
  list: '<line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/>',
  search: '<circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>',
  moon: '<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/>',
  plus: '<line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>',
  rocket: '<path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z"/><path d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z"/><path d="M9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0"/><path d="M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5"/>',
  refresh: '<polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>',
  download: '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>',
  pause: '<rect x="6" y="4" width="4" height="16" rx="1"/><rect x="14" y="4" width="4" height="16" rx="1"/>',
  stop: '<rect x="5" y="5" width="14" height="14" rx="2"/>',
  trash: '<polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>',
  history: '<path d="M3 3v5h5"/><path d="M3.05 13A9 9 0 1 0 6 5.3L3 8"/><path d="M12 7v5l4 2"/>',
  columns: '<rect x="3" y="3" width="18" height="18" rx="2"/><line x1="9" y1="3" x2="9" y2="21"/><line x1="15" y1="3" x2="15" y2="21"/>',
  webhook: '<path d="M18 16.98h-5.99c-1.1 0-1.95.94-2.48 1.9A4 4 0 0 1 2 17c.01-.7.2-1.4.57-2"/><path d="m6 17 3.13-5.78c.53-.97.1-2.18-.5-3.1a4 4 0 1 1 6.89-4.06"/><path d="m12 6 3.13 5.73C15.66 12.7 16.9 13 18 13a4 4 0 0 1 0 8"/>',
  check: '<polyline points="20 6 9 17 4 12"/>',
  x: '<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>',
  loader: '<path d="M21 12a9 9 0 1 1-6.22-8.56"/>',
  down: '<polyline points="6 9 12 15 18 9"/>',
  right: '<polyline points="9 18 15 12 9 6"/>',
  left: '<line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/>',
  copy: '<rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>',
  eye: '<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>',
  eyeoff: '<path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/>',
  alert: '<path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>',
  info: '<circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>',
  git: '<circle cx="18" cy="18" r="3"/><circle cx="6" cy="6" r="3"/><path d="M13 6h3a2 2 0 0 1 2 2v7"/><line x1="6" y1="9" x2="6" y2="21"/>',
  clock: '<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>',
  logout: '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>',
  user: '<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>',
  shield: '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>',
  more: '<circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/><circle cx="5" cy="12" r="1"/>',
  edit: '<path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4z"/>',
  arrowdown: '<line x1="12" y1="5" x2="12" y2="19"/><polyline points="19 12 12 19 5 12"/>',
  globe: '<circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>',
  lock: '<rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>',
  play: '<polygon points="6 3 20 12 6 21 6 3"/>',
  rollback: '<polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>',
  power: '<path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><line x1="12" y1="2" x2="12" y2="12"/>',
  box: '<path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/>',
  activity: '<polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>',
  mail: '<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/>',
};

function icon(name, cls) {
  const s = document.createElementNS(SVGNS, 'svg');
  s.setAttribute('viewBox', '0 0 24 24');
  s.setAttribute('class', 'i' + (cls ? ' ' + cls : ''));
  s.setAttribute('aria-hidden', 'true');
  s.innerHTML = ICONS[name] || ''; // constant markup only
  return s;
}

// The Dockup mark: a "D" with an up-arrow cut out (D + up), on a gradient tile.
const BRAND_GLYPH = 'M20 13h12a19 19 0 0 1 0 38H20a3 3 0 0 1-3-3V16a3 3 0 0 1 3-3zM31 21.5 20.5 32H28v11h6V32h7.5z';
function brandMark(cls) {
  const tile = document.createElement('span');
  tile.className = 'logo-mark' + (cls ? ' ' + cls : '');
  tile.setAttribute('aria-hidden', 'true');
  const s = document.createElementNS(SVGNS, 'svg');
  s.setAttribute('viewBox', '0 0 64 64');
  const p = document.createElementNS(SVGNS, 'path');
  p.setAttribute('fill', '#fff'); p.setAttribute('fill-rule', 'evenodd'); p.setAttribute('d', BRAND_GLYPH);
  s.append(p); tile.append(s);
  return tile;
}
const wordmark = (cls) => h('span', { class: 'wordmark' + (cls ? ' ' + cls : '') }, 'dock', h('span', {}, 'up'));

function h(tag, attrs, ...kids) {
  const el = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs || {})) {
    if (v === null || v === undefined || v === false) continue;
    if (k.startsWith('on')) el.addEventListener(k.slice(2), v);
    else if (k === 'class') el.className = v;
    else if (k === 'style' && typeof v === 'object') Object.assign(el.style, v);
    else if (k === 'value' || k === 'checked' || k === 'disabled' || k === 'selected' || k === 'indeterminate') el[k] = v;
    else if (k === 'dataset') Object.assign(el.dataset, v);
    else el.setAttribute(k, v === true ? '' : v);
  }
  const add = (c) => {
    if (c === null || c === undefined || c === false) return;
    if (Array.isArray(c)) return c.forEach(add);
    el.append(c.nodeType ? c : document.createTextNode(String(c)));
  };
  kids.forEach(add);
  return el;
}
const clear = (el) => { while (el.firstChild) el.removeChild(el.firstChild); return el; };
const $ = (sel, root) => (root || document).querySelector(sel);

// =====================================================================
//  API
// =====================================================================
class ApiError extends Error { constructor(msg, status) { super(msg); this.status = status; } }
async function api(method, path, body) {
  const r = await fetch(path, {
    method, credentials: 'same-origin', cache: 'no-store',
    headers: { 'X-Dockup': '1', ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}) },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  const ct = r.headers.get('content-type') || '';
  const data = ct.includes('json') ? await r.json().catch(() => null) : await r.text();
  if (r.status === 401 && path !== '/api/login' && path !== '/api/boot') { boot.user = null; renderAuth(); throw new ApiError('signed out', 401); }
  if (!r.ok) throw new ApiError((data && data.error) || r.statusText || ('HTTP ' + r.status), r.status);
  return data;
}

// =====================================================================
//  formatting
// =====================================================================
const T0 = '0001-01-01T00:00:00Z';
const isZero = (t) => !t || t === T0 || t.startsWith('0001-');
function ago(t) {
  if (isZero(t)) return 'never';
  const s = Math.round((Date.now() - new Date(t).getTime()) / 1000);
  if (s < 0) return 'in ' + dur(-s * 1000);
  if (s < 10) return 'just now';
  if (s < 60) return s + 's ago';
  if (s < 3600) return Math.floor(s / 60) + 'm ago';
  if (s < 86400) return Math.floor(s / 3600) + 'h ago';
  if (s < 172800) return 'yesterday';
  if (s < 30 * 86400) return Math.floor(s / 86400) + 'd ago';
  return new Date(t).toLocaleDateString();
}
function dur(ms) {
  if (ms < 0 || !isFinite(ms)) return '–';
  const s = Math.round(ms / 1000);
  if (s < 60) return s + 's';
  const m = Math.floor(s / 60);
  if (m < 60) return m + 'm ' + String(s % 60).padStart(2, '0') + 's';
  return Math.floor(m / 60) + 'h ' + String(m % 60).padStart(2, '0') + 'm';
}
function buildDur(b) {
  if (isZero(b.started)) return null;
  const end = isZero(b.finished) ? Date.now() : new Date(b.finished).getTime();
  return end - new Date(b.started).getTime();
}
const fmtDate = (t) => isZero(t) ? '–' : new Date(t).toLocaleString(undefined, { year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit' });
const short = (c) => (c || '').slice(0, 7);
const initials = (u) => ((u.name || u.login || '?').split(/[\s._@-]+/).filter(Boolean).slice(0, 2).map((s) => s[0]).join('') || '?').toUpperCase();
const STATUS = {
  success: ['ok', 'Success'], failed: ['bad', 'Failed'], running: ['run', 'Running'], queued: ['warn', 'Queued'], cancelled: ['mute', 'Cancelled'],
};
function badge(status) {
  const [c, label] = STATUS[status] || ['mute', status];
  return h('span', { class: 'badge ' + c }, status === 'running' ? icon('loader', 'spin') : null, label);
}
function dot(status) {
  const c = { success: 'ok', failed: 'bad', running: 'run pulse', queued: 'warn pulse' }[status] || '';
  return h('span', { class: 'dot ' + c });
}
const stc = (s) => ({ success: 'okc', failed: 'badc', running: 'runc', queued: 'warnc' }[s] || 'mute');
const TRIGGER = { manual: 'manual', webhook: 'webhook', schedule: 'cron', scm: 'git poll' };

// =====================================================================
//  state, roles, cleanup
// =====================================================================
let boot = { providers: [] };
let summary = null; // last /api/dashboard
let cleanups = [];
const can = (role) => boot.user && ({ viewer: 1, operator: 2, admin: 3 }[boot.user.role] || 0) >= ({ viewer: 1, operator: 2, admin: 3 }[role] || 9);
function onCleanup(fn) { cleanups.push(fn); }
function every(ms, fn) {
  const id = setInterval(() => { if (!document.hidden) fn(); }, ms);
  onCleanup(() => clearInterval(id));
}
function runCleanups() { cleanups.forEach((f) => { try { f(); } catch (e) { /* ignore */ } }); cleanups = []; }

// =====================================================================
//  toasts, menus, modals
// =====================================================================
function toast(msg, kind) {
  kind = kind || 'ok';
  const el = h('div', { class: 'toast ' + kind, role: 'status' },
    icon(kind === 'ok' ? 'check' : kind === 'bad' ? 'alert' : 'info'), h('div', { class: 'grow' }, msg),
    h('span', { class: 'x', onclick: () => el.remove() }, icon('x')));
  $('#toasts').append(el);
  setTimeout(() => el.remove(), kind === 'bad' ? 8000 : 4000);
}
const fail = (e) => { if (e && e.status !== 401) toast(e.message || String(e), 'bad'); };

async function copy(text, what) {
  try { await navigator.clipboard.writeText(text); toast((what || 'Copied') + ' to clipboard'); }
  catch (e) {
    const ta = h('textarea', { style: { position: 'fixed', opacity: '0' } }, text);
    document.body.append(ta); ta.select();
    try { document.execCommand('copy'); toast((what || 'Copied') + ' to clipboard'); } catch (e2) { toast('Copy failed: select and copy manually', 'bad'); }
    ta.remove();
  }
}
const copyBtn = (text, what) => h('button', { class: 'btn sm', type: 'button', onclick: (e) => { e.stopPropagation(); copy(typeof text === 'function' ? text() : text, what); } }, icon('copy'), 'Copy');

let openMenuEl = null;
function closeMenu() { if (openMenuEl) { openMenuEl.remove(); openMenuEl = null; } }
function menu(anchor, items) {
  closeMenu();
  const m = h('div', { class: 'menu', role: 'menu' });
  for (const it of items) {
    if (!it) continue;
    if (it === '-') { m.append(h('hr')); continue; }
    if (it.header) { m.append(h('div', { class: 'mh' }, it.header)); continue; }
    m.append(h('button', { type: 'button', class: it.danger ? 'danger' : '', onclick: () => { closeMenu(); it.run(); } }, it.icon ? icon(it.icon) : null, it.label));
  }
  document.body.append(m);
  const r = anchor.getBoundingClientRect();
  const mw = m.offsetWidth, mh = m.offsetHeight;
  let left = r.right - mw; if (left < 8) left = r.left;
  let top = r.bottom + 4; if (top + mh > window.innerHeight - 8) top = r.top - mh - 4;
  m.style.left = Math.max(8, left) + 'px'; m.style.top = Math.max(8, top) + 'px';
  openMenuEl = m;
  setTimeout(() => document.addEventListener('click', closeMenu, { once: true }), 0);
}
window.addEventListener('scroll', closeMenu, true);
window.addEventListener('resize', closeMenu);

function modal(opts) {
  // opts: title, sub, body (node|array), actions: [{label, kind, run(close) -> bool|Promise}], wide, onclose
  const prev = document.activeElement;
  const ov = h('div', { class: 'overlay', onmousedown: (e) => { if (e.target === ov && !opts.sticky) close(); } });
  const close = () => { ov.remove(); document.removeEventListener('keydown', esc); if (opts.onclose) opts.onclose(); if (prev && prev.focus) prev.focus(); };
  const esc = (e) => { if (e.key === 'Escape' && !opts.sticky) close(); };
  document.addEventListener('keydown', esc);
  const foot = h('div', { class: 'modal-f' });
  const errEl = h('div', { class: 'err hidden' });
  const setBusy = (b) => foot.querySelectorAll('button').forEach((x) => { x.disabled = b; });
  for (const a of (opts.actions || [{ label: 'Close', run: () => true }])) {
    const btn = h('button', { type: a.submit ? 'submit' : 'button', class: 'btn ' + (a.kind || '') }, a.icon ? icon(a.icon) : null, a.label);
    btn.addEventListener('click', async (e) => {
      e.preventDefault();
      errEl.classList.add('hidden');
      try {
        setBusy(true);
        const res = a.run ? await a.run(close) : true;
        if (res !== false) close();
      } catch (err) {
        if (err && err.status === 401) return close();
        errEl.textContent = err.message || String(err); errEl.classList.remove('hidden');
      } finally { setBusy(false); }
    });
    foot.append(btn);
  }
  const form = h('form', { class: 'modal' + (opts.wide ? ' wide' : ''), role: 'dialog', 'aria-modal': 'true', onsubmit: (e) => {
    e.preventDefault(); const p = foot.querySelector('.btn.primary,.btn.go,.btn.danger.solid'); if (p) p.click();
  } },
  h('div', { class: 'modal-h' }, h('div', {}, h('h3', {}, opts.title), opts.sub ? h('div', { class: 'sub' }, opts.sub) : null),
    h('button', { type: 'button', class: 'icon-btn', 'aria-label': 'Close', onclick: close }, icon('x'))),
  h('div', { class: 'modal-b' }, opts.body, errEl), foot);
  ov.append(form);
  document.body.append(ov);
  const first = form.querySelector('.modal-b input:not([type=checkbox]):not([readonly]),.modal-b select,.modal-b textarea');
  if (first) setTimeout(() => first.focus(), 30);
  return { close, el: form, foot };
}

function confirmDialog(title, text, opts) {
  opts = opts || {};
  return new Promise((resolve) => {
    let done = false;
    const typed = opts.typeToConfirm ? h('input', { class: 'inp mono', placeholder: opts.typeToConfirm, autocomplete: 'off' }) : null;
    modal({
      title, body: [h('div', { class: 'sub', style: { marginTop: 0 } }, text),
        typed ? h('label', { class: 'f' }, h('span', {}, 'Type ', h('b', { class: 'mono' }, opts.typeToConfirm), ' to confirm'), typed) : null],
      onclose: () => { if (!done) resolve(false); },
      actions: [{ label: 'Cancel', run: () => true },
        { label: opts.ok || 'Confirm', kind: opts.danger ? 'danger solid' : 'primary', run: () => {
          if (typed && typed.value !== opts.typeToConfirm) throw new Error('Confirmation text does not match');
          done = true; resolve(true); return true;
        } }],
    });
  });
}

// =====================================================================
//  theme
// =====================================================================
function theme() { return document.documentElement.getAttribute('data-theme') || 'dark'; }
function toggleTheme() {
  const t = theme() === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', t);
  try { localStorage.setItem('dockup-theme', t); } catch (e) { /* storage blocked */ }
  document.querySelectorAll('.theme-t').forEach((b) => { clear(b).append(icon(t === 'dark' ? 'sun' : 'moon')); });
}
const themeBtn = () => h('button', { class: 'icon-btn theme-t', type: 'button', title: 'Toggle theme', 'aria-label': 'Toggle theme', onclick: toggleTheme }, icon(theme() === 'dark' ? 'sun' : 'moon'));

// =====================================================================
//  auth screen
// =====================================================================
function renderAuth() {
  runCleanups();
  const app = clear($('#app'));
  const q = new URLSearchParams(location.search).get('error');
  if (q) history.replaceState(null, '', location.pathname + location.hash);
  let mode = boot.needSetup ? 'setup' : 'signin';
  const card = h('div', { class: 'card' });
  const draw = () => {
    clear(card);
    if (boot.needSetup && boot.providers.length) { // sign-in only makes sense for an invited SSO admin
      card.append(h('div', { class: 'seg', style: { width: '100%' } },
        h('button', { type: 'button', class: mode === 'signin' ? 'on' : '', style: { flex: 1, justifyContent: 'center' }, onclick: () => { mode = 'signin'; draw(); } }, 'Sign in'),
        h('button', { type: 'button', class: mode === 'setup' ? 'on' : '', style: { flex: 1, justifyContent: 'center' }, onclick: () => { mode = 'setup'; draw(); } }, 'First-run setup')));
    }
    if (q) card.append(h('div', { class: 'callout bad' }, icon('alert'), h('div', {}, q)));
    if (mode === 'setup') card.append(setupForm()); else card.append(signinForm());
  };
  draw();
  app.append(h('div', { class: 'auth' }, themeBtn(),
    h('div', { class: 'auth-card' },
      h('div', { class: 'auth-head' }, brandMark('xl'),
        h('div', {}, h('h1', {}, boot.needSetup ? 'Set up Dockup' : 'Sign in to Dockup'),
          h('div', { class: 'sub' }, 'Self-hosted CI/CD and Docker Compose control plane'))),
      card,
      h('div', { class: 'auth-foot' }, h('span', {}, 'Instance: ', boot.instance || location.host),
        h('span', {}, '•'), boot.tls ? h('span', { class: 'okc' }, 'TLS') : h('span', { class: 'warnc', title: 'Put a TLS reverse proxy in front and set DOCKUP_URL' }, 'not encrypted (http)'),
        h('span', {}, '•'), h('span', {}, 'v' + (boot.version || 'dev'))))));
}

function signinForm() {
  const login = h('input', { class: 'inp', autocomplete: 'username', required: true, autofocus: true });
  const pw = h('input', { class: 'inp', type: 'password', autocomplete: 'current-password', required: true });
  const err = h('div', { class: 'err' });
  const btn = h('button', { class: 'btn primary', type: 'submit', style: { height: '36px', width: '100%' } }, 'Continue with password');
  const f = h('form', { class: 'stack', onsubmit: async (e) => {
    e.preventDefault(); err.textContent = ''; btn.disabled = true;
    try { await api('POST', '/api/login', { login: login.value.trim(), password: pw.value }); await start(); }
    catch (x) { err.textContent = x.message; pw.select(); } finally { btn.disabled = false; }
  } },
  h('label', { class: 'f' }, 'Username', login), h('label', { class: 'f' }, 'Password', pw), err, btn);
  const out = [f];
  if (boot.providers.length) {
    out.push(h('div', { class: 'or' }, 'or'));
    out.push(h('div', { class: 'sso' }, boot.providers.map((p) => h('a', { class: 'btn', href: '/auth/' + encodeURIComponent(p.id) + '/login' },
      icon(p.kind === 'github' ? 'git' : p.kind === 'google' ? 'globe' : 'shield'), 'Continue with ' + p.name))));
  }
  return out;
}

function setupForm() {
  const code = h('input', { class: 'inp mono', placeholder: 'XXXX-XXXX-XXXX', autocomplete: 'off', required: true, autofocus: true });
  const login = h('input', { class: 'inp', autocomplete: 'username', required: true, value: 'admin' });
  const name = h('input', { class: 'inp', autocomplete: 'name' });
  const email = h('input', { class: 'inp', type: 'email', autocomplete: 'email' });
  const pw = h('input', { class: 'inp', type: 'password', autocomplete: 'new-password', required: true, minlength: '10' });
  const pw2 = h('input', { class: 'inp', type: 'password', autocomplete: 'new-password', required: true });
  const err = h('div', { class: 'err' });
  const btn = h('button', { class: 'btn primary', type: 'submit', style: { height: '36px', width: '100%' } }, 'Create admin account', icon('right'));
  return h('form', { class: 'stack', onsubmit: async (e) => {
    e.preventDefault(); err.textContent = '';
    if (pw.value !== pw2.value) { err.textContent = 'Passwords do not match'; return; }
    if (pw.value.length < 10) { err.textContent = 'Use at least 10 characters'; return; }
    btn.disabled = true;
    try {
      await api('POST', '/api/setup', { code: code.value, login: login.value.trim(), name: name.value.trim(), email: email.value.trim(), password: pw.value });
      history.replaceState(null, '', location.pathname + '#/welcome');
      await start();
    } catch (x) { err.textContent = x.message; } finally { btn.disabled = false; }
  } },
  h('div', { class: 'callout info' }, icon('lock'), h('div', {}, 'Enter the one-time setup code printed in the hub log (',
    h('code', {}, 'journalctl -u dockup'), ' or ', h('code', {}, 'docker logs dockup'), ') and stored in ', h('code', {}, '<data dir>/setup-code'), '.')),
  h('label', { class: 'f' }, 'Setup code', code),
  h('div', { class: 'fgrid' }, h('label', { class: 'f' }, 'Username', login), h('label', { class: 'f' }, 'Display name', name)),
  h('label', { class: 'f' }, h('span', {}, 'Email ', h('span', { class: 'hint' }, '(optional, lets you link SSO later)')), email),
  h('div', { class: 'fgrid' }, h('label', { class: 'f' }, 'Password', pw), h('label', { class: 'f' }, 'Confirm', pw2)),
  h('div', { class: 'hint' }, 'At least 10 characters. A passphrase works well.'), err, btn,
  h('div', { class: 'hint', style: { textAlign: 'center' } }, 'Next: a two-minute guided setup to get your first stack live.'));
}

// =====================================================================
//  shell
// =====================================================================
const NAV = [
  { h: 'Views' },
  { id: 'dashboard', href: '#/', icon: 'grid', label: 'Dashboard' },
  { id: 'projects', href: '#/projects', icon: 'folder', label: 'Projects', n: () => summary && summary.stats.projects },
  { id: 'builds', href: '#/builds', icon: 'terminal', label: 'Builds & Logs', live: () => summary && summary.stats.running > 0 },
  { id: 'servers', href: '#/servers', icon: 'server', label: 'Servers', n: () => summary && (summary.stats.online + '/' + summary.stats.servers), ok: true },
  { h: 'Administration', role: 'admin' },
  { id: 'credentials', href: '#/credentials', icon: 'key', label: 'Credentials', role: 'admin' },
  { id: 'users', href: '#/users', icon: 'users', label: 'Users & Roles', role: 'admin' },
  { id: 'settings', href: '#/settings', icon: 'sliders', label: 'System Settings', role: 'admin' },
  { id: 'audit', href: '#/audit', icon: 'list', label: 'Audit Log', role: 'admin' },
];

let shellEls = null;
function shell(active) {
  const nav = h('div');
  let cur = null;
  for (const n of NAV) {
    if (n.role && !can(n.role)) continue;
    if (n.h) { nav.append(h('div', { class: 'nav-h' }, n.h)); cur = h('nav', { class: 'nav' }); nav.append(cur); continue; }
    const count = n.n ? n.n() : null;
    cur.append(h('a', { href: n.href, class: active === n.id ? 'on' : '' }, icon(n.icon), h('span', {}, n.label),
      n.live && n.live() ? h('span', { class: 'n' }, h('span', { class: 'dot run pulse' })) :
        count !== null && count !== undefined ? h('span', { class: 'n' + (n.ok ? ' ok' : '') }, count) : null));
  }
  const cap = h('div', { class: 'capacity' });
  if (summary) {
    const st = summary.stats; const pct = st.servers ? Math.round(st.online / st.servers * 100) : 0;
    const off = summary.servers.filter((s) => !s.online).map((s) => s.name);
    cap.append(h('div', { class: 'row' }, h('span', {}, 'Agents online'), h('b', { class: off.length ? 'warnc' : 'okc' }, st.online + '/' + st.servers)),
      h('div', { class: 'bar' }, h('i', { style: { width: pct + '%', background: off.length ? 'var(--warn)' : 'var(--ok)' } })),
      h('div', { class: 'row', style: { fontSize: '10px' } }, h('span', {}, st.running + ' running · ' + st.queued + ' queued'),
        off.length ? h('span', { class: 'badc ellip', style: { maxWidth: '90px' } }, off[0] + ' off') : null));
  }
  const q = summary ? summary.stats.queued + summary.stats.running : 0;
  const u = boot.user;
  const userBtn = h('button', { class: 'avatar', type: 'button', title: u.name || u.login, onclick: (e) => {
    e.stopPropagation();
    menu(e.currentTarget, [{ header: (u.name || u.login) + (u.email ? ' · ' + u.email : '') },
      { label: 'Account & API tokens', icon: 'user', run: () => go('#/account') },
      can('admin') ? { label: 'Getting started guide', icon: 'play', run: async () => { await setDismissed(false); go('#/welcome'); } } : null,
      { label: theme() === 'dark' ? 'Light theme' : 'Dark theme', icon: theme() === 'dark' ? 'sun' : 'moon', run: toggleTheme },
      '-', { label: 'Sign out', icon: 'logout', run: signOut }]);
  } }, initials(u));
  const top = h('header', { class: 'top' },
    h('div', { class: 'left' }, h('a', { class: 'brand', href: '#/', 'aria-label': 'Dockup home' }, brandMark(), wordmark()),
      h('div', { class: 'vsep' }), h('div', { class: 'instance' }, h('b', {}, boot.instance || location.host), h('span', { class: 'tag' }, 'v' + boot.version))),
    h('button', { class: 'search-btn', type: 'button', onclick: openPalette }, h('span', {}, icon('search'), 'Jump to project, build, server…'), h('kbd', {}, isMac ? '⌘K' : 'Ctrl K')),
    h('div', { class: 'right' },
      h('a', { class: 'pill', href: '#/builds', title: 'Running and queued builds' }, h('span', { class: 'dot sm ' + (q ? 'warn pulse' : 'ok') }), q ? q + ' active' : 'idle'),
      themeBtn(), h('div', { class: 'vsep' }), h('span', { class: 'role ' + u.role }, u.role), userBtn));
  const main = h('main', { class: 'main' });
  const app = clear($('#app'));
  app.append(top, h('div', { class: 'frame' }, h('aside', { class: 'side' }, nav, cap), main));
  shellEls = { main, active };
  return main;
}
const isMac = /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent);

async function refreshSummary() {
  try { summary = await api('GET', '/api/dashboard'); } catch (e) { /* keep last */ }
  return summary;
}

function pageHeader(crumbs, title, opts) {
  opts = opts || {};
  return h('div', { class: 'ph' },
    h('div', { class: 'grow' }, crumbs && crumbs.length ? h('div', { class: 'crumbs' }, crumbs.map((c, i) => [i ? h('span', {}, '/') : null, c.href ? h('a', { href: c.href }, c.label) : h('span', {}, c.label)])) : null,
      h('h1', {}, title, opts.badge || null), opts.sub ? h('div', { class: 'sub' }, opts.sub) : null, opts.meta || null),
    opts.actions ? h('div', { class: 'actions' }, opts.actions) : null);
}

async function signOut() {
  try { await api('POST', '/api/logout'); } catch (e) { /* ignore */ }
  boot.user = null; summary = null; location.hash = '#/'; renderAuth();
}

// =====================================================================
//  command palette
// =====================================================================
function openPalette() {
  if ($('.palette')) return;
  const input = h('input', { placeholder: 'Type a project, build number, server or action…', autocomplete: 'off', 'aria-label': 'Search' });
  const res = h('div', { class: 'res' });
  const ov = h('div', { class: 'overlay', onmousedown: (e) => { if (e.target === ov) close(); } });
  const close = () => { ov.remove(); };
  let items = [], sel = 0;
  const all = () => {
    const out = [];
    const nav = [['Dashboard', '#/', 'grid'], ['Projects', '#/projects', 'folder'], ['Builds & Logs', '#/builds', 'terminal'], ['Servers', '#/servers', 'server'], ['Account & API tokens', '#/account', 'user']];
    if (can('admin')) nav.push(['Credentials', '#/credentials', 'key'], ['Users & Roles', '#/users', 'users'], ['System Settings', '#/settings', 'sliders'], ['Audit Log', '#/audit', 'list']);
    if (can('admin')) out.push({ g: 'Actions', label: 'New compose project', icon: 'plus', run: () => go('#/projects/new') }, { g: 'Actions', label: 'Add deployment server', icon: 'server', run: () => { go('#/servers'); setTimeout(addServerFlow, 50); } });
    out.push({ g: 'Actions', label: 'Toggle light/dark theme', icon: 'moon', run: toggleTheme });
    for (const p of (summary ? summary.projects : [])) {
      out.push({ g: 'Projects', label: p.name, sub: p.server, icon: 'folder', run: () => go('#/projects/' + p.id) });
      if (can('operator') && !p.disabled) out.push({ g: 'Deploy', label: 'Deploy ' + p.name + (p.server ? ' to ' + p.server : ''), icon: 'rocket', run: () => deployFlow(p) });
    }
    for (const b of (summary ? summary.recent : [])) out.push({ g: 'Recent builds', label: (b.project || '?') + ' #' + b.number + ' (' + b.status + ')', icon: 'terminal', run: () => go('#/builds/' + b.id) });
    for (const s of (summary ? summary.servers : [])) out.push({ g: 'Servers', label: s.name, sub: s.online ? 'online' : 'offline', icon: 'server', run: () => go('#/servers') });
    for (const [l, href, ic] of nav) out.push({ g: 'Go to', label: l, icon: ic, run: () => go(href) });
    return out;
  };
  const source = all();
  const draw = () => {
    const q = input.value.trim().toLowerCase();
    const words = q.split(/\s+/).filter(Boolean);
    items = source.filter((it) => words.every((w) => (it.label + ' ' + (it.sub || '') + ' ' + it.g).toLowerCase().includes(w))).slice(0, 40);
    if (sel >= items.length) sel = 0;
    clear(res);
    if (!items.length) { res.append(h('div', { class: 'empty' }, 'No matches')); return; }
    let g = null;
    items.forEach((it, i) => {
      if (it.g !== g) { g = it.g; res.append(h('div', { class: 'grp-t' }, g)); }
      const el = h('div', { class: 'it' + (i === sel ? ' sel' : ''), onmousemove: () => { if (sel !== i) { sel = i; draw(); } }, onclick: () => { close(); it.run(); } },
        icon(it.icon), h('span', { class: 'ellip' }, it.label), it.sub ? h('span', { class: 'faint m' }, it.sub) : null, h('span', { class: 'r' }, i === sel ? '↵' : ''));
      res.append(el);
      if (i === sel) setTimeout(() => el.scrollIntoView({ block: 'nearest' }), 0);
    });
  };
  input.addEventListener('input', () => { sel = 0; draw(); });
  input.addEventListener('keydown', (e) => {
    if (e.key === 'ArrowDown') { sel = Math.min(items.length - 1, sel + 1); draw(); e.preventDefault(); }
    else if (e.key === 'ArrowUp') { sel = Math.max(0, sel - 1); draw(); e.preventDefault(); }
    else if (e.key === 'Enter') { const it = items[sel]; if (it) { close(); it.run(); } }
    else if (e.key === 'Escape') close();
  });
  ov.append(h('div', { class: 'modal palette' }, h('div', { class: 'q' }, icon('search'), input, h('kbd', {}, 'ESC')), res));
  document.body.append(ov);
  draw(); input.focus();
}
document.addEventListener('keydown', (e) => {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k' && boot.user) { e.preventDefault(); openPalette(); }
  if (e.key === '/' && boot.user && !/INPUT|TEXTAREA|SELECT/.test(document.activeElement.tagName) && !$('.overlay')) { e.preventDefault(); openPalette(); }
});

// =====================================================================
//  deploy actions
// =====================================================================
const ACTIONS = {
  deploy: { label: 'Deploy', icon: 'rocket' }, restart: { label: 'Restart', icon: 'refresh' }, pull: { label: 'Pull images', icon: 'download' },
  stop: { label: 'Stop', icon: 'pause' }, down: { label: 'Down (remove containers)', icon: 'stop' },
};
async function trigger(pid, action, params, commit) {
  const b = await api('POST', '/api/projects/' + pid + '/builds', { action, params: params || undefined, commit: commit || undefined });
  toast((ACTIONS[action] ? ACTIONS[action].label : action) + ' queued as #' + b.number);
  return b;
}
async function deployFlow(p, action) {
  action = action || 'deploy';
  try {
    const full = await api('GET', '/api/projects/' + p.id);
    if (action === 'deploy' && full.params && full.params.length) return paramsDialog(full);
    if (action === 'stop' || action === 'down') {
      const ok = await confirmDialog((action === 'down' ? 'Take down ' : 'Stop ') + full.name + '?',
        action === 'down' ? 'Runs docker compose down: containers and networks of this stack are removed (volumes are kept).' : 'Runs docker compose stop: containers are stopped but kept.',
        { danger: true, ok: action === 'down' ? 'Take down' : 'Stop' });
      if (!ok) return;
    }
    const b = await trigger(full.id, action);
    go('#/builds/' + b.id);
  } catch (e) { fail(e); }
}
function paramsDialog(p, preset) {
  const inputs = {};
  const rows = (p.params || []).map((d) => {
    let el;
    const val = preset && preset[d.name] !== undefined && preset[d.name] !== '****' ? preset[d.name] : d.default;
    if (d.choices) {
      el = h('select', { class: 'inp' }, d.choices.split(',').map((c) => c.trim()).map((c) => h('option', { value: c, selected: c === val }, c)));
    } else {
      el = h('input', { class: 'inp mono', type: d.secret ? 'password' : 'text', value: val || '', autocomplete: 'off' });
    }
    inputs[d.name] = el;
    return h('label', { class: 'f' }, h('span', { class: 'flex' }, h('span', { class: 'mono' }, d.name), d.secret ? h('span', { class: 'badge mute' }, icon('lock'), 'secret') : null),
      el, d.help ? h('span', { class: 'hint' }, d.help) : null);
  });
  modal({
    title: 'Deploy ' + p.name + ' with parameters', sub: 'Values are exported as environment variables for compose interpolation and hooks.',
    body: rows,
    actions: [{ label: 'Cancel', run: () => true }, { label: 'Deploy', kind: 'go', icon: 'rocket', run: async () => {
      const params = {}; for (const k in inputs) params[k] = inputs[k].value;
      const b = await trigger(p.id, 'deploy', params); go('#/builds/' + b.id);
    } }],
  });
}
function actionMenu(anchor, p) {
  const items = [{ header: p.name }];
  if (can('operator') && !p.disabled) {
    items.push({ label: 'Deploy', icon: 'rocket', run: () => deployFlow(p) });
    if (p.hasParams) items.push({ label: 'Deploy with parameters…', icon: 'sliders', run: () => deployFlow(p) });
    items.push({ label: 'Restart', icon: 'refresh', run: () => deployFlow(p, 'restart') }, { label: 'Pull images', icon: 'download', run: () => deployFlow(p, 'pull') },
      { label: 'Stop', icon: 'pause', run: () => deployFlow(p, 'stop') }, { label: 'Down', icon: 'stop', danger: true, run: () => deployFlow(p, 'down') }, '-');
  }
  items.push({ label: 'Open project', icon: 'folder', run: () => go('#/projects/' + p.id) });
  if (can('admin')) items.push({ label: 'Configure', icon: 'edit', run: () => go('#/projects/' + p.id + '/edit') });
  menu(anchor, items);
}

// =====================================================================
//  shared widgets
// =====================================================================
function healthPill(hist) {
  const last = (hist || []).slice(-5);
  if (!last.length) return h('span', { class: 'health none', title: 'No finished builds yet' }, '–');
  const ok = last.filter((s) => s === 'success').length;
  const c = ok === last.length ? 'ok' : ok >= Math.ceil(last.length / 2) ? 'warn' : 'bad';
  return h('span', { class: 'health ' + c, title: ok + ' of the last ' + last.length + ' builds passed' }, ok + '/' + last.length);
}
function histBars(hist, running) {
  const arr = (hist || []).slice(-12);
  const cells = [];
  for (let i = arr.length; i < (running ? 11 : 12); i++) cells.push(h('i'));
  arr.slice(running ? -11 : -12).forEach((s) => cells.push(h('i', { class: s, title: s })));
  if (running) cells.push(h('i', { class: 'running' }));
  return h('span', { class: 'hist' }, cells);
}
function emptyState(ic, title, text, action) {
  return h('div', { class: 'empty' }, icon(ic), h('b', {}, title), h('div', {}, text), action ? h('div', { class: 'mt12' }, action) : null);
}

function projectTable(projects, opts) {
  opts = opts || {};
  const filter = (opts.filter || '').toLowerCase();
  const list = projects.filter((p) => !filter || (p.name + ' ' + (p.folder || '') + ' ' + (p.server || '') + ' ' + (p.source || '')).toLowerCase().includes(filter));
  if (!projects.length) {
    return h('div', { class: 'tbl-wrap' }, emptyState('box', 'No projects yet', 'A project is one Docker Compose stack on one server, from a path or a git repository.',
      can('admin') ? h('div', { class: 'flex', style: { justifyContent: 'center' } }, h('a', { class: 'btn primary', href: '#/welcome/project' }, icon('play'), 'Guided setup'),
        h('a', { class: 'btn', href: '#/projects/new' }, icon('plus'), 'New project')) : null));
  }
  const groups = new Map();
  list.forEach((p) => { const k = p.folder || ''; if (!groups.has(k)) groups.set(k, []); groups.get(k).push(p); });
  const keys = [...groups.keys()].sort((a, b) => (a === '') - (b === '') || a.localeCompare(b));
  const tb = h('tbody');
  const showFolders = keys.length > 1 || (keys.length === 1 && keys[0] !== '');
  const collapsed = opts.collapsed || new Set();
  for (const k of keys) {
    const ps = groups.get(k).sort((a, b) => a.name.localeCompare(b.name));
    if (showFolders) {
      tb.append(h('tr', { class: 'folder', onclick: () => { collapsed.has(k) ? collapsed.delete(k) : collapsed.add(k); if (opts.redraw) opts.redraw(); } },
        h('td', { colspan: '8' }, icon(collapsed.has(k) ? 'right' : 'folder'), k || 'Ungrouped', h('span', { class: 'c' }, '(' + ps.length + (ps.length === 1 ? ' stack)' : ' stacks)')))));
      if (collapsed.has(k)) continue;
    }
    for (const p of ps) {
      const run = p.running, last = p.last;
      const lastCell = run ? h('a', { class: 'runc m', href: '#/builds/' + run.id, style: { fontWeight: 600 } }, '#' + run.number, ' ', icon('loader', 'spin'), run.status === 'queued' ? ' queued' : '')
        : last ? h('span', {}, h('a', { class: stc(last.status) + ' m', href: '#/builds/' + last.id, style: { fontWeight: 600 } }, '#' + last.number), h('span', { class: 'faint m', style: { marginLeft: '6px', fontSize: '10.5px' } }, ago(last.finished))) : h('span', { class: 'faint m' }, 'never');
      const d = run ? buildDur(run) : last ? buildDur(last) : null;
      const durCell = h('span', { class: 'm ' + (run ? 'runc' : last && last.status === 'failed' ? 'badc' : 'mute') }, d === null ? '–' : dur(d) + (run ? '…' : last && last.status === 'failed' ? ' (failed)' : ''));
      let act;
      if (run) act = h('a', { class: 'btn xs primary', href: '#/builds/' + run.id }, 'Live log');
      else if (can('operator') && !p.disabled) {
        act = h('div', { class: 'split' }, h('button', { class: 'btn xs', type: 'button', onclick: () => deployFlow(p) }, last && last.status === 'failed' ? 'Re-run' : 'Deploy'),
          h('button', { class: 'btn xs', type: 'button', 'aria-label': 'More actions', onclick: (e) => { e.stopPropagation(); actionMenu(e.currentTarget, p); } }, icon('down')));
      } else act = h('button', { class: 'btn xs ghost', type: 'button', onclick: (e) => { e.stopPropagation(); actionMenu(e.currentTarget, p); } }, icon('more'));
      tb.append(h('tr', { class: run ? 'running' : '' },
        h('td', { class: 'c' }, healthPill(p.history)),
        h('td', {}, h('div', { class: 'flex gap4' }, h('a', { class: 'cell-title', href: '#/projects/' + p.id }, p.name),
          p.disabled ? h('span', { class: 'badge mute' }, 'disabled') : null, run ? h('span', { class: 'badge run' }, run.status) : null),
          h('span', { class: 'cell-sub', title: p.source }, p.source)),
        h('td', {}, h('span', { class: 'tag' }, p.server || '?')),
        h('td', { class: 'nowrap' }, lastCell), h('td', { class: 'nowrap' }, durCell), h('td', {}, histBars(p.history, !!run)),
        h('td', { class: 'c' }, h('span', { class: 'm faint', style: { fontSize: '10.5px' } }, p.trigger)),
        h('td', { class: 'num' }, act)));
      if (run && run.status === 'running') tb.append(h('tr', {}, h('td', { colspan: '8', style: { padding: 0, border: 0 } }, h('div', { class: 'progress' }, h('i')))));
    }
  }
  if (!list.length) tb.append(h('tr', {}, h('td', { colspan: '8' }, h('div', { class: 'empty' }, 'No projects match “' + opts.filter + '”'))));
  return h('div', { class: 'tbl-wrap' }, h('table', {},
    h('thead', {}, h('tr', {}, h('th', { class: 'c', title: 'Health: passed of the last 5 builds' }, 'H'), h('th', {}, 'Name'), h('th', {}, 'Server'), h('th', {}, 'Last build'),
      h('th', {}, 'Duration'), h('th', {}, 'History (12)'), h('th', { class: 'c' }, 'Trigger'), h('th', { class: 'num' }, 'Action'))), tb));
}

function agentCard(s) {
  const busy = s.running && s.running.length;
  return h('div', { class: 'agent' + (s.online ? '' : ' off') },
    h('div', { class: 't' }, h('b', {}, h('span', { class: 'dot ' + (s.online ? 'ok' : 'bad') }), s.name),
      h('span', { class: 'badge ' + (s.online ? 'ok' : 'bad') }, s.local ? 'hub · local' : s.online ? 'online' : 'offline')),
    h('dl', {}, h('dt', {}, 'Host'), h('dd', { title: s.hostname }, s.hostname || '–'), h('dt', {}, s.online ? 'Docker' : 'Last seen'),
      h('dd', { class: s.online ? '' : 'badc' }, s.online ? (s.docker ? 'v' + s.docker : 'not found') : ago(s.lastSeen))),
    h('div', { class: 'f' }, h('span', {}, s.projects + (s.projects === 1 ? ' stack' : ' stacks')),
      busy ? h('span', { class: 'runc ellip', title: s.running.join(', ') }, h('span', { class: 'dot sm run pulse' }), ' ', s.running[0]) : h('span', {}, s.online ? 'idle' : isZero(s.lastSeen) ? 'never connected' : 'heartbeat lost')));
}

function feedItem(b) {
  const d = buildDur(b);
  const sub = b.status === 'failed' && b.reason ? b.reason : (b.commit ? 'Commit ' + short(b.commit) + ' · ' : '') + (TRIGGER[b.trigger] || b.trigger) + (b.by ? ' by ' + b.by : '') + ' · ' + b.server;
  return h('a', { class: 'feed-item', href: '#/builds/' + b.id },
    h('div', { class: 't' }, h('span', { class: stc(b.status) + ' ellip' }, (b.project || '?') + ' #' + b.number + (b.action !== 'deploy' ? ' · ' + b.action : '')),
      h('span', { class: stc(b.status), style: { fontWeight: 400, fontSize: '10.5px' } }, b.status === 'success' ? 'passed' : b.status, d !== null ? ' (' + dur(d) + ')' : '')),
    h('div', { class: 's' }, sub + ' · ' + ago(b.queued)));
}

// =====================================================================
//  pages: dashboard & projects
// =====================================================================
async function pageDashboard() {
  const main = shell('dashboard');
  const collapsed = new Set();
  let filter = '';
  const statsSlot = h('div', { class: 'grid g4' }), agentsSlot = h('div', { class: 'grid g4' }), tableSlot = h('div'), sideSlot = h('div', { class: 'stack' });
  const agentsHead = h('a', { href: '#/servers' });
  const filterInp = h('input', { class: 'inp sm', placeholder: 'Filter projects…', style: { width: '190px' }, oninput: () => { filter = filterInp.value; drawTable(); } });
  const gsSlot = h('div');
  main.append(pageHeader([{ label: 'overview' }], 'Deploy Dashboard', {
    actions: can('admin') ? [h('a', { class: 'btn go', href: '#/projects/new' }, icon('plus'), 'New compose project')] : null }),
  gsSlot, statsSlot,
  h('section', {}, h('div', { class: 'section-h' }, h('h2', {}, icon('server'), 'Deployment agents'), agentsHead), agentsSlot),
  h('div', { class: 'cols' },
    h('div', {}, h('div', { class: 'section-h' }, h('div', { class: 'flex' }, h('h2', {}, 'Projects & stacks'), h('span', { class: 'faint m', style: { fontSize: '11px' } }, '(grouped by folder)')), filterInp), tableSlot),
    sideSlot));
  const drawTable = () => { clear(tableSlot).append(projectTable(summary.projects, { filter, collapsed, redraw: drawTable })); };
  const draw = () => {
    const d = summary, st = d.stats;
    const fin = d.week.reduce((a, x) => a + x.success + x.failed, 0), okc = d.week.reduce((a, x) => a + x.success, 0);
    const rate = fin ? (okc / fin * 100) : null;
    const maxDay = Math.max(1, ...d.week.map((x) => x.success + x.failed));
    const folders = new Set(d.projects.map((p) => p.folder).filter(Boolean)).size;
    clear(statsSlot).append(
      h('div', { class: 'stat' }, h('div', { class: 'k' }, 'Active projects'), h('div', { class: 'v' }, h('b', {}, d.projects.filter((p) => !p.disabled).length),
        h('span', {}, folders ? 'in ' + folders + ' folder' + (folders > 1 ? 's' : '') : d.projects.length + ' total'))),
      h('div', { class: 'stat' }, h('div', { class: 'k' }, 'Executors online'), h('div', { class: 'v' }, h('b', { class: st.online < st.servers ? 'warnc' : 'okc' }, st.online),
        h('span', {}, '/ ' + st.servers + ' server' + (st.servers > 1 ? 's' : '')))),
      h('div', { class: 'stat' }, h('div', { class: 'k' }, 'Deployments today'), h('div', { class: 'v' }, h('b', {}, st.today),
        h('span', { class: st.running ? 'runc m' : 'm' }, st.running ? st.running + ' executing' : st.queued ? st.queued + ' queued' : 'idle'))),
      h('div', { class: 'stat' }, h('div', { class: 'k' }, h('span', {}, 'Pass rate (7d)'), h('b', { class: rate === null ? 'faint' : rate >= 90 ? 'okc' : rate >= 70 ? 'warnc' : 'badc' }, rate === null ? '–' : rate.toFixed(1) + '%')),
        h('div', { class: 'spark' }, d.week.map((x, i) => {
          const n = x.success + x.failed;
          const dd = new Date(); dd.setDate(dd.getDate() - (6 - i));
          return h('i', { class: n === 0 ? 'none' : x.failed ? 'bad' : '', style: { height: (n ? Math.max(18, n / maxDay * 100) : 8) + '%' },
            title: dd.toLocaleDateString(undefined, { weekday: 'short' }) + ': ' + x.success + ' passed, ' + x.failed + ' failed' });
        }))));
    agentsHead.textContent = 'View all servers (' + d.servers.length + ') →';
    clear(agentsSlot).append(d.servers.slice(0, 8).map(agentCard));
    clear(sideSlot).append(queueCard(d), recentCard(d));
    drawTable();
  };
  await refreshSummary();
  if (!summary) { main.append(emptyState('alert', 'Could not load the dashboard', 'Check that the hub is running.')); return; }
  const drawGs = async () => {
    if (!can('admin')) return;
    const ob = await getOnboarding();
    clear(gsSlot);
    if (!ob || ob.dismissed) return;
    // first visit of a brand-new hub goes straight to the guide
    let seen = false;
    try { seen = sessionStorage.getItem('dockup-guide-seen') === '1'; sessionStorage.setItem('dockup-guide-seen', '1'); } catch (e) { seen = true; }
    if (!seen && summary.projects.length === 0) return go('#/welcome');
    if (ob.steps.every((s) => s.done)) return;
    gsSlot.append(gettingStarted(ob, drawGs));
  };
  await drawGs();
  if (location.hash.startsWith('#/welcome')) return;
  refreshNav();
  draw();
  every(4000, async () => { await refreshSummary(); if (summary) { draw(); refreshNav(); } });
}
function refreshNav() {
  if (!shellEls) return;
  const main = shellEls.main;
  const scroll = window.scrollY;
  shell(shellEls.active);
  $('.main').replaceWith(main);
  shellEls.main = main;
  window.scrollTo(0, scroll);
}
function queueCard(d) {
  const q = [];
  for (const p of d.projects) if (p.running) q.push(Object.assign({ project: p.name, server: p.server }, p.running));
  q.sort((a, b) => (a.status === 'running' ? 0 : 1) - (b.status === 'running' ? 0 : 1));
  return h('div', { class: 'card pad' }, h('div', { class: 'section-h' }, h('h2', {}, h('span', { class: 'dot ' + (q.length ? 'warn' : 'ok') }), 'Build queue'),
    h('span', { class: 'badge ' + (q.length ? 'warn' : 'mute') }, q.length ? q.length + ' active' : 'empty')),
  q.length ? h('div', { class: 'feed' }, q.map((b) => h('div', { class: 'feed-item' },
    h('div', { class: 't' }, h('a', { href: '#/builds/' + b.id, class: 'ellip' }, b.project + ' #' + b.number), h('span', { class: stc(b.status) }, b.status)),
    h('div', { class: 's' }, b.status === 'queued' ? 'Waiting for ' + (b.server || 'server') + ' · ' + ago(b.queued) : 'On ' + b.server + ' · ' + dur(buildDur(b) || 0)),
    can('operator') ? h('div', { class: 'flex between', style: { marginTop: '3px' } }, h('span', { class: 's' }, (TRIGGER[b.trigger] || b.trigger) + (b.by ? ' · ' + b.by : '')),
      h('button', { class: 'btn xs danger', type: 'button', onclick: async () => { try { await api('POST', '/api/builds/' + b.id + '/cancel'); toast('Cancel requested'); } catch (e) { fail(e); } } }, 'Cancel')) : null)))
    : h('div', { class: 'faint', style: { fontSize: '12px' } }, 'Nothing running. Deploys, webhooks and schedules show up here.'));
}
function recentCard(d) {
  return h('div', { class: 'card pad' }, h('div', { class: 'section-h' }, h('h2', {}, h('span', { class: 'dot ok pulse' }), 'Recent deployments'), h('a', { href: '#/builds' }, 'All builds →')),
    d.recent.length ? h('div', { class: 'feed' }, d.recent.slice(0, 8).map(feedItem)) : h('div', { class: 'faint', style: { fontSize: '12px' } }, 'No builds yet.'));
}

async function pageProjects() {
  const main = shell('projects');
  let filter = '';
  const collapsed = new Set();
  const slot = h('div');
  const filterInp = h('input', { class: 'inp', placeholder: 'Filter by name, folder, server, repo…', oninput: () => { filter = filterInp.value; draw(); } });
  main.append(pageHeader([{ label: 'projects' }], 'Projects', { sub: 'Each project is one Docker Compose stack on one server.',
    actions: can('admin') ? [h('a', { class: 'btn go', href: '#/projects/new' }, icon('plus'), 'New compose project')] : null }),
  h('div', { class: 'filters' }, filterInp), slot);
  const draw = () => { if (summary) clear(slot).append(projectTable(summary.projects, { filter, collapsed, redraw: draw })); };
  await refreshSummary(); draw();
  every(5000, async () => { await refreshSummary(); draw(); });
}

// =====================================================================
//  project page
// =====================================================================
const STAGE_ORDER = ['poll', 'checkout', 'prepare', 'pre-deploy', 'pull', 'build', 'up', 'restart', 'stop', 'down', 'health', 'post-deploy'];
const STAGE_LABEL = { checkout: 'Checkout', prepare: 'Prepare', 'pre-deploy': 'Pre hook', pull: 'Pull', build: 'Build', up: 'Compose up', restart: 'Restart', stop: 'Stop', down: 'Down', health: 'Health check', 'post-deploy': 'Post hook', poll: 'Poll' };
const stageDur = (s) => isZero(s.end) ? Date.now() - new Date(s.start).getTime() : new Date(s.end).getTime() - new Date(s.start).getTime();

async function pageProject(id, tab) {
  const main = shell('projects');
  let p, builds;
  try { [p, builds] = await Promise.all([api('GET', '/api/projects/' + id), api('GET', '/api/builds?project=' + id + '&limit=200')]); }
  catch (e) { main.append(emptyState('alert', 'Project not found', e.message)); return; }
  await refreshSummary(); refreshNav();
  const sv = summary && summary.servers.find((s) => s.id === p.serverId);
  const active = builds.find((b) => b.status === 'running' || b.status === 'queued');
  const finished = builds.filter((b) => b.status === 'success' || b.status === 'failed');
  const lastOk = finished.find((b) => b.status === 'success'), lastBad = finished.find((b) => b.status === 'failed');
  const src = p.source.kind === 'git' ? p.source.repo + ' @ ' + p.source.branch : p.source.path;
  const op = can('operator') && !p.disabled;
  const actions = [];
  if (active) actions.push(h('a', { class: 'btn primary', href: '#/builds/' + active.id }, icon('terminal'), 'Live log #' + active.number));
  if (op) {
    actions.push(h('button', { class: 'btn go', type: 'button', onclick: () => deployFlow(p) }, icon('rocket'), 'Deploy'));
    if (p.params && p.params.length) actions.push(h('button', { class: 'btn', type: 'button', onclick: () => paramsDialog(p) }, icon('sliders'), 'With parameters'));
    actions.push(h('button', { class: 'btn', type: 'button', onclick: () => deployFlow(p, 'restart') }, icon('refresh'), 'Restart'),
      h('button', { class: 'btn', type: 'button', onclick: () => deployFlow(p, 'pull') }, icon('download'), 'Pull'),
      h('button', { class: 'btn', type: 'button', onclick: () => deployFlow(p, 'stop') }, icon('pause'), 'Stop'),
      h('button', { class: 'btn danger', type: 'button', onclick: () => deployFlow(p, 'down') }, icon('stop'), 'Down'));
  }
  if (can('admin')) actions.push(h('a', { class: 'btn', href: '#/projects/' + p.id + '/edit' }, icon('edit'), 'Configure'));
  main.append(pageHeader([{ label: 'projects', href: '#/projects' }, p.folder ? { label: p.folder } : null, { label: p.name }].filter(Boolean), p.name, {
    badge: [sv ? h('span', { class: 'tag' }, sv.name) : null, p.disabled ? h('span', { class: 'badge mute' }, 'disabled') : h('span', { class: 'badge ok' }, 'enabled')],
    meta: h('div', {}, h('div', { class: 'meta' },
      h('span', {}, icon(p.source.kind === 'git' ? 'git' : 'folder'), h('b', {}, src)),
      p.cron ? h('span', {}, icon('clock'), p.cron, p.pollScm ? ' (poll git)' : '') : null,
      p.lastCommit ? h('span', {}, 'deployed ', h('b', {}, short(p.lastCommit))) : null),
    p.desc ? h('div', { class: 'desc' }, p.desc) : null),
    actions: actions }));
  if (p.disabled) main.append(h('div', { class: 'callout' }, icon('alert'), h('div', {}, 'This project is disabled: webhooks and schedules are ignored. Admins can still deploy manually after enabling it in Configure.')));
  const tabs = [['overview', 'Overview', 'grid'], ['stages', 'Stage view', 'columns'], ['history', 'Build history', 'history', builds.filter((b) => b.action !== 'poll').length]];
  if (can('admin')) tabs.push(['webhook', 'Webhook & API', 'webhook']);
  tab = tab || 'overview';
  main.append(h('div', { class: 'tabs' }, tabs.map(([k, l, ic, n]) => h('a', { href: '#/projects/' + p.id + (k === 'overview' ? '' : '/' + k), class: tab === k ? 'on' : '' }, icon(ic), l, n !== undefined ? h('span', { class: 'count' }, n) : null))));
  const real = builds.filter((b) => b.action !== 'poll');
  if (tab === 'overview') main.append(projectOverview(p, real, lastOk, lastBad, finished));
  else if (tab === 'stages') main.append(stageMatrix(real));
  else if (tab === 'history') main.append(historyTable(real, p));
  else if (tab === 'webhook') main.append(webhookPanel(p));
  every(5000, async () => {
    try {
      const nb = await api('GET', '/api/builds?project=' + id + '&limit=200');
      const sig = (arr) => arr.map((b) => b.id + b.status + (b.stages || []).length).join();
      if (sig(nb) !== sig(builds)) route();
    } catch (e) { /* ignore */ }
  });
}

function projectOverview(p, builds, lastOk, lastBad, finished) {
  const recent = finished.slice(0, 12);
  const pass = recent.length ? Math.round(recent.filter((b) => b.status === 'success').length / recent.length * 100) : null;
  const durs = finished.filter((b) => b.status === 'success').slice(0, 20).map(buildDur).filter((x) => x !== null);
  const avg = durs.length ? durs.reduce((a, b) => a + b, 0) / durs.length : null;
  const card = (title, b, extra) => h('div', { class: 'stat' }, h('div', { class: 'k' }, title),
    b ? [h('div', { class: 'kpi-line' }, h('a', { class: 'm ' + stc(b.status), href: '#/builds/' + b.id, style: { fontWeight: 700, fontSize: '16px' } }, '#' + b.number),
      b.commit ? h('span', { class: 'm mute' }, 'commit ' + short(b.commit)) : null), h('div', { class: 'faint', style: { fontSize: '11.5px', marginTop: '3px' } }, extra(b))]
      : h('div', { class: 'faint', style: { fontSize: '12px', marginTop: '6px' } }, 'None yet'));
  const params = (p.params || []);
  return h('div', { class: 'stack' },
    h('div', { class: 'grid g4' },
      card('Last successful build', lastOk, (b) => 'Completed in ' + dur(buildDur(b)) + ' · ' + ago(b.finished)),
      card('Last failed build', lastBad, (b) => (b.reason ? b.reason.slice(0, 80) : 'failed') + ' · ' + ago(b.finished)),
      h('div', { class: 'stat' }, h('div', { class: 'k' }, 'Stack health'), h('div', { class: 'v' }, h('b', { class: pass === null ? 'faint' : pass >= 90 ? 'okc' : pass >= 60 ? 'warnc' : 'badc' }, pass === null ? '–' : pass + '%'),
        h('span', {}, recent.length ? '(' + recent.filter((b) => b.status === 'success').length + '/' + recent.length + ' passed)' : 'no builds'))),
      h('div', { class: 'stat' }, h('div', { class: 'k' }, 'Avg duration'), h('div', { class: 'v' }, h('b', {}, avg === null ? '–' : dur(avg)),
        h('span', {}, durs.length > 1 && avg ? (Math.abs(durs[0] - avg) < 1000 ? 'last run on par' : durs[0] < avg ? '↓ ' + dur(avg - durs[0]) + ' faster last run' : '↑ ' + dur(durs[0] - avg) + ' slower last run') : '')))),
    h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('div', {}, h('h2', {}, icon('activity'), 'Build duration & stability'), h('div', { class: 'sub' }, 'Last ' + Math.min(30, finished.length) + ' finished builds')),
      h('div', { class: 'legend' }, h('span', {}, h('i', { style: { background: 'var(--ok-bar)' } }), 'Success'), h('span', {}, h('i', { style: { background: 'var(--bad-bar)' } }), 'Failure'))),
    h('div', { class: 'card-b' }, finished.length ? durationChart(finished.slice(0, 30).reverse()) : h('div', { class: 'faint' }, 'Deploy once to see the trend.'))),
    h('div', { class: 'cols' },
      h('div', {}, h('div', { class: 'section-h' }, h('h2', {}, icon('history'), 'Recent builds'), h('a', { href: '#/projects/' + p.id + '/history' }, 'Full history →')), historyTable(builds.slice(0, 6), p)),
      h('div', { class: 'stack' },
        h('div', { class: 'card pad' }, h('h2', {}, icon('box'), 'Configuration'), h('dl', { class: 'agent', style: { border: 0, padding: '10px 0 0', background: 'none' } },
          h('dl', {}, h('dt', {}, 'Source'), h('dd', {}, p.source.kind), h('dt', {}, 'Compose file'), h('dd', {}, p.source.composeFile || 'auto'),
            h('dt', {}, 'Env'), h('dd', {}, [p.hasEnv ? 'encrypted inline' : null, p.envFile ? 'file' : null].filter(Boolean).join(' + ') || 'none'),
            h('dt', {}, 'Profiles'), h('dd', {}, p.profiles || '–'), h('dt', {}, 'Health check'), h('dd', {}, p.healthUrl ? 'yes (' + (p.healthSecs || 60) + 's)' : 'off'),
            h('dt', {}, 'Retention'), h('dd', {}, (p.keep || 50) + ' builds'), h('dt', {}, 'Notify'), h('dd', {}, (p.notifyUrls || []).length ? p.notifyOn : 'off')))),
        params.length ? h('div', { class: 'card pad' }, h('h2', {}, icon('sliders'), 'Build parameters'),
          h('div', { class: 'feed mt8' }, params.map((d) => h('div', { class: 'feed-item' }, h('div', { class: 't' }, h('span', {}, d.name), d.secret ? h('span', { class: 'badge mute' }, 'secret') : null),
            h('div', { class: 's' }, d.choices ? 'one of ' + d.choices : 'default ' + (d.secret ? '••••' : (d.default || '""'))))))) : null)));
}

function durationChart(bs) {
  const W = 760, H = 150, padL = 34, padB = 18, padT = 8;
  const vals = bs.map((b) => (buildDur(b) || 0) / 1000);
  const max = Math.max(5, ...vals) * 1.15;
  const avg = vals.reduce((a, b) => a + b, 0) / (vals.length || 1);
  const bw = (W - padL - 6) / bs.length;
  const s = document.createElementNS(SVGNS, 'svg');
  s.setAttribute('viewBox', `0 0 ${W} ${H}`); s.setAttribute('class', 'chart'); s.setAttribute('preserveAspectRatio', 'none');
  const add = (tag, attrs, text) => { const e = document.createElementNS(SVGNS, tag); for (const k in attrs) e.setAttribute(k, attrs[k]); if (text !== undefined) e.textContent = text; s.append(e); return e; };
  const y = (v) => padT + (H - padT - padB) * (1 - v / max);
  for (let i = 0; i <= 3; i++) { const v = max / 3 * i; add('line', { x1: padL, x2: W, y1: y(v), y2: y(v), class: 'grid-l' }); add('text', { x: 0, y: y(v) + 3 }, Math.round(v) + 's'); }
  bs.forEach((b, i) => {
    const v = vals[i];
    const r = add('rect', { x: padL + i * bw + bw * 0.18, y: y(v), width: Math.max(2, bw * 0.64), height: Math.max(1, y(0) - y(v)), rx: 2, class: b.status });
    const t = document.createElementNS(SVGNS, 'title'); t.textContent = '#' + b.number + ' · ' + b.status + ' · ' + dur(v * 1000) + ' · ' + fmtDate(b.finished); r.append(t);
    r.style.cursor = 'pointer'; r.addEventListener('click', () => go('#/builds/' + b.id));
  });
  add('line', { x1: padL, x2: W, y1: y(avg), y2: y(avg), class: 'avg' });
  add('text', { x: W - 90, y: y(avg) - 4 }, 'avg ' + dur(avg * 1000));
  if (bs.length) { add('text', { x: padL, y: H - 3 }, '#' + bs[0].number); add('text', { x: W - 30, y: H - 3 }, '#' + bs[bs.length - 1].number); }
  return s;
}

function stageMatrix(builds) {
  const withStages = builds.filter((b) => b.stages && b.stages.length).slice(0, 25);
  if (!withStages.length) return h('div', { class: 'tbl-wrap' }, emptyState('columns', 'No stage data yet', 'Stages appear for builds run by Dockup 0.2 or later.'));
  const names = STAGE_ORDER.filter((n) => withStages.some((b) => b.stages.some((s) => s.name === n)));
  return h('div', { class: 'tbl-wrap' }, h('table', { class: 'matrix' },
    h('thead', {}, h('tr', {}, h('th', {}, 'Build'), names.map((n, i) => h('th', { class: 'c' }, (i + 1) + '. ' + (STAGE_LABEL[n] || n))), h('th', { class: 'num' }, 'Total'))),
    h('tbody', {}, withStages.map((b) => h('tr', {},
      h('td', {}, h('a', { class: 'm ' + stc(b.status), href: '#/builds/' + b.id, style: { fontWeight: 600 } }, '#' + b.number), h('div', { class: 'faint m', style: { fontSize: '10px' } }, ago(b.queued))),
      names.map((n) => {
        const s = b.stages.find((x) => x.name === n);
        if (!s) return h('td', { class: 'st none' }, h('div', {}, '—'));
        return h('td', { class: 'st ' + s.status, title: s.status }, h('div', {}, s.status === 'failed' ? 'FAILED ' : s.status === 'cancelled' ? 'CANCELLED ' : '', dur(stageDur(s))));
      }),
      h('td', { class: 'num m ' + stc(b.status) }, buildDur(b) !== null ? dur(buildDur(b)) : '–'))))));
}

function historyTable(builds, p) {
  if (!builds.length) return h('div', { class: 'tbl-wrap' }, emptyState('history', 'No builds yet', 'Deploy this project to start its history.'));
  const isGit = p && p.source && p.source.kind === 'git';
  const live = p && p.lastCommit ? builds.find((b) => b.status === 'success' && b.commit === p.lastCommit && (b.action === 'deploy' || b.action === 'rollback')) : null;
  return h('div', { class: 'tbl-wrap' }, h('table', {},
    h('thead', {}, h('tr', {}, h('th', {}, 'Build'), h('th', {}, 'Status'), h('th', {}, 'Action'), h('th', {}, 'Commit'), h('th', {}, 'Triggered by'), h('th', {}, 'Server'), h('th', {}, 'Duration'), h('th', {}, 'Started'), h('th', { class: 'num' }, ''))),
    h('tbody', {}, builds.map((b) => {
      const acts = [h('a', { class: 'btn xs', href: '#/builds/' + b.id }, icon('terminal'), 'Console')];
      if (isGit && can('operator') && b.status === 'success' && b.commit && !p.disabled && b.commit !== p.lastCommit) {
        acts.push(h('button', { class: 'btn xs', type: 'button', title: 'Redeploy commit ' + short(b.commit), onclick: async () => {
          if (!(await confirmDialog('Roll back ' + p.name + '?', 'Deploys commit ' + short(b.commit) + ' (build #' + b.number + ') again. The branch is not changed; the next normal deploy returns to its head.', { ok: 'Roll back' }))) return;
          try { const nb = await trigger(p.id, 'rollback', null, b.commit); go('#/builds/' + nb.id); } catch (e) { fail(e); }
        } }, icon('rollback'), 'Rollback here'));
      }
      return h('tr', {},
        h('td', {}, h('a', { class: 'm ' + stc(b.status), href: '#/builds/' + b.id, style: { fontWeight: 600 } }, '#' + b.number)),
        h('td', {}, badge(b.status)), h('td', { class: 'm' }, b.action),
        h('td', { class: 'm' }, b.commit ? short(b.commit) : h('span', { class: 'faint' }, '–'), live && b.id === live.id ? h('span', { class: 'badge ok', style: { marginLeft: '6px' }, title: 'Currently deployed commit' }, 'live') : null),
        h('td', { class: 'nowrap' }, h('span', { class: 'm' }, TRIGGER[b.trigger] || b.trigger), b.by ? h('span', { class: 'faint' }, ' · ' + b.by) : null),
        h('td', {}, h('span', { class: 'tag' }, b.server || '?')),
        h('td', { class: 'm nowrap' }, buildDur(b) === null ? '–' : dur(buildDur(b))),
        h('td', { class: 'm mute nowrap', title: fmtDate(b.queued) }, ago(b.queued)),
        h('td', { class: 'num' }, h('div', { class: 'row-actions', style: { justifyContent: 'flex-end' } }, acts)));
    }))));
}

function webhookPanel(p) {
  const tokUrl = p.webhookUrl + '?token=' + p.webhook;
  const curl = `curl -X POST -H "Authorization: Bearer dk_YOUR_TOKEN" -H "Content-Type: application/json" \\\n  -d '{"action":"deploy"}' ${location.origin}/api/projects/${p.id}/builds`;
  return h('div', { class: 'grid g2', style: { alignItems: 'start' } },
    h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('div', {}, h('h2', {}, icon('webhook'), 'Git webhook'), h('div', { class: 'sub' }, 'Pushes to ' + (p.source.branch || 'the branch') + ' deploy automatically.'))),
      h('div', { class: 'card-b stack' },
        h('label', { class: 'f' }, 'Payload URL', h('div', { class: 'code-box' }, h('code', {}, p.webhookUrl), copyBtn(p.webhookUrl, 'Webhook URL'))),
        h('label', { class: 'f' }, 'Secret', h('div', { class: 'code-box' }, h('code', {}, '•'.repeat(16)), copyBtn(p.webhook, 'Webhook secret'))),
        h('div', { class: 'hint' }, h('b', {}, 'GitHub / Gitea: '), 'content type application/json, secret = the value above (HMAC-SHA256 is verified). ',
          h('b', {}, 'GitLab: '), 'put the secret in “Secret token”.'),
        h('details', {}, h('summary', { class: 'hint', style: { cursor: 'pointer' } }, 'Other tools (secret in the URL, less secure)'),
          h('div', { class: 'code-box mt8' }, h('code', {}, p.webhookUrl + '?token=••••'), copyBtn(tokUrl, 'URL'))),
        h('div', {}, h('button', { class: 'btn sm danger', type: 'button', onclick: async () => {
          if (!(await confirmDialog('Regenerate webhook secret?', 'Existing webhooks stop working until you update them with the new secret.', { danger: true, ok: 'Regenerate' }))) return;
          try { const full = await api('GET', '/api/projects/' + p.id); await api('PUT', '/api/projects/' + p.id, Object.assign(projectInput(full), { regenWebhook: true })); toast('New webhook secret generated'); route(); } catch (e) { fail(e); }
        } }, icon('refresh'), 'Regenerate secret')))),
    h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('div', {}, h('h2', {}, icon('terminal'), 'API trigger'), h('div', { class: 'sub' }, 'Use a personal API token from Account → API tokens.'))),
      h('div', { class: 'card-b stack' }, h('pre', { class: 'code' }, curl), h('div', {}, copyBtn(curl, 'Command')),
        h('div', { class: 'hint' }, 'Actions: deploy, restart, pull, stop, down, rollback (with "commit"). Parameters: "params": {"NAME": "value"}.'))));
}

// =====================================================================
//  builds list & build console
// =====================================================================
async function pageBuilds() {
  const main = shell('builds');
  let status = '', project = '';
  const slot = h('div');
  const pSel = h('select', { class: 'inp', onchange: () => { project = pSel.value; load(); } }, h('option', { value: '' }, 'All projects'));
  const sSel = h('select', { class: 'inp', onchange: () => { status = sSel.value; draw(); } }, ['', 'running', 'queued', 'success', 'failed', 'cancelled'].map((s) => h('option', { value: s }, s || 'Any status')));
  main.append(pageHeader([{ label: 'builds' }], 'Builds & Logs', { sub: 'Every deployment, restart and rollback with its console output.' }), h('div', { class: 'filters' }, pSel, sSel), slot);
  await refreshSummary(); refreshNav();
  (summary ? summary.projects : []).slice().sort((a, b) => a.name.localeCompare(b.name)).forEach((p) => pSel.append(h('option', { value: p.id }, p.name)));
  let builds = [];
  const draw = () => {
    const list = builds.filter((b) => !status || b.status === status);
    clear(slot).append(list.length ? h('div', { class: 'tbl-wrap' }, h('table', {},
      h('thead', {}, h('tr', {}, h('th', {}, ''), h('th', {}, 'Project'), h('th', {}, 'Build'), h('th', {}, 'Action'), h('th', {}, 'Status'), h('th', {}, 'Trigger'), h('th', {}, 'Server'), h('th', {}, 'Commit'), h('th', {}, 'Duration'), h('th', {}, 'Queued'))),
      h('tbody', {}, list.map((b) => h('tr', { style: { cursor: 'pointer' }, onclick: (e) => { if (e.target.tagName !== 'A') go('#/builds/' + b.id); } },
        h('td', { class: 'c' }, dot(b.status)), h('td', {}, h('a', { class: 'cell-title', href: '#/projects/' + b.projectId }, b.project || '?')),
        h('td', {}, h('a', { class: 'm ' + stc(b.status), href: '#/builds/' + b.id, style: { fontWeight: 600 } }, '#' + b.number)),
        h('td', { class: 'm' }, b.action), h('td', {}, badge(b.status)),
        h('td', {}, h('span', { class: 'm' }, TRIGGER[b.trigger] || b.trigger), b.by ? h('span', { class: 'faint' }, ' · ' + b.by) : null),
        h('td', {}, h('span', { class: 'tag' }, b.server || '?')), h('td', { class: 'm' }, short(b.commit) || h('span', { class: 'faint' }, '–')),
        h('td', { class: 'm' }, buildDur(b) === null ? '–' : dur(buildDur(b))), h('td', { class: 'm mute', title: fmtDate(b.queued) }, ago(b.queued)))))))
      : h('div', { class: 'tbl-wrap' }, emptyState('terminal', 'No builds', status || project ? 'Nothing matches the filters.' : 'Deploy a project to see builds here.')));
  };
  const load = async () => { try { builds = await api('GET', '/api/builds?limit=300' + (project ? '&project=' + project : '')); draw(); } catch (e) { fail(e); } };
  await load();
  every(4000, load);
}

function parseLine(raw) {
  const tab = raw.indexOf('\t');
  if (tab === 29 || (tab > 18 && tab < 40 && /^\d{4}-\d\d-\d\dT/.test(raw))) return { ts: raw.slice(0, tab), tx: raw.slice(tab + 1) };
  return { ts: '', tx: raw };
}
function lineClass(tx) {
  if (tx.startsWith('$ ')) return 'cmd';
  if (tx.startsWith('==> SUCCESS') || /^health OK/.test(tx)) return 'ok';
  if (tx.startsWith('==> FAILED') || /\b(error|fatal|failed|denied|panic)\b/i.test(tx)) return 'err';
  if (/\b(warn|warning|deprecated)\b/i.test(tx)) return 'warn';
  if (tx.startsWith('==> ') || tx.startsWith('Started by ')) return 'hdr';
  return '';
}
const fmtTs = (ts) => { if (!ts) return ''; const d = new Date(ts); return isNaN(d) ? '' : d.toLocaleTimeString(undefined, { hour12: false }) + '.' + String(d.getMilliseconds()).padStart(3, '0'); };

async function pageBuild(id) {
  const main = shell('builds');
  let b;
  try { b = await api('GET', '/api/builds/' + id); } catch (e) { main.append(emptyState('alert', 'Build not found', e.message)); return; }
  let project = null;
  try { project = await api('GET', '/api/projects/' + b.projectId); } catch (e) { /* project deleted */ }
  const head = h('div'), stagesEl = h('div'), reasonEl = h('div');
  const prefs = (() => { try { return JSON.parse(localStorage.getItem('dockup-console') || '{}'); } catch (e) { return {}; } })();
  let follow = prefs.follow !== false, showTs = !!prefs.ts, wrap = !!prefs.wrap;
  const savePrefs = () => { try { localStorage.setItem('dockup-console', JSON.stringify({ follow, ts: showTs, wrap })); } catch (e) { /* ignore */ } };
  const con = h('div', { class: 'console' + (showTs ? ' show-ts' : '') + (wrap ? ' wrap' : ''), role: 'log', 'aria-live': 'off' });
  const foot = h('div', { class: 'console-foot' });
  let lineNo = 0, root = con, group = null, buf = '';
  const groups = {};
  const filterInp = h('input', { class: 'inp', placeholder: 'Filter lines…', autocomplete: 'off' });
  const tgl = (label, on, fn) => { const bt = h('button', { type: 'button', class: 'tgl' + (on ? ' on' : '') }, label); bt.addEventListener('click', () => { const v = fn(); bt.classList.toggle('on', v); savePrefs(); }); return bt; };
  const followBtn = tgl('Follow tail', follow, () => { follow = !follow; if (follow) con.scrollTop = con.scrollHeight; return follow; });
  const bar = h('div', { class: 'console-bar' },
    h('div', { class: 'flex' }, icon('search'), filterInp,
      h('button', { type: 'button', class: 'tgl', onclick: () => { const e = con.querySelector('.ln.err'); if (e) { openParents(e); e.scrollIntoView({ block: 'center' }); follow = false; followBtn.classList.remove('on'); } else toast('No error lines', 'info'); } }, icon('arrowdown'), 'Jump to error')),
    h('div', { class: 'flex' }, followBtn,
      tgl('Timestamps', showTs, () => { showTs = !showTs; con.classList.toggle('show-ts', showTs); return showTs; }),
      tgl('Wrap', wrap, () => { wrap = !wrap; con.classList.toggle('wrap', wrap); return wrap; }),
      h('button', { type: 'button', class: 'tgl', onclick: () => con.querySelectorAll('.grp').forEach((g) => g.classList.add('closed')) }, 'Collapse all'),
      h('button', { type: 'button', class: 'tgl', onclick: () => con.querySelectorAll('.grp').forEach((g) => g.classList.remove('closed')) }, 'Expand all')));
  const openParents = (el) => { let g = el.closest('.grp'); while (g) { g.classList.remove('closed'); g = g.parentElement.closest('.grp'); } };
  con.addEventListener('scroll', () => {
    const atEnd = con.scrollHeight - con.scrollTop - con.clientHeight < 30;
    if (!atEnd && follow && con.dataset.auto !== '1') { follow = false; followBtn.classList.remove('on'); }
    con.dataset.auto = '';
  });
  let fq = '';
  const matches = (el) => !fq || el.dataset.tx.toLowerCase().includes(fq);
  filterInp.addEventListener('input', () => {
    fq = filterInp.value.trim().toLowerCase();
    con.querySelectorAll('.ln').forEach((el) => { const m = matches(el); el.classList.toggle('hidden', !m); el.classList.toggle('hit', !!fq && m); if (fq && m) openParents(el); });
  });
  const addLine = (raw) => {
    const { ts, tx } = parseLine(raw);
    if (tx.startsWith('##[stage] ')) {
      const [name, st] = tx.slice(10).split(' ');
      if (st === 'running') {
        const hdrSt = h('span', { class: 'st' });
        const g = h('div', { class: 'grp' }, h('div', { class: 'grp-h', onclick: () => g.classList.toggle('closed') }, icon('down', 'chev'), (STAGE_LABEL[name] || name), hdrSt), h('div', { class: 'grp-b' }));
        g._st = hdrSt; g._start = ts; con.append(g); groups[name] = g; group = g; root = g.lastChild;
        clear(hdrSt).append(icon('loader', 'run'), 'running');
      } else if (groups[name]) {
        const g = groups[name];
        const d = g._start && ts ? new Date(ts) - new Date(g._start) : null;
        clear(g._st).append(icon(st === 'success' ? 'check' : 'x', st === 'success' ? 'ok' : 'bad'), st + (d !== null ? ' · ' + dur(d) : ''));
        if (st === 'success' && b.status !== 'running' && Object.keys(groups).length > 2) g.classList.add('closed');
        if (g === group) { group = null; root = con; }
      }
      return;
    }
    lineNo++;
    const el = h('div', { class: 'ln ' + lineClass(tx), dataset: { tx } }, h('span', { class: 'no' }, lineNo), h('span', { class: 'ts' }, fmtTs(ts)), h('span', { class: 'tx' }, tx || ' '));
    if (fq) { const m = matches(el); el.classList.toggle('hidden', !m); el.classList.toggle('hit', m); }
    root.append(el);
  };
  const ingest = (chunk) => {
    buf += chunk;
    const parts = buf.split('\n'); buf = parts.pop();
    parts.forEach(addLine);
    foot.firstChild && (foot.firstChild.textContent = lineNo + ' lines');
    if (follow) { con.dataset.auto = '1'; con.scrollTop = con.scrollHeight; }
  };
  let elapsedTimer = null;
  const drawHead = () => {
    const live = b.status === 'running' || b.status === 'queued';
    const acts = [];
    if (live && can('operator')) acts.push(h('button', { class: 'btn danger', type: 'button', onclick: async () => { try { await api('POST', '/api/builds/' + b.id + '/cancel'); toast('Cancel requested'); } catch (e) { fail(e); } } }, icon('stop'), 'Cancel'));
    acts.push(h('a', { class: 'btn', href: '/api/builds/' + b.id + '/log?download=1' }, icon('download'), 'Download log'));
    if (!live && can('operator') && project && !project.disabled && b.action !== 'poll') {
      acts.push(h('button', { class: 'btn', type: 'button', onclick: async () => {
        try {
          const action = b.action === 'rollback' ? 'rollback' : b.action;
          if (project.params && project.params.length && action === 'deploy') return paramsDialog(project, b.params);
          const nb = await trigger(project.id, action, null, b.action === 'rollback' ? b.commit : ''); go('#/builds/' + nb.id);
        } catch (e) { fail(e); }
      } }, icon('refresh'), 'Re-run'));
    }
    const el = h('span', { class: 'm' }, buildDur(b) === null ? '–' : dur(buildDur(b)));
    clearInterval(elapsedTimer);
    if (b.status === 'running') { elapsedTimer = setInterval(() => { el.textContent = dur(buildDur(b)); }, 1000); onCleanup(() => clearInterval(elapsedTimer)); }
    clear(head).append(pageHeader([{ label: 'projects', href: '#/projects' }, { label: b.project || 'deleted project', href: project ? '#/projects/' + b.projectId : null }, { label: 'build #' + b.number }],
      '#' + b.number + ' ' + (b.action === 'deploy' ? '' : b.action), {
        badge: badge(b.status),
        meta: h('div', { class: 'meta' },
          h('span', {}, 'Trigger: ', h('b', {}, TRIGGER[b.trigger] || b.trigger)), b.by ? h('span', {}, 'By: ', h('b', {}, b.by)) : null,
          h('span', {}, icon('server'), h('b', {}, b.server || '?')), b.commit ? h('span', {}, icon('git'), h('b', {}, short(b.commit))) : null,
          h('span', {}, icon('clock'), el), h('span', { title: fmtDate(b.queued) }, 'queued ', ago(b.queued)),
          b.params && Object.keys(b.params).length ? h('span', {}, icon('sliders'), Object.entries(b.params).map(([k, v]) => k + '=' + v).join(' ')) : null),
        actions: acts }));
    clear(reasonEl);
    if (b.status === 'failed' && b.reason) reasonEl.append(h('div', { class: 'callout bad' }, icon('alert'), h('div', {}, h('b', {}, 'Failed: '), b.reason)));
    if (b.status === 'queued') reasonEl.append(h('div', { class: 'callout info' }, icon('clock'), h('div', {}, 'Waiting: another build of this project is running, or the server is offline (queued builds fail after 5 minutes offline).')));
    const st = b.stages || [];
    clear(stagesEl);
    if (st.length) {
      const doneN = st.filter((s) => s.status === 'success').length;
      stagesEl.append(h('div', { class: 'section-h' }, h('h2', {}, icon('columns'), 'Pipeline'), h('span', { class: 'faint m', style: { fontSize: '11px' } }, doneN + ' of ' + st.length + ' stages done')),
        h('div', { class: 'stages' }, st.map((s, i) => h('div', { class: 'stage ' + s.status, style: { cursor: 'pointer' }, onclick: () => { const g = groups[s.name]; if (g) { g.classList.remove('closed'); g.scrollIntoView({ block: 'start' }); } } },
          h('div', { class: 'n' }, h('span', { class: 'ellip' }, (i + 1) + '. ' + (STAGE_LABEL[s.name] || s.name)), icon(s.status === 'success' ? 'check' : s.status === 'running' ? 'loader' : 'x')),
          h('div', { class: 'd' }, s.status === 'running' ? 'running (' + dur(stageDur(s)) + ')' : dur(stageDur(s)))))));
    }
  };
  drawHead();
  foot.append(h('span', {}, '0 lines'), h('span', {}, 'Agent: ' + (b.server || '?')), h('span', { class: 'conn' }, ''));
  main.append(head, reasonEl, stagesEl, h('div', { class: 'console-wrap' }, bar, con, foot));
  // stream (falls back to the static log for finished builds)
  const connEl = foot.querySelector('.conn');
  if (b.status === 'running' || b.status === 'queued') {
    const es = new EventSource('/api/builds/' + id + '/stream');
    onCleanup(() => es.close());
    es.onopen = () => { connEl.textContent = '● live'; connEl.style.color = '#3FB950'; };
    es.addEventListener('log', (e) => ingest(JSON.parse(e.data)));
    es.addEventListener('build', (e) => { const nb = JSON.parse(e.data); const changed = nb.status !== b.status || JSON.stringify(nb.stages) !== JSON.stringify(b.stages); b = nb; if (changed) drawHead(); });
    es.addEventListener('end', () => { es.close(); connEl.textContent = 'finished'; connEl.style.color = ''; if (buf) { addLine(buf); buf = ''; }
      api('GET', '/api/builds/' + id).then((nb) => { b = nb; drawHead(); refreshSummary().then(refreshNav); }).catch(() => {}); });
    es.onerror = () => { connEl.textContent = 'reconnecting…'; connEl.style.color = '#D29922'; };
  } else {
    try {
      const txt = await fetch('/api/builds/' + id + '/log', { credentials: 'same-origin' }).then((r) => r.text());
      ingest(txt + (txt.endsWith('\n') ? '' : '\n'));
      connEl.textContent = 'archived';
      if (b.status === 'failed') { const e = con.querySelector('.ln.err:last-of-type') || [...con.querySelectorAll('.ln.err')].pop(); if (e) { openParents(e); } }
      if (follow) { con.scrollTop = con.scrollHeight; }
    } catch (e) { fail(e); }
  }
  if (!lineNo && b.status === 'queued') con.append(h('div', { class: 'ln' }, h('span', { class: 'no' }, ''), h('span', { class: 'tx', style: { color: '#8B949E' } }, 'Waiting for an executor…')));
}

// =====================================================================
//  configure project
// =====================================================================
function projectInput(p) {
  return { name: p.name, desc: p.desc || '', folder: p.folder || '', serverId: p.serverId,
    source: Object.assign({ kind: 'path', path: '', repo: '', branch: 'main', credentialId: '', composeFile: '' }, p.source || {}),
    envFile: p.envFile || '', params: (p.params || []).map((x) => Object.assign({}, x)), profiles: p.profiles || '', preCmd: p.preCmd || '', postCmd: p.postCmd || '',
    healthUrl: p.healthUrl || '', healthSecs: p.healthSecs || 0, cron: p.cron || '', pollScm: !!p.pollScm, notifyUrls: (p.notifyUrls || []).slice(),
    notifyOn: p.notifyOn || 'always', keep: p.keep || 0, noCache: !!p.noCache, pull: !!p.pull, disabled: !!p.disabled };
}
function parseEnv(text) {
  const out = [];
  for (const l of (text || '').split('\n')) {
    const t = l.trim(); if (!t) continue;
    if (t.startsWith('#')) { out.push({ k: '', v: t, comment: true }); continue; }
    const i = t.indexOf('=');
    out.push(i < 0 ? { k: t, v: '' } : { k: t.slice(0, i).trim(), v: t.slice(i + 1) });
  }
  return out;
}
const envText = (rows) => rows.filter((r) => r.comment ? r.v : r.k.trim()).map((r) => r.comment ? r.v : r.k.trim() + '=' + r.v).join('\n');

async function pageEdit(id) {
  const main = shell('projects');
  if (!can('admin')) { main.append(emptyState('lock', 'Admins only', 'Only admins can change projects, because hooks and compose files run with full privileges.')); return; }
  const isNew = !id;
  let p = null, servers = [], creds = [];
  try {
    [servers, creds] = await Promise.all([api('GET', '/api/servers'), api('GET', '/api/credentials')]);
    if (!isNew) p = await api('GET', '/api/projects/' + id);
  } catch (e) { main.append(emptyState('alert', 'Could not load', e.message)); return; }
  await refreshSummary(); refreshNav();
  const folders = [...new Set((summary ? summary.projects : []).map((x) => x.folder).filter(Boolean))].sort();
  const f = projectInput(p || { serverId: (servers.find((s) => s.local) || servers[0] || {}).id, source: { kind: 'git', branch: 'main' } });
  let envRows = null; // null = not loaded (keep stored value)
  let envRaw = false;
  const initial = JSON.stringify(f);
  let dirty = false;
  const isDirty = () => JSON.stringify(f) !== initial || envRows !== null;
  const bindings = [];
  const markDirty = () => { dirty = isDirty(); savebar.classList.toggle('hidden', !dirty && !isNew); };
  const inp = (path, opts) => {
    const { tag, mono, after, ...attrs } = opts || {};
    const get = () => path.split('.').reduce((o, k) => o[k], f);
    const set = (v) => { const ks = path.split('.'); const last = ks.pop(); ks.reduce((o, k) => o[k], f)[last] = v; };
    const num = attrs.type === 'number';
    const el = h(tag === 'textarea' ? 'textarea' : 'input', Object.assign({ class: 'inp' + (mono ? ' mono' : ''), spellcheck: mono ? 'false' : null }, attrs));
    el.value = num && !get() ? '' : get();
    el.addEventListener('input', () => { set(num ? (parseInt(el.value, 10) || 0) : el.value); markDirty(); if (after) after(); });
    bindings.push(el);
    return el;
  };
  const chk = (path, title, text) => {
    const el = h('input', { type: 'checkbox', checked: path.split('.').reduce((o, k) => o[k], f) });
    el.addEventListener('change', () => { const ks = path.split('.'); const last = ks.pop(); ks.reduce((o, k) => o[k], f)[last] = el.checked; markDirty(); });
    return h('label', { class: 'check' }, el, h('div', {}, h('b', {}, title), text ? h('span', { class: 'hint' }, text) : null));
  };
  const section = (sid, title, sub, ...content) => h('section', { id: 'sec-' + sid, class: 'card' }, h('div', { class: 'card-h' }, h('div', {}, h('h2', {}, title), h('div', { class: 'sub' }, sub))), h('div', { class: 'card-b stack' }, content));

  // --- general
  const serverSel = h('select', { class: 'inp', onchange: () => { f.serverId = serverSel.value; markDirty(); } },
    servers.map((s) => h('option', { value: s.id, selected: s.id === f.serverId }, s.name + (s.hostname ? ' (' + s.hostname + ')' : '') + (s.online ? ' · online' : ' · offline') + (s.docker ? ', Docker ' + s.docker : ''))));
  const folderInp = inp('folder', { list: 'dockup-folders', placeholder: 'e.g. Core Services (optional)' });
  const general = section('general', 'General', 'Project identity and the server that runs it.',
    h('div', { class: 'fgrid' }, h('label', { class: 'f' }, 'Project name', inp('name', { required: true, maxlength: '80', placeholder: 'payments-api' })),
      h('label', { class: 'f' }, 'Folder', folderInp, h('datalist', { id: 'dockup-folders' }, folders.map((x) => h('option', { value: x })))),
      h('label', { class: 'f full' }, h('span', {}, 'Description ', h('span', { class: 'hint' }, '(optional)')), inp('desc', { tag: 'textarea', maxlength: '500', rows: '2', style: { minHeight: '52px', fontFamily: 'var(--font)' } })),
      h('label', { class: 'f full' }, 'Target server (agent)', serverSel, h('span', { class: 'hint' }, 'Builds run on this server. Add servers on the Servers page.'))));

  // --- source
  const srcBody = h('div', { class: 'stack' });
  const credSel = () => h('select', { class: 'inp', onchange: (e) => { f.source.credentialId = e.target.value; markDirty(); } },
    h('option', { value: '' }, 'None (public repository)'), creds.map((c) => h('option', { value: c.id, selected: c.id === f.source.credentialId }, c.name + ' (' + (c.kind === 'ssh-key' ? 'SSH key' : 'access token') + ')')));
  const drawSrc = () => {
    clear(srcBody);
    srcBody.append(h('div', { class: 'seg' },
      h('button', { type: 'button', class: f.source.kind === 'git' ? 'on' : '', onclick: () => { f.source.kind = 'git'; markDirty(); drawSrc(); } }, icon('git'), 'Git repository'),
      h('button', { type: 'button', class: f.source.kind === 'path' ? 'on' : '', onclick: () => { f.source.kind = 'path'; markDirty(); drawSrc(); } }, icon('folder'), 'Compose path on server')));
    if (f.source.kind === 'git') {
      srcBody.append(h('div', { class: 'fgrid' },
        h('label', { class: 'f full' }, 'Repository URL', inp('source.repo', { mono: true, placeholder: 'https://github.com/org/app.git  or  git@github.com:org/app.git' }),
          h('span', { class: 'hint' }, 'Do not put tokens in the URL: add a credential instead.')),
        h('label', { class: 'f' }, 'Branch', inp('source.branch', { mono: true, placeholder: 'main' })),
        h('label', { class: 'f' }, h('span', {}, 'Compose file ', h('span', { class: 'hint' }, '(relative to repo root)')), inp('source.composeFile', { mono: true, placeholder: 'auto: compose.yaml, docker-compose.yml…' })),
        h('label', { class: 'f full' }, 'Git credential', credSel(), creds.length ? null : h('span', { class: 'hint' }, 'No credentials yet. ', h('a', { href: '#/credentials' }, 'Add one'), ' for private repositories.'))));
    } else {
      srcBody.append(h('label', { class: 'f' }, 'Compose directory or file on the server', inp('source.path', { mono: true, placeholder: '/opt/stacks/blog' }),
        h('span', { class: 'hint' }, 'Absolute path on the target server. With Docker, mount it at the same path inside the container.')));
    }
  };
  drawSrc();
  const source = section('source', 'Source & Compose', 'Where Dockup gets your compose file.', srcBody);

  // --- environment
  const envSlot = h('div');
  const drawEnv = () => {
    clear(envSlot);
    if (envRows === null) {
      const has = p && p.hasEnv;
      envSlot.append(h('div', { class: 'flex wrap' }, h('span', { class: 'hint' }, has ? 'Inline variables are stored encrypted (AES-256-GCM). Viewing them is recorded in the audit log.' : 'No inline variables yet.'),
        h('button', { type: 'button', class: 'btn sm', onclick: async () => {
          try { envRows = has ? parseEnv((await api('GET', '/api/projects/' + p.id + '/env')).env) : []; if (!envRows.length) envRows.push({ k: '', v: '' }); markDirty(); drawEnv(); } catch (e) { fail(e); }
        } }, icon(has ? 'eye' : 'plus'), has ? 'Decrypt & edit variables' : 'Add variables')));
      return;
    }
    const tools = h('div', { class: 'flex between' }, h('span', { class: 'hint' }, 'Passed to compose as an extra --env-file. Values of sensitive-looking keys are masked in logs.'),
      h('div', { class: 'seg' }, h('button', { type: 'button', class: envRaw ? '' : 'on', onclick: () => { envRaw = false; drawEnv(); } }, 'Table'),
        h('button', { type: 'button', class: envRaw ? 'on' : '', onclick: () => { envRaw = true; drawEnv(); } }, 'Raw .env')));
    envSlot.append(tools);
    if (envRaw) {
      const ta = h('textarea', { class: 'inp', rows: '10', spellcheck: 'false', placeholder: 'KEY=value' });
      ta.value = envText(envRows);
      ta.addEventListener('input', () => { envRows = parseEnv(ta.value); markDirty(); });
      envSlot.append(h('div', { class: 'mt8' }, ta));
      return;
    }
    const list = h('div', { class: 'stack mt8', style: { gap: '6px' } });
    envRows.forEach((r, i) => {
      if (r.comment) return;
      const k = h('input', { class: 'inp mono sm', value: r.k, placeholder: 'KEY', oninput: (e) => { r.k = e.target.value; markDirty(); } });
      const v = h('input', { class: 'inp mono sm', type: 'password', value: r.v, placeholder: 'value', autocomplete: 'off', oninput: (e) => { r.v = e.target.value; markDirty(); } });
      const eye = h('button', { type: 'button', class: 'icon-btn', title: 'Show / hide', onclick: () => { v.type = v.type === 'password' ? 'text' : 'password'; clear(eye).append(icon(v.type === 'password' ? 'eye' : 'eyeoff')); } }, icon('eye'));
      list.append(h('div', { class: 'kv' }, k, h('span', { class: 'eq' }, '='), v, eye,
        h('button', { type: 'button', class: 'icon-btn', title: 'Remove', onclick: () => { envRows.splice(i, 1); markDirty(); drawEnv(); } }, icon('x'))));
    });
    envSlot.append(list, h('div', { class: 'mt8' }, h('button', { type: 'button', class: 'btn sm', onclick: () => { envRows.push({ k: '', v: '' }); markDirty(); drawEnv(); } }, icon('plus'), 'Add variable')));
  };
  drawEnv();
  const paramSlot = h('div');
  const drawParams = () => {
    clear(paramSlot);
    if (!f.params.length) { paramSlot.append(h('div', { class: 'hint' }, 'No parameters. Parameters are asked for on “Deploy with parameters” and exported as environment variables.')); }
    else {
      paramSlot.append(h('div', { class: 'tbl-wrap' }, h('table', {}, h('thead', {}, h('tr', {}, h('th', {}, 'Name'), h('th', {}, 'Default'), h('th', {}, 'Choices (comma separated)'), h('th', { class: 'c' }, 'Secret'), h('th', {}, ''))),
        h('tbody', {}, f.params.map((d, i) => h('tr', {},
          h('td', {}, h('input', { class: 'inp mono sm', value: d.name, placeholder: 'RUN_MIGRATIONS', oninput: (e) => { d.name = e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g, '_'); e.target.value = d.name; markDirty(); } })),
          h('td', {}, h('input', { class: 'inp mono sm', type: d.secret ? 'password' : 'text', value: d.default || '', oninput: (e) => { d.default = e.target.value; markDirty(); } })),
          h('td', {}, h('input', { class: 'inp mono sm', value: d.choices || '', placeholder: 'free text', oninput: (e) => { d.choices = e.target.value; markDirty(); } })),
          h('td', { class: 'c' }, h('input', { type: 'checkbox', checked: !!d.secret, onchange: (e) => { d.secret = e.target.checked; markDirty(); drawParams(); } })),
          h('td', { class: 'num' }, h('button', { type: 'button', class: 'btn xs danger', onclick: () => { f.params.splice(i, 1); markDirty(); drawParams(); } }, 'Delete'))))))));
    }
    paramSlot.append(h('div', { class: 'mt8' }, h('button', { type: 'button', class: 'btn sm', onclick: () => { f.params.push({ name: '', default: '', choices: '', secret: false }); markDirty(); drawParams(); } }, icon('plus'), 'Add parameter')));
  };
  drawParams();
  const env = section('env', 'Environment & Variables', 'Env files, encrypted variables, compose profiles and deploy parameters.',
    h('div', { class: 'fgrid' }, h('label', { class: 'f' }, h('span', {}, 'Server .env file ', h('span', { class: 'hint' }, '(optional)')), inp('envFile', { mono: true, placeholder: '.env or /etc/stacks/app.env' })),
      h('label', { class: 'f' }, h('span', {}, 'Compose profiles ', h('span', { class: 'hint' }, '(--profile, comma separated)')), inp('profiles', { mono: true, placeholder: 'web,worker' }))),
    h('div', {}, h('div', { class: 'f', style: { fontWeight: 500, marginBottom: '6px' } }, 'Encrypted inline variables'), envSlot),
    h('div', {}, h('div', { class: 'f', style: { fontWeight: 500, marginBottom: '6px' } }, 'Deploy parameters'), paramSlot));

  // --- steps
  const steps = section('steps', 'Build Steps & Lifecycle', 'Hooks, image handling and the health gate.',
    h('div', { class: 'callout' }, icon('shield'), h('div', {}, 'Hooks run as shell commands with the permissions of the hub/agent (root-equivalent via Docker). They never see Dockup’s own secrets.')),
    h('div', { class: 'fgrid' },
      h('label', { class: 'f' }, 'Pre-deploy command', inp('preCmd', { tag: 'textarea', mono: true, rows: '3', placeholder: './scripts/migrate.sh' })),
      h('label', { class: 'f' }, 'Post-deploy command', inp('postCmd', { tag: 'textarea', mono: true, rows: '3', placeholder: 'docker image prune -f' })),
      h('label', { class: 'f' }, 'Health check URL', inp('healthUrl', { mono: true, placeholder: 'http://localhost:3000/healthz' }), h('span', { class: 'hint' }, 'Polled after deploy/restart until it answers < 400.')),
      h('label', { class: 'f' }, 'Health timeout (seconds)', inp('healthSecs', { type: 'number', min: '0', max: '3600', placeholder: '60' }))),
    h('div', { class: 'grid g2' }, chk('pull', 'Pull images before deploy', 'docker compose pull --ignore-buildable'), chk('noCache', 'Build without cache', 'docker compose build --no-cache')));

  // --- triggers
  const cronPrev = h('div', { class: 'hint' });
  let cronT = null;
  const previewCron = () => {
    clearTimeout(cronT);
    cronT = setTimeout(async () => {
      clear(cronPrev);
      if (!f.cron.trim()) { cronPrev.append('No schedule. Examples: ', h('code', {}, '0 3 * * *'), ' daily 03:00 · ', h('code', {}, '*/15 * * * *'), ' every 15 min · ', h('code', {}, '@hourly')); return; }
      try {
        const r = await api('GET', '/api/cron/preview?spec=' + encodeURIComponent(f.cron));
        cronPrev.append(h('span', { class: 'okc' }, '✓ '), 'Next runs (' + r.zone + '): ', r.next.map((t) => new Date(t).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })).join(' · '));
      } catch (e) { cronPrev.append(h('span', { class: 'badc' }, '✗ ' + e.message)); }
    }, 300);
  };
  previewCron();
  const trig = section('triggers', 'Triggers & Scheduling', 'Webhooks, schedules and git polling.',
    p ? h('label', { class: 'f' }, 'Webhook endpoint', h('div', { class: 'code-box' }, h('code', {}, p.webhookUrl), copyBtn(p.webhookUrl, 'Webhook URL')),
      h('span', { class: 'hint' }, 'Secret and setup instructions are on the project’s ', h('a', { href: '#/projects/' + p.id + '/webhook' }, 'Webhook & API'), ' tab.'))
      : h('div', { class: 'hint' }, 'A webhook URL and secret are generated when the project is created.'),
    h('label', { class: 'f' }, 'Schedule (cron, 5 fields, server time)', inp('cron', { mono: true, placeholder: '0 3 * * *', after: previewCron }), cronPrev),
    chk('pollScm', 'Poll git: only build if the branch changed', 'With a schedule on a git project, checks the remote head first and deploys only when it moved.'));

  // --- retention & alerts
  const notifyTa = h('textarea', { class: 'inp', rows: '3', placeholder: 'https://hooks.slack.com/services/…\nhttps://discord.com/api/webhooks/…', spellcheck: 'false' });
  notifyTa.value = f.notifyUrls.join('\n');
  notifyTa.addEventListener('input', () => { f.notifyUrls = notifyTa.value.split('\n').map((s) => s.trim()).filter(Boolean); markDirty(); });
  const notifySel = h('select', { class: 'inp', onchange: () => { f.notifyOn = notifySel.value; markDirty(); } },
    [['always', 'Every build'], ['failure', 'Failures only'], ['change', 'When the status changes']].map(([v, l]) => h('option', { value: v, selected: f.notifyOn === v }, l)));
  const alerts = section('alerts', 'Retention & Alerts', 'History size and notifications.',
    h('div', { class: 'fgrid' }, h('label', { class: 'f' }, 'Builds to keep', inp('keep', { type: 'number', min: '0', max: '1000', placeholder: '50' })),
      h('label', { class: 'f' }, 'Notify on', notifySel),
      h('label', { class: 'f full' }, h('span', {}, 'Notification webhooks ', h('span', { class: 'hint' }, '(Slack, Discord or any JSON endpoint, one per line)')), notifyTa)));

  // --- danger
  const danger = p ? section('danger', 'Danger zone', 'Disable automation or delete the project.',
    chk('disabled', 'Disable this project', 'Webhooks and schedules are ignored while disabled.'),
    h('div', { class: 'flex between wrap' }, h('div', {}, h('b', {}, 'Delete project'), h('div', { class: 'hint' }, 'Removes the project and its build history. Running containers are not touched.')),
      h('button', { type: 'button', class: 'btn danger', onclick: async () => {
        if (!(await confirmDialog('Delete ' + p.name + '?', 'This deletes the project, its webhook and all build logs. Containers keep running; use “Down” first to remove them.', { danger: true, ok: 'Delete project', typeToConfirm: p.name }))) return;
        try { await api('DELETE', '/api/projects/' + p.id); toast('Project deleted'); dirty = false; leaveGuard = null; go('#/projects'); } catch (e) { fail(e); }
      } }, icon('trash'), 'Delete project'))) : null;

  const secs = [['general', 'General'], ['source', 'Source & Compose'], ['env', 'Environment & Secrets'], ['steps', 'Build Steps & Lifecycle'], ['triggers', 'Triggers & Scheduling'], ['alerts', 'Retention & Alerts']];
  if (p) secs.push(['danger', 'Danger zone']);
  const navEl = h('nav', { class: 'cfg-nav' }, h('div', { class: 'nav-h' }, 'Config sections'), secs.map(([k, l]) => h('a', { href: '#', dataset: { sec: k }, onclick: (e) => { e.preventDefault(); $('#sec-' + k).scrollIntoView({ behavior: 'smooth', block: 'start' }); } }, l)));
  const saveBtn = h('button', { type: 'button', class: 'btn primary', onclick: () => save() }, icon('check'), isNew ? 'Create project' : 'Save changes');
  const savebar = h('div', { class: 'savebar' + (isNew ? '' : ' hidden') }, h('div', { class: 's' }, h('span', { class: 'dot warn' }), isNew ? 'New project' : 'Unsaved changes'),
    h('div', { class: 'flex' }, h('button', { type: 'button', class: 'btn ghost', onclick: () => { dirty = false; leaveGuard = null; go(p ? '#/projects/' + p.id : '#/projects'); } }, 'Cancel'), saveBtn));
  const save = async () => {
    if (!f.name.trim()) { toast('Project name is required', 'bad'); $('#sec-general').scrollIntoView({ block: 'start' }); return; }
    const body = Object.assign({}, f, { params: f.params.filter((x) => x.name.trim()) });
    if (envRows !== null) body.env = envText(envRows);
    saveBtn.disabled = true;
    try {
      const saved = await api(isNew ? 'POST' : 'PUT', isNew ? '/api/projects' : '/api/projects/' + p.id, body);
      dirty = false; leaveGuard = null; toast(isNew ? 'Project created' : 'Changes saved');
      go('#/projects/' + saved.id);
    } catch (e) { fail(e); } finally { saveBtn.disabled = false; }
  };
  const hdr = pageHeader([{ label: 'projects', href: '#/projects' }, p ? { label: p.name, href: '#/projects/' + p.id } : null, { label: isNew ? 'new' : 'configure' }].filter(Boolean),
    isNew ? 'New compose project' : 'Configure ' + p.name, { sub: 'Compose stack, target server, repository, variables, hooks and triggers.',
      actions: [p ? h('a', { class: 'btn ghost', href: '#/projects/' + p.id }, icon('left'), 'Back to project') : null] });
  main.append(hdr, h('div', { class: 'cfg' }, navEl, h('div', {}, general, source, env, steps, trig, alerts, danger, savebar)));
  // scroll spy
  const spy = () => {
    let cur = secs[0][0];
    for (const [k] of secs) { const el = $('#sec-' + k); if (el && el.getBoundingClientRect().top < 140) cur = k; }
    navEl.querySelectorAll('a').forEach((a) => a.classList.toggle('on', a.dataset.sec === cur));
  };
  window.addEventListener('scroll', spy, { passive: true }); onCleanup(() => window.removeEventListener('scroll', spy)); spy();
  const keySave = (e) => { if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's') { e.preventDefault(); save(); } };
  document.addEventListener('keydown', keySave); onCleanup(() => document.removeEventListener('keydown', keySave));
  leaveGuard = () => (isNew ? (f.name.trim() ? 'Discard this new project?' : null) : (dirty ? 'You have unsaved changes. Leave anyway?' : null));
  onCleanup(() => { leaveGuard = null; });
}

// =====================================================================
//  servers
// =====================================================================
function installCommands(hub, tok) {
  return {
    docker: `docker run -d --name dockup-agent --restart unless-stopped \\\n  -v /var/run/docker.sock:/var/run/docker.sock \\\n  -v /var/lib/dockup-agent:/var/lib/dockup-agent \\\n  -v /opt/stacks:/opt/stacks \\\n  -e DOCKUP_HUB="${hub}" \\\n  -e DOCKUP_TOKEN="${tok}" \\\n  dockup:latest agent`,
    apt: `sudo apt install ./dockup_${boot.version}_amd64.deb   # or _arm64\nsudo tee /etc/dockup/agent.env >/dev/null <<'EOF'\nDOCKUP_HUB=${hub}\nDOCKUP_TOKEN=${tok}\nEOF\nsudo systemctl disable --now dockup      # this machine is only an agent\nsudo systemctl enable --now dockup-agent`,
    binary: `DOCKUP_HUB="${hub}" DOCKUP_TOKEN="${tok}" \\\n  DOCKUP_DATA=/var/lib/dockup-agent ./dockup agent`,
  };
}
function tokenPanel(res, waitFor) {
  const cmds = installCommands(res.hub, res.token);
  let tab = 'docker';
  const pre = h('pre', { class: 'code' });
  const tabsEl = h('div', { class: 'seg' });
  const drawTabs = () => { clear(tabsEl).append([['docker', 'Docker'], ['apt', 'apt / Debian'], ['binary', 'Binary']].map(([k, l]) => h('button', { type: 'button', class: tab === k ? 'on' : '', onclick: () => { tab = k; drawTabs(); } }, l))); pre.textContent = cmds[tab]; };
  drawTabs();
  const status = h('div', { class: 'callout info' }, icon('loader', 'spin'), h('div', {}, 'Waiting for the agent to connect…'));
  if (waitFor) {
    const iv = setInterval(async () => {
      try {
        const list = await api('GET', '/api/servers');
        const s = list.find((x) => x.id === res.id);
        if (s && s.online) { clearInterval(iv); clear(status).append(icon('check'), h('div', {}, h('b', {}, s.name + ' connected. '), (s.hostname || '') + (s.docker ? ' · Docker ' + s.docker : ''))); status.className = 'callout info'; status.style.borderColor = 'var(--ok)'; }
      } catch (e) { clearInterval(iv); }
    }, 2500);
    onCleanup(() => clearInterval(iv));
    status._stop = () => clearInterval(iv);
  }
  const panel = h('div', { class: 'stack' },
    h('div', { class: 'secret-once' }, h('div', { class: 'flex between' }, h('b', {}, 'Agent token'), h('span', { class: 'hint' }, 'shown once, store it safely')),
      h('div', { class: 'code-box' }, h('code', {}, res.token), copyBtn(res.token, 'Token'))),
    h('div', { class: 'flex between' }, h('b', {}, 'Install on the remote host'), tabsEl), pre, h('div', {}, copyBtn(() => cmds[tab], 'Install command')),
    h('div', { class: 'hint' }, 'The agent dials out to ', h('code', {}, res.hub), ' (no inbound ports). Use https in production.'), waitFor ? status : null);
  panel._status = status;
  return panel;
}
function addServerFlow() {
  const name = h('input', { class: 'inp', placeholder: 'prod-eu-1', maxlength: '60' });
  let created = null;
  const m = modal({
    title: 'Add deployment server', sub: 'Install the Dockup agent on a Linux host with Docker to deploy stacks there.', wide: true,
    body: h('label', { class: 'f' }, 'Server name', name),
    actions: [{ label: 'Cancel', run: () => true }, { label: 'Create & get token', kind: 'primary', run: async () => {
      if (!name.value.trim()) throw new Error('Give the server a name');
      created = await api('POST', '/api/servers', { name: name.value.trim() });
      const body = m.el.querySelector('.modal-b');
      clear(body).append(tokenPanel(created, true));
      clear(m.foot).append(h('button', { type: 'button', class: 'btn primary', onclick: () => { m.close(); route(); } }, 'Done'));
      return false;
    } }],
    onclose: () => { if (created) route(); },
  });
}
async function pageServers() {
  const main = shell('servers');
  const slot = h('div');
  main.append(pageHeader([{ label: 'infrastructure' }, { label: 'servers' }], 'Deployment Servers (Agents)', { sub: 'Linux hosts that run the Dockup agent for Compose deployments. The hub itself is the local server.',
    actions: can('admin') ? [h('button', { class: 'btn go', type: 'button', onclick: addServerFlow }, icon('plus'), 'Add server')] : null }), slot);
  const draw = (list) => {
    clear(slot).append(h('div', { class: 'tbl-wrap' }, h('table', {},
      h('thead', {}, h('tr', {}, h('th', {}, 'Status'), h('th', {}, 'Server name'), h('th', {}, 'Host'), h('th', {}, 'OS'), h('th', {}, 'Docker engine'), h('th', {}, 'Agent version'), h('th', { class: 'num' }, 'Projects'), h('th', {}, 'Activity'), h('th', {}, 'Last seen'), h('th', { class: 'num' }, ''))),
      h('tbody', {}, list.map((s) => h('tr', { style: s.online ? null : { opacity: '.75' } },
        h('td', {}, h('span', { class: 'badge ' + (s.online ? 'ok' : 'bad') }, h('span', { class: 'dot sm ' + (s.online ? 'ok' : 'bad') }), s.online ? 'online' : 'offline')),
        h('td', {}, h('div', { class: 'flex gap4' }, h('span', { class: 'cell-title m' }, s.name), s.local ? h('span', { class: 'badge run' }, 'hub') : null)),
        h('td', { class: 'm' }, s.hostname || '–'), h('td', { class: 'mute' }, s.os || '–'), h('td', { class: 'm' }, s.docker ? 'v' + s.docker : h('span', { class: 'warnc' }, s.online ? 'not found' : '–')),
        h('td', { class: 'm' }, s.version ? (s.version === boot.version ? s.version : h('span', { class: 'warnc', title: 'Hub runs ' + boot.version }, s.version + ' ⚠')) : '–'),
        h('td', { class: 'num m' }, s.projects), h('td', {}, s.running.length ? h('span', { class: 'runc m' }, h('span', { class: 'dot sm run pulse' }), ' ', s.running.join(', ')) : h('span', { class: 'faint m' }, 'idle')),
        h('td', { class: 'm ' + (s.online ? 'mute' : 'badc') }, s.local ? 'now' : isZero(s.lastSeen) ? 'never connected' : ago(s.lastSeen)),
        h('td', { class: 'num' }, can('admin') && !s.local ? h('button', { class: 'btn xs ghost', type: 'button', 'aria-label': 'Server actions', onclick: (e) => {
          e.stopPropagation();
          menu(e.currentTarget, [{ header: s.name },
            { label: 'Rotate token / reinstall', icon: 'refresh', run: async () => {
              if (!(await confirmDialog('Rotate token for ' + s.name + '?', 'The current agent is disconnected immediately and must be restarted with the new token.', { danger: true, ok: 'Rotate token' }))) return;
              try { const r = await api('POST', '/api/servers/' + s.id + '/rotate'); modal({ title: 'New token for ' + s.name, wide: true, body: tokenPanel(Object.assign({ id: s.id }, r), true), actions: [{ label: 'Done', kind: 'primary', run: () => { route(); return true; } }] }); } catch (x) { fail(x); }
            } },
            { label: 'Remove server', icon: 'trash', danger: true, run: async () => {
              if (!(await confirmDialog('Remove ' + s.name + '?', 'Its token stops working. Servers that still have projects cannot be removed.', { danger: true, ok: 'Remove' }))) return;
              try { await api('DELETE', '/api/servers/' + s.id); toast('Server removed'); route(); } catch (x) { fail(x); }
            } }]);
        } }, icon('more')) : null)))))));
  };
  const load = async () => { try { draw(await api('GET', '/api/servers')); } catch (e) { fail(e); } };
  await refreshSummary(); refreshNav();
  await load();
  every(5000, load);
}

// =====================================================================
//  admin: credentials, users, settings, audit
// =====================================================================
function adminOnly(main) { if (!can('admin')) { main.append(emptyState('lock', 'Admins only', 'Ask an admin for access.')); return true; } return false; }

async function pageCredentials() {
  const main = shell('credentials'); if (adminOnly(main)) return;
  const list = await api('GET', '/api/credentials').catch((e) => { fail(e); return []; });
  const add = () => {
    let kind = 'git-token';
    const name = h('input', { class: 'inp', placeholder: 'github-deploy-key' });
    const secret = h('textarea', { class: 'inp', rows: '5', spellcheck: 'false', autocomplete: 'off' });
    const seg = h('div', { class: 'seg' });
    const hint = h('div', { class: 'hint' });
    const drawSeg = () => {
      clear(seg).append([['git-token', 'Access token (HTTPS)'], ['ssh-key', 'SSH private key']].map(([k, l]) => h('button', { type: 'button', class: kind === k ? 'on' : '', onclick: () => { kind = k; drawSeg(); } }, l)));
      secret.placeholder = kind === 'ssh-key' ? '-----BEGIN OPENSSH PRIVATE KEY-----\n…' : 'ghp_… / glpat-…';
      hint.textContent = kind === 'ssh-key' ? 'Use a read-only deploy key. Host keys are trusted on first use (accept-new).' : 'Sent as HTTP basic auth to the repository host only. Use a read-only, fine-grained token.';
    };
    drawSeg();
    modal({ title: 'Add credential', sub: 'Encrypted at rest with AES-256-GCM. The secret is never shown again.', body: [h('label', { class: 'f' }, 'Name', name), h('div', { class: 'f' }, 'Type', seg), h('label', { class: 'f' }, 'Secret', secret), hint],
      actions: [{ label: 'Cancel', run: () => true }, { label: 'Save credential', kind: 'primary', run: async () => { await api('POST', '/api/credentials', { name: name.value.trim(), kind, secret: secret.value }); toast('Credential saved'); route(); } }] });
  };
  main.append(pageHeader([{ label: 'admin' }, { label: 'credentials' }], 'Git Credentials', { sub: 'Encrypted secrets for checking out private repositories.',
    actions: [h('button', { class: 'btn go', type: 'button', onclick: add }, icon('plus'), 'Add credential')] }),
  h('div', { class: 'tbl-wrap' }, list.length ? h('table', {},
    h('thead', {}, h('tr', {}, h('th', {}, 'Name'), h('th', {}, 'Type'), h('th', {}, 'Fingerprint'), h('th', {}, 'Used by'), h('th', {}, 'Last used'), h('th', {}, 'Created'), h('th', { class: 'num' }, ''))),
    h('tbody', {}, list.map((c) => h('tr', {},
      h('td', {}, h('div', { class: 'flex' }, icon('key'), h('span', { class: 'cell-title m' }, c.name))),
      h('td', {}, h('span', { class: 'tag' }, c.kind === 'ssh-key' ? 'SSH private key' : 'Access token')),
      h('td', { class: 'm mute' }, c.hint || '–'),
      h('td', {}, (c.projects || []).length ? c.projects.join(', ') : h('span', { class: 'faint' }, 'unused')),
      h('td', { class: 'm mute', title: c.usedBy || '' }, isZero(c.used) ? 'never' : ago(c.used) + (c.usedBy ? ' (' + c.usedBy + ')' : '')),
      h('td', { class: 'm mute' }, fmtDate(c.created)),
      h('td', { class: 'num' }, h('button', { class: 'btn xs danger', type: 'button', onclick: async () => {
        if (!(await confirmDialog('Delete ' + c.name + '?', 'Projects using it must be switched to another credential first.', { danger: true, ok: 'Delete' }))) return;
        try { await api('DELETE', '/api/credentials/' + c.id); toast('Credential deleted'); route(); } catch (e) { fail(e); }
      } }, 'Delete'))))))
    : emptyState('key', 'No credentials', 'Add an access token or SSH deploy key to deploy private repositories.')));
}

async function pageUsers() {
  const main = shell('users'); if (adminOnly(main)) return;
  const users = await api('GET', '/api/users').catch((e) => { fail(e); return []; });
  const invite = () => {
    let mode = boot.providers.length ? 'sso' : 'local';
    const body = h('div', { class: 'stack' });
    const f = { login: '', name: '', email: '', role: 'viewer', password: '' };
    const field = (k, label, attrs) => h('label', { class: 'f' }, label, h('input', Object.assign({ class: 'inp', value: f[k], oninput: (e) => { f[k] = e.target.value; } }, attrs || {})));
    const roleSel = () => h('label', { class: 'f' }, 'Role', h('select', { class: 'inp', onchange: (e) => { f.role = e.target.value; } },
      [['viewer', 'Viewer: read-only'], ['operator', 'Operator: deploy, restart, cancel'], ['admin', 'Admin: everything, including hooks (root-equivalent)']].map(([v, l]) => h('option', { value: v, selected: f.role === v }, l))));
    const draw = () => {
      clear(body).append(h('div', { class: 'seg' }, h('button', { type: 'button', class: mode === 'sso' ? 'on' : '', onclick: () => { mode = 'sso'; draw(); } }, icon('mail'), 'Invite by email (SSO)'),
        h('button', { type: 'button', class: mode === 'local' ? 'on' : '', onclick: () => { mode = 'local'; draw(); } }, icon('lock'), 'Local account')));
      if (mode === 'sso') {
        body.append(boot.providers.length ? null : h('div', { class: 'callout' }, icon('alert'), h('div', {}, 'No SSO provider is configured yet; the invite works once one is. See System Settings.')),
          field('email', 'Email', { type: 'email', placeholder: 'alex@company.com' }), roleSel(), h('div', { class: 'hint' }, 'Binds on first sign-in with a provider that reports this email as verified.'));
      } else {
        body.append(h('div', { class: 'fgrid' }, field('login', 'Username', { autocomplete: 'off' }), field('name', 'Display name')), field('password', 'Initial password (10+ chars)', { type: 'password', autocomplete: 'new-password' }), roleSel());
      }
    };
    draw();
    modal({ title: 'Add user', body, actions: [{ label: 'Cancel', run: () => true }, { label: 'Add user', kind: 'primary', run: async () => {
      const b = mode === 'sso' ? { email: f.email.trim(), role: f.role } : { login: f.login.trim(), name: f.name.trim(), password: f.password, role: f.role };
      await api('POST', '/api/users', b); toast('User added'); route();
    } }] });
  };
  main.append(pageHeader([{ label: 'admin' }, { label: 'users' }], 'Users & Roles', { sub: 'Viewer: read-only · Operator: deploy & cancel · Admin: configure (root-equivalent).',
    actions: [h('button', { class: 'btn go', type: 'button', onclick: invite }, icon('plus'), 'Add user')] }),
  h('div', { class: 'tbl-wrap' }, h('table', {},
    h('thead', {}, h('tr', {}, h('th', {}, 'User'), h('th', {}, 'Sign-in'), h('th', {}, 'Role'), h('th', {}, 'Status'), h('th', {}, 'Last login'), h('th', {}, 'Created'), h('th', { class: 'num' }, ''))),
    h('tbody', {}, users.map((u) => {
      const me = u.id === boot.user.id;
      const roleSel = h('select', { class: 'inp sm', style: { width: '120px' }, disabled: me, title: me ? 'You cannot change your own role' : '', onchange: async (e) => {
        try { await api('PUT', '/api/users/' + u.id, { role: e.target.value }); toast(u.login + ' is now ' + e.target.value); } catch (x) { fail(x); e.target.value = u.role; }
      } }, ['viewer', 'operator', 'admin'].map((r) => h('option', { value: r, selected: r === u.role }, r)));
      const acts = [];
      if (u.provider === 'local') acts.push(h('button', { class: 'btn xs', type: 'button', onclick: () => {
        const pw = h('input', { class: 'inp', type: 'password', autocomplete: 'new-password' });
        modal({ title: 'Reset password for ' + u.login, sub: 'All of their sessions are signed out.', body: h('label', { class: 'f' }, 'New password (10+ chars)', pw),
          actions: [{ label: 'Cancel', run: () => true }, { label: 'Reset password', kind: 'primary', run: async () => { await api('PUT', '/api/users/' + u.id, { password: pw.value }); toast('Password reset'); } }] });
      } }, 'Reset password'));
      if (!me) acts.push(h('button', { class: 'btn xs danger', type: 'button', onclick: async () => {
        if (!(await confirmDialog('Remove ' + (u.name || u.login) + '?', 'Their sessions and API tokens are revoked immediately.', { danger: true, ok: 'Remove user' }))) return;
        try { await api('DELETE', '/api/users/' + u.id); toast('User removed'); route(); } catch (e) { fail(e); }
      } }, 'Remove'));
      return h('tr', {},
        h('td', {}, h('div', { class: 'flex' }, h('span', { class: 'avatar', style: { cursor: 'default' } }, initials(u)),
          h('div', {}, h('div', { class: 'cell-title' }, (u.name || u.login) + (me ? ' (you)' : '')), h('span', { class: 'cell-sub' }, u.email || u.login)))),
        h('td', {}, h('span', { class: 'tag' }, u.invited ? 'invited' : u.provider === 'local' ? 'password' : u.provider)),
        h('td', {}, roleSel),
        h('td', {}, u.invited ? h('span', { class: 'badge warn' }, 'pending') : h('span', { class: 'badge ok' }, 'active')),
        h('td', { class: 'm mute' }, isZero(u.lastLogin) ? 'never' : ago(u.lastLogin)), h('td', { class: 'm mute' }, fmtDate(u.created)),
        h('td', { class: 'num' }, h('div', { class: 'row-actions', style: { justifyContent: 'flex-end' } }, acts)));
    })))));
}

async function pageSettings() {
  const main = shell('settings'); if (adminOnly(main)) return;
  let d;
  try { d = await api('GET', '/api/settings'); } catch (e) { fail(e); return; }
  const st = d.settings;
  const allow = h('textarea', { class: 'inp', rows: '3', placeholder: '@company.com\nalex@partner.io', spellcheck: 'false' });
  allow.value = (st.allowedEmail || []).join('\n');
  const signup = h('input', { type: 'checkbox', checked: st.allowSignup });
  const role = h('select', { class: 'inp' }, [['viewer', 'Viewer'], ['operator', 'Operator']].map(([v, l]) => h('option', { value: v, selected: st.defaultRole === v }, l)));
  const warn = [];
  if (!boot.tls) warn.push(h('div', { class: 'callout' }, icon('alert'), h('div', {}, h('b', {}, 'Not served over HTTPS. '), 'Put Caddy, nginx or Traefik with TLS in front, then set DOCKUP_URL=https://… and DOCKUP_TRUST_PROXY=1.')));
  if (d.providers.length && !d.publicUrl) warn.push(h('div', { class: 'callout' }, icon('alert'), h('div', {}, 'SSO is enabled but DOCKUP_URL is not set: callback URLs are derived from the request Host header.')));
  const provs = [['github', 'GitHub', 'DOCKUP_GITHUB_CLIENT_ID / _SECRET'], ['google', 'Google Workspace', 'DOCKUP_GOOGLE_CLIENT_ID / _SECRET'], ['oidc', 'Corporate SSO (OIDC)', 'DOCKUP_OIDC_ISSUER / _CLIENT_ID / _CLIENT_SECRET / _NAME']];
  main.append(pageHeader([{ label: 'admin' }, { label: 'settings' }], 'System Settings', { sub: 'Single sign-on, sign-up policy and instance information.' }), warn,
    h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('div', {}, h('h2', {}, icon('shield'), 'SSO & OAuth providers'), h('div', { class: 'sub' }, 'Configured with environment variables on the hub; restart after changes.'))),
      h('div', { class: 'card-b stack' }, provs.map(([k, l, envs]) => {
        const on = d.providers.includes(k);
        return h('div', { class: 'flex between wrap', style: { gap: '12px', paddingBottom: '10px', borderBottom: '1px solid var(--line-soft)' } },
          h('div', {}, h('div', { class: 'flex' }, h('b', {}, l), h('span', { class: 'badge ' + (on ? 'ok' : 'mute') }, on ? 'enabled' : 'not configured')), h('div', { class: 'hint m' }, envs)),
          on ? h('div', { class: 'code-box', style: { maxWidth: '520px', flex: 1 } }, h('code', {}, d.callbacks[k]), copyBtn(d.callbacks[k], 'Callback URL')) : null);
      }))),
    h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('div', {}, h('h2', {}, icon('users'), 'Auto sign-up'), h('div', { class: 'sub' }, 'Let people from your organisation create their own account on first SSO sign-in.'))),
      h('div', { class: 'card-b stack' },
        h('label', { class: 'check' }, signup, h('div', {}, h('b', {}, 'Allow auto sign-up'), h('span', { class: 'hint' }, 'Only verified emails that match the allow-list. Without sign-up, admins invite people by email.'))),
        h('div', { class: 'fgrid' }, h('label', { class: 'f' }, 'Allowed emails / domains (one per line)', allow), h('label', { class: 'f' }, 'Role for new accounts', role, h('span', { class: 'hint' }, 'Admin can never be granted automatically.'))),
        h('div', {}, h('button', { class: 'btn primary', type: 'button', onclick: async () => {
          try { await api('PUT', '/api/settings', { allowSignup: signup.checked, allowedEmail: allow.value.split('\n').map((s) => s.trim()).filter(Boolean), defaultRole: role.value }); toast('Settings saved'); } catch (e) { fail(e); }
        } }, icon('check'), 'Save sign-up policy')))),
    h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('h2', {}, icon('info'), 'Instance')),
      h('div', { class: 'card-b' }, h('div', { class: 'agent', style: { border: 0, padding: 0, background: 'none' } }, h('dl', { style: { maxWidth: '520px' } },
        h('dt', {}, 'Version'), h('dd', {}, d.version), h('dt', {}, 'Public URL'), h('dd', {}, d.publicUrl || '(from request)'),
        h('dt', {}, 'Behind trusted proxy'), h('dd', {}, d.trustProxy ? 'yes' : 'no'), h('dt', {}, 'Transport'), h('dd', { class: boot.tls ? 'okc' : 'warnc' }, boot.tls ? 'HTTPS' : 'HTTP'))))));
}

async function pageAudit() {
  const main = shell('audit'); if (adminOnly(main)) return;
  const slot = h('div');
  const q = h('input', { class: 'inp', placeholder: 'Search actor, action, target, IP…' });
  let t = null;
  q.addEventListener('input', () => { clearTimeout(t); t = setTimeout(load, 250); });
  const exp = h('a', { class: 'btn', href: '/api/audit?format=csv' }, icon('download'), 'Export CSV');
  main.append(pageHeader([{ label: 'admin' }, { label: 'audit' }], 'Security & Audit Log', { sub: 'Who did what, when and from where. Keeps the latest 5,000 events.', actions: [exp] }), h('div', { class: 'filters' }, q), slot);
  const color = (a) => /failed|denied|delete|revoke|rotate/.test(a) ? 'bad' : /login|setup/.test(a) ? 'run' : /build/.test(a) ? 'ok' : 'mute';
  async function load() {
    try {
      const list = await api('GET', '/api/audit?q=' + encodeURIComponent(q.value.trim()));
      exp.href = '/api/audit?format=csv&q=' + encodeURIComponent(q.value.trim());
      clear(slot).append(h('div', { class: 'tbl-wrap' }, list.length ? h('table', {},
        h('thead', {}, h('tr', {}, h('th', {}, 'Timestamp'), h('th', {}, 'Actor'), h('th', {}, 'Action'), h('th', {}, 'Target'), h('th', {}, 'IP address'))),
        h('tbody', {}, list.map((a) => h('tr', {}, h('td', { class: 'm mute nowrap' }, new Date(a.time).toLocaleString()), h('td', { class: 'm' }, a.user),
          h('td', {}, h('span', { class: 'badge ' + color(a.action) }, a.action.replace(/\./g, '_'))), h('td', { class: 'm ellip', style: { maxWidth: '360px' } }, a.target || '–'), h('td', { class: 'm mute' }, a.ip || '–')))))
        : emptyState('list', 'No events', q.value ? 'Nothing matches.' : 'Actions are recorded here.')));
    } catch (e) { fail(e); }
  }
  await load();
}

// =====================================================================
//  account
// =====================================================================
async function pageAccount() {
  const main = shell('account');
  const u = boot.user;
  const [sessions, tokens] = await Promise.all([api('GET', '/api/me/sessions').catch(() => []), api('GET', '/api/tokens').catch(() => [])]);
  const cur = h('input', { class: 'inp', type: 'password', autocomplete: 'current-password' });
  const nw = h('input', { class: 'inp', type: 'password', autocomplete: 'new-password' });
  const nw2 = h('input', { class: 'inp', type: 'password', autocomplete: 'new-password' });
  const newToken = () => {
    const name = h('input', { class: 'inp', placeholder: 'ci-github-actions' });
    const days = h('select', { class: 'inp' }, [['30', '30 days'], ['90', '90 days'], ['365', '1 year'], ['0', 'No expiry']].map(([v, l]) => h('option', { value: v, selected: v === '90' }, l)));
    const m = modal({ title: 'New API token', sub: 'Acts with your role (' + u.role + '). Use as: Authorization: Bearer dk_…', body: [h('label', { class: 'f' }, 'Name', name), h('label', { class: 'f' }, 'Expires', days)],
      actions: [{ label: 'Cancel', run: () => true }, { label: 'Create token', kind: 'primary', run: async () => {
        const r = await api('POST', '/api/tokens', { name: name.value.trim(), days: parseInt(days.value, 10) });
        clear(m.el.querySelector('.modal-b')).append(h('div', { class: 'secret-once' }, h('b', {}, 'Copy your token now: it is not shown again.'), h('div', { class: 'code-box' }, h('code', {}, r.token), copyBtn(r.token, 'Token'))));
        clear(m.foot).append(h('button', { type: 'button', class: 'btn primary', onclick: () => { m.close(); route(); } }, 'Done'));
        return false;
      } }] });
  };
  main.append(pageHeader([{ label: 'account' }], 'Account & Security', { sub: 'Your profile, password, active sessions and API tokens.' }),
    h('div', { class: 'grid g2', style: { alignItems: 'start' } },
      h('div', { class: 'card pad' }, h('div', { class: 'flex gap12' }, h('span', { class: 'avatar', style: { width: '44px', height: '44px', fontSize: '14px', cursor: 'default' } }, initials(u)),
        h('div', {}, h('h3', {}, u.name || u.login), h('div', { class: 'sub m' }, u.email || u.login), h('div', { class: 'flex mt8' }, h('span', { class: 'role ' + u.role }, u.role), h('span', { class: 'tag' }, u.provider === 'local' ? 'password sign-in' : u.provider + ' sign-in'))))),
      u.provider === 'local' ? h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('h2', {}, icon('lock'), 'Change password')),
        h('form', { class: 'card-b stack', onsubmit: async (e) => {
          e.preventDefault();
          if (nw.value !== nw2.value) return toast('New passwords do not match', 'bad');
          try { await api('PUT', '/api/me/password', { current: cur.value, new: nw.value }); toast('Password changed: other sessions were signed out'); cur.value = nw.value = nw2.value = ''; } catch (x) { fail(x); }
        } }, h('label', { class: 'f' }, 'Current password', cur), h('div', { class: 'fgrid' }, h('label', { class: 'f' }, 'New password', nw), h('label', { class: 'f' }, 'Confirm', nw2)),
        h('div', {}, h('button', { class: 'btn primary', type: 'submit' }, 'Update password'))))
        : h('div', { class: 'card pad' }, h('h2', {}, icon('shield'), 'Single sign-on'), h('div', { class: 'sub mt8' }, 'You sign in with ' + u.provider + '. Password and MFA are managed there.'))),
    h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('div', {}, h('h2', {}, icon('globe'), 'Active sessions'), h('div', { class: 'sub' }, 'Browsers signed in to your account.')),
      sessions.length > 1 ? h('button', { class: 'btn sm danger', type: 'button', onclick: async () => { try { const r = await api('DELETE', '/api/me/sessions/others'); toast(r.revoked + ' session(s) signed out'); route(); } catch (e) { fail(e); } } }, 'Sign out other sessions') : null),
    h('table', {}, h('thead', {}, h('tr', {}, h('th', {}, 'Device'), h('th', {}, 'IP'), h('th', {}, 'Signed in'), h('th', {}, 'Last active'), h('th', {}, 'Expires'), h('th', { class: 'num' }, ''))),
      h('tbody', {}, sessions.sort((a, b) => b.current - a.current).map((s) => h('tr', {},
        h('td', {}, h('div', { class: 'ellip', style: { maxWidth: '360px' }, title: s.agent }, uaName(s.agent)), s.current ? h('span', { class: 'badge ok' }, 'this browser') : null),
        h('td', { class: 'm' }, s.ip || '–'), h('td', { class: 'm mute' }, ago(s.created)), h('td', { class: 'm mute' }, ago(s.seen)), h('td', { class: 'm mute' }, fmtDate(s.expires)),
        h('td', { class: 'num' }, s.current ? null : h('button', { class: 'btn xs danger', type: 'button', onclick: async () => { try { await api('DELETE', '/api/me/sessions/' + s.id); toast('Session signed out'); route(); } catch (e) { fail(e); } } }, 'Revoke'))))))),
    h('div', { class: 'card' }, h('div', { class: 'card-h' }, h('div', {}, h('h2', {}, icon('key'), 'API tokens'), h('div', { class: 'sub' }, 'For CI pipelines and scripts. Tokens carry your role.')),
      h('button', { class: 'btn sm go', type: 'button', onclick: newToken }, icon('plus'), 'New token')),
    tokens.length ? h('table', {}, h('thead', {}, h('tr', {}, h('th', {}, 'Name'), h('th', {}, 'Created'), h('th', {}, 'Last used'), h('th', {}, 'Expires'), h('th', { class: 'num' }, ''))),
      h('tbody', {}, tokens.map((t) => {
        const expired = !isZero(t.expires) && new Date(t.expires) < new Date();
        return h('tr', {}, h('td', { class: 'm cell-title' }, t.name), h('td', { class: 'm mute' }, fmtDate(t.created)), h('td', { class: 'm mute' }, isZero(t.used) ? 'never' : ago(t.used)),
          h('td', { class: 'm ' + (expired ? 'badc' : 'mute') }, isZero(t.expires) ? 'never' : expired ? 'expired' : fmtDate(t.expires)),
          h('td', { class: 'num' }, h('button', { class: 'btn xs danger', type: 'button', onclick: async () => {
            if (!(await confirmDialog('Revoke token ' + t.name + '?', 'Scripts using it stop working immediately.', { danger: true, ok: 'Revoke' }))) return;
            try { await api('DELETE', '/api/tokens/' + t.id); toast('Token revoked'); route(); } catch (e) { fail(e); }
          } }, 'Revoke')));
      }))) : h('div', { class: 'card-b' }, h('div', { class: 'faint' }, 'No API tokens.'))));
}
function uaName(ua) {
  if (!ua) return 'Unknown device';
  const b = /Edg\//.test(ua) ? 'Edge' : /Firefox\//.test(ua) ? 'Firefox' : /Chrome\//.test(ua) ? 'Chrome' : /Safari\//.test(ua) ? 'Safari' : /curl/.test(ua) ? 'curl' : 'Browser';
  const o = /Windows/.test(ua) ? 'Windows' : /Mac OS X|Macintosh/.test(ua) ? 'macOS' : /Android/.test(ua) ? 'Android' : /iPhone|iPad/.test(ua) ? 'iOS' : /Linux/.test(ua) ? 'Linux' : '';
  return b + (o ? ' on ' + o : '');
}

// =====================================================================
//  onboarding: first-run guide + getting-started checklist
// =====================================================================
const OB_ITEMS = {
  admin: { t: 'Create your admin account', d: 'Done during setup.', ic: 'user' },
  server: { t: 'Connect a Docker host', d: 'Docker on this server, or an agent online.', ic: 'server', href: '#/welcome/server' },
  project: { t: 'Add your first project', d: 'A Compose stack from git or a folder.', ic: 'folder', href: '#/welcome/project' },
  deploy: { t: 'Run your first deploy', d: 'Watch it live; roll back any time.', ic: 'rocket', href: '#/welcome/deploy' },
  webhook: { t: 'Deploy on every git push', d: 'Add a project webhook in GitHub/GitLab.', ic: 'webhook' },
  team: { t: 'Invite your team', d: 'Viewers, operators and admins.', ic: 'users', href: '#/users' },
  agent: { t: 'Add another server', d: 'The agent needs no inbound ports.', ic: 'server', href: '#/servers' },
  credential: { t: 'Connect a private repo', d: 'Store a token or deploy key.', ic: 'key', href: '#/credentials' },
  sso: { t: 'Enable single sign-on', d: 'GitHub, Google or your OIDC.', ic: 'shield', href: '#/settings' },
};
const wiz = { serverId: null, projectId: null, kind: 'sample' };

async function getOnboarding() {
  try { return await api('GET', '/api/onboarding'); } catch (e) { return null; }
}
async function setDismissed(v) {
  try { await api('PUT', '/api/onboarding', { dismissed: v }); } catch (e) { fail(e); }
}

function gettingStarted(ob, onChange) {
  const steps = ob.steps;
  const done = steps.filter((s) => s.done).length;
  const firstGit = summary && summary.projects.find((p) => p.kind === 'git');
  const card = h('div', { class: 'gs' },
    h('div', { class: 'gs-h' },
      h('div', { class: 'flex gap12' }, brandMark(), h('div', {}, h('h2', {}, 'Get started with Dockup'),
        h('div', { class: 'sub', style: { marginTop: 0 } }, done + ' of ' + steps.length + ' done · the first four get your first stack live'))),
      h('div', { class: 'flex' },
        h('a', { class: 'btn sm primary', href: '#/welcome' }, icon('play'), 'Open the guide'),
        h('button', { class: 'btn sm ghost', type: 'button', title: 'Hide this checklist', onclick: async () => { await setDismissed(true); toast('Checklist hidden: reopen the guide from the account menu', 'info'); onChange(); } }, icon('x'), 'Dismiss'))),
    h('div', { class: 'gs-bar' }, h('i', { style: { width: Math.round(done / steps.length * 100) + '%' } })),
    h('div', { class: 'gs-items' }, steps.map((s) => {
      const it = OB_ITEMS[s.id] || { t: s.id, d: '' };
      let href = it.href;
      if (s.id === 'webhook') href = firstGit ? '#/projects/' + firstGit.id + '/webhook' : '#/projects/new';
      const inner = [h('span', { class: 'ck' }, s.done ? icon('check') : null),
        h('div', {}, h('b', {}, it.t, s.optional && !s.done ? h('span', { class: 'opt' }, 'optional') : null), h('span', {}, it.d))];
      return s.done || !href ? h('div', { class: 'gs-item' + (s.done ? ' done' : '') }, inner) : h('a', { class: 'gs-item', href }, inner);
    })));
  return card;
}

function celebrate() {
  if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  const colors = ['#3B82F6', '#22C55E', '#F59E0B', '#EC4899', '#8B5CF6', '#06B6D4'];
  const box = h('div', { class: 'confetti', 'aria-hidden': 'true' });
  for (let i = 0; i < 90; i++) {
    box.append(h('i', { style: { left: Math.random() * 100 + 'vw', background: colors[i % colors.length], animationDuration: (1.8 + Math.random() * 1.6) + 's',
      animationDelay: (Math.random() * 0.4) + 's', transform: 'rotate(' + Math.random() * 360 + 'deg)', width: (6 + Math.random() * 5) + 'px' } }));
  }
  document.body.append(box);
  setTimeout(() => box.remove(), 4200);
}

const OB_STEPS = [['intro', 'Welcome'], ['server', 'Choose a server'], ['project', 'First project'], ['deploy', 'First deploy']];

async function pageWelcome(step) {
  if (!can('admin')) return go('#/');
  shellEls = null;
  step = OB_STEPS.some((s) => s[0] === step) ? step : 'intro';
  const idx = OB_STEPS.findIndex((s) => s[0] === step);
  const [ob] = await Promise.all([getOnboarding(), refreshSummary()]);
  if (!ob) return go('#/');
  const finish = async (to) => { await setDismissed(true); go(to || '#/'); };
  const body = h('div', { class: 'ob-card' });
  const stepper = h('div', { class: 'ob-steps', role: 'list' });
  OB_STEPS.forEach(([k, label], i) => {
    if (i) stepper.append(h('span', { class: 'ob-line' + (i <= idx ? ' done' : '') }));
    stepper.append(h('a', { href: '#/welcome/' + k, role: 'listitem', class: 'ob-step' + (i === idx ? ' on' : i < idx ? ' done' : ''), style: { textDecoration: 'none' } },
      h('i', {}, i < idx ? icon('check') : String(i + 1)), h('span', {}, label)));
  });
  clear($('#app')).append(h('div', { class: 'ob' },
    h('header', { class: 'ob-top' }, h('a', { class: 'brand', href: '#/', 'aria-label': 'Dockup home' }, brandMark(), wordmark()),
      h('div', { class: 'flex' }, themeBtn(), h('button', { class: 'btn ghost', type: 'button', onclick: () => finish('#/') }, 'Skip setup'))),
    h('div', { class: 'ob-wrap' }, stepper, body, h('div', { class: 'ob-foot' }, 'You can come back to this guide any time from the dashboard checklist.'))));
  const nav = (back, next) => h('div', { class: 'ob-actions' },
    back ? h('a', { class: 'btn ghost', href: '#/welcome/' + back }, icon('left'), 'Back') : h('span'), next);
  const servers = summary ? summary.servers : [];
  const local = servers.find((s) => s.local);
  if (!wiz.serverId) wiz.serverId = local && local.docker ? local.id : ((servers.find((s) => !s.local && s.online) || local || {}).id);

  // ---------- 1. welcome + system check ----------
  if (step === 'intro') {
    const hst = ob.host;
    const rows = [
      ['ok', 'Dockup hub is running', 'v' + hst.version],
      [boot.user ? 'ok' : 'warn', 'Admin account', (boot.user.name || boot.user.login)],
      hst.docker ? ['ok', 'Docker Engine on this server', 'v' + hst.docker]
        : ['warn', 'Docker Engine on this server', 'not found', 'You can still deploy to other servers with the agent (next step), or install Docker here.'],
      hst.compose ? ['ok', 'Docker Compose', 'v' + hst.compose]
        : ['warn', 'Docker Compose', hst.docker ? 'plugin not found' : 'needs Docker', hst.docker ? 'Install the docker-compose-plugin package.' : null],
      hst.https ? ['ok', 'Secure connection', 'HTTPS']
        : ['warn', 'Secure connection', 'HTTP only', 'Fine for a first look. Before inviting your team, put Caddy, nginx or Traefik with TLS in front and set DOCKUP_URL.'],
    ];
    body.append(
      h('div', { class: 'ob-hero' }, brandMark('xl'), h('div', {},
        h('h1', {}, 'Welcome to Dockup, ' + ((boot.user.name || boot.user.login).split(' ')[0]) + '.'),
        h('p', { class: 'lead' }, 'Push-to-deploy for Docker Compose. Let’s get your first stack live; it takes about two minutes.'))),
      h('div', { class: 'ob-features' },
        [['git', 'Git push → deploy', 'Webhooks, schedules and one-click rollback.'], ['terminal', 'Logs you can read', 'Every stage folded, failures explained.'], ['server', 'Agents, no open ports', 'Remote servers dial out to this hub.']]
          .map(([ic, t, d]) => h('div', { class: 'ob-feature' }, h('div', { class: 'ic' }, icon(ic)), h('b', {}, t), h('span', {}, d)))),
      h('div', { class: 'ob-h' }, 'System check'),
      h('div', { class: 'checks' }, rows.map(([st, t, v, hint], i) => h('div', { class: 'check-row', style: { animationDelay: (120 + i * 140) + 'ms' } },
        h('span', { class: 'st ' + st }, icon(st === 'ok' ? 'check' : 'alert')), h('div', {}, h('b', {}, t), hint ? h('span', { class: 'hint' }, hint) : null), h('span', { class: 'v' }, v)))),
      nav(null, h('a', { class: 'btn primary lg', href: '#/welcome/server' }, 'Let’s go', icon('right'))));
    return;
  }

  // ---------- 2. where to deploy ----------
  if (step === 'server') {
    const panel = h('div');
    const choices = h('div', { class: 'choices' });
    const draw = () => {
      clear(choices);
      const opts = servers.filter((s) => s.local || s.online);
      for (const s of opts) {
        const ok = s.local ? !!s.docker : s.online;
        choices.append(h('button', { type: 'button', class: 'choice' + (wiz.serverId === s.id ? ' sel' : ''), disabled: !ok, onclick: () => { wiz.serverId = s.id; clear(panel); draw(); } },
          h('span', { class: 'radio' }), h('div', { class: 'ic' }, icon('server')),
          h('b', {}, s.local ? 'This server' : s.name), h('span', {}, s.local ? (s.hostname || 'the hub') + (s.docker ? ' · Docker ' + s.docker : ' · Docker not found') : 'Agent online · ' + (s.hostname || '') + (s.docker ? ' · Docker ' + s.docker : '')),
          s.local && ok ? h('span', { class: 'faint', style: { fontSize: '11.5px' } }, 'Simplest: stacks run next to Dockup.') : null));
      }
      choices.append(h('button', { type: 'button', class: 'choice' + (wiz.serverId === 'new' ? ' sel' : ''), onclick: () => { wiz.serverId = 'new'; draw(); drawPanel(); } },
        h('span', { class: 'radio' }), h('div', { class: 'ic' }, icon('plus')), h('b', {}, 'Another server'),
        h('span', {}, 'Install a small agent on any Linux host with Docker. It dials out to this hub; no inbound ports needed.')));
    };
    const drawPanel = () => {
      clear(panel);
      if (wiz.serverId !== 'new') return;
      const name = h('input', { class: 'inp', placeholder: 'prod-eu-1', maxlength: '60' });
      const btn = h('button', { class: 'btn primary', type: 'button', onclick: async () => {
        if (!name.value.trim()) return toast('Give the server a name', 'bad');
        btn.disabled = true;
        try {
          const res = await api('POST', '/api/servers', { name: name.value.trim() });
          clear(panel).append(h('div', { class: 'ob-panel' }, tokenPanel(res, true)));
          wiz.serverId = res.id;
          const iv = setInterval(async () => {
            const list = await api('GET', '/api/servers').catch(() => []);
            const s = list.find((x) => x.id === res.id);
            if (s && s.online) { clearInterval(iv); toast(s.name + ' connected'); }
          }, 3000);
          onCleanup(() => clearInterval(iv));
        } catch (e) { fail(e); btn.disabled = false; }
      } }, 'Create & get install command');
      panel.append(h('div', { class: 'ob-panel stack' }, h('label', { class: 'f' }, 'Name this server', h('div', { class: 'input-row' }, name, btn))));
      setTimeout(() => name.focus(), 30);
    };
    draw(); drawPanel();
    body.append(h('h1', {}, 'Where should Dockup deploy?'), h('p', { class: 'lead' }, 'Pick the Docker host for your first project. You can add more servers any time.'),
      h('div', { class: 'mt16' }, choices), panel,
      nav('intro', h('button', { class: 'btn primary lg', type: 'button', onclick: () => {
        if (wiz.serverId === 'new') return toast('Create the server first, or pick an existing one', 'bad');
        go('#/welcome/project');
      } }, 'Continue', icon('right'))));
    return;
  }

  // ---------- 3. first project ----------
  if (step === 'project') {
    const target = servers.find((s) => s.id === wiz.serverId) || local;
    const localDocker = local && local.docker;
    if (wiz.kind === 'sample' && !localDocker) wiz.kind = 'git';
    const form = h('div');
    const choices = h('div', { class: 'choices' });
    const f = { name: '', repo: '', branch: 'main', composeFile: '', path: '' };
    const field = (k, label, attrs, hint) => h('label', { class: 'f' }, label, h('input', Object.assign({ class: 'inp' + (attrs && attrs.mono ? ' mono' : ''), value: f[k], oninput: (e) => { f[k] = e.target.value; } }, attrs || {})), hint ? h('span', { class: 'hint' }, hint) : null);
    const drawForm = () => {
      clear(form);
      if (wiz.kind === 'sample') {
        form.append(h('div', { class: 'ob-panel' }, h('div', { class: 'flex' }, icon('box'), h('b', {}, 'hello-dockup')),
          h('div', { class: 'hint mt8' }, 'Runs ', h('code', {}, 'traefik/whoami'), ' on this server: a tiny web service that answers with request details. It publishes no ports, so it can’t clash with anything. Remove it later with “Down” and “Delete”.')));
      } else if (wiz.kind === 'git') {
        form.append(h('div', { class: 'ob-panel fgrid' },
          field('name', 'Project name', { placeholder: 'payments-api' }),
          field('branch', 'Branch', { mono: true, placeholder: 'main' }),
          h('div', { class: 'full' }, field('repo', 'Repository URL', { mono: true, placeholder: 'https://github.com/acme/app.git' }, 'Public repos work right away. For private ones, add a credential later and select it in Configure.')),
          h('div', { class: 'full' }, field('composeFile', 'Compose file (optional)', { mono: true, placeholder: 'auto: compose.yaml, docker-compose.yml…' }))));
      } else {
        form.append(h('div', { class: 'ob-panel fgrid' },
          field('name', 'Project name', { placeholder: 'blog' }),
          field('path', 'Compose folder on the server', { mono: true, placeholder: '/opt/stacks/blog' }, 'An absolute path on ' + (target ? target.name : 'the server') + ' that contains compose.yaml.')));
      }
      setTimeout(() => { const i = form.querySelector('input'); if (i) i.focus(); }, 30);
    };
    const drawChoices = () => {
      clear(choices).append(
        h('button', { type: 'button', class: 'choice' + (wiz.kind === 'sample' ? ' sel' : ''), disabled: !localDocker, onclick: () => { wiz.kind = 'sample'; drawChoices(); drawForm(); } },
          h('span', { class: 'rec badge ok' }, 'Recommended'), h('div', { class: 'ic' }, icon('play')), h('b', {}, 'Try a sample stack'),
          h('span', {}, localDocker ? 'See a full deploy in ~30 seconds. No setup.' : 'Needs Docker on this server.')),
        h('button', { type: 'button', class: 'choice' + (wiz.kind === 'git' ? ' sel' : ''), onclick: () => { wiz.kind = 'git'; drawChoices(); drawForm(); } },
          h('span', { class: 'radio' }), h('div', { class: 'ic' }, icon('git')), h('b', {}, 'From a Git repository'), h('span', {}, 'Clone, build and deploy a repo with a compose file.')),
        h('button', { type: 'button', class: 'choice' + (wiz.kind === 'path' ? ' sel' : ''), onclick: () => { wiz.kind = 'path'; drawChoices(); drawForm(); } },
          h('span', { class: 'radio' }), h('div', { class: 'ic' }, icon('folder')), h('b', {}, 'From a folder on the server'), h('span', {}, 'Manage a stack that already lives on disk.')));
    };
    drawChoices(); drawForm();
    const create = h('button', { class: 'btn primary lg', type: 'button', onclick: async () => {
      create.disabled = true;
      try {
        let p;
        if (wiz.kind === 'sample') p = await api('POST', '/api/onboarding/sample');
        else {
          const src = wiz.kind === 'git' ? { kind: 'git', repo: f.repo.trim(), branch: f.branch.trim() || 'main', composeFile: f.composeFile.trim() } : { kind: 'path', path: f.path.trim() };
          p = await api('POST', '/api/projects', { name: f.name.trim(), serverId: (target || {}).id, source: src, params: [], notifyUrls: [] });
        }
        wiz.projectId = p.id; toast('Project ' + p.name + ' created');
        go('#/welcome/deploy');
      } catch (e) { fail(e); } finally { create.disabled = false; }
    } }, 'Create project', icon('right'));
    body.append(h('h1', {}, 'Add your first project'), h('p', { class: 'lead' }, 'A project is one Docker Compose stack on ' + (target ? target.name === 'local' ? 'this server' : target.name : 'a server') + '. Start with the sample to see the whole flow, or bring your own.'),
      h('div', { class: 'mt16' }, choices), form, nav('server', create));
    return;
  }

  // ---------- 4. first deploy ----------
  let pid = wiz.projectId || ob.sampleId || ob.firstProject;
  if (!pid) return go('#/welcome/project');
  let p;
  try { p = await api('GET', '/api/projects/' + pid); } catch (e) { return go('#/welcome/project'); }
  const sv = servers.find((s) => s.id === p.serverId);
  const area = h('div');
  const deployBtn = h('button', { class: 'btn go lg', type: 'button' }, icon('rocket'), 'Deploy ' + p.name);
  const hero = h('div', { class: 'deploy-hero' },
    h('div', {}, h('div', { class: 'flex' }, h('b', { style: { fontSize: '15px', color: 'var(--fg-strong)' } }, p.name), h('span', { class: 'tag' }, sv ? sv.name : '?')),
      h('div', { class: 'cell-sub', style: { maxWidth: '520px' } }, p.source.kind === 'git' ? p.source.repo + ' @ ' + p.source.branch : p.source.path)),
    deployBtn);
  const result = h('div');
  body.append(h('h1', {}, 'Ship it'), h('p', { class: 'lead' }, 'Run the first deployment and watch every stage live. This is exactly what happens on each git push later.'),
    h('div', { class: 'mt16' }, hero), area, result);
  const showWin = (b) => {
    celebrate();
    const firstGit = p.source.kind === 'git' ? p : (summary.projects.find((x) => x.kind === 'git') || null);
    clear(result).append(h('div', { class: 'win mt16' },
      h('div', { class: 'big' }, icon('check')), h('h1', { style: { justifyContent: 'center' } }, 'Your stack is live'),
      h('p', { class: 'lead', style: { margin: '6px auto 0' } }, p.name + ' deployed in ' + dur(buildDur(b) || 0) + ' on ' + (b.server || 'your server') + '. Nicely done.'),
      h('div', { class: 'next' },
        [[firstGit ? '#/projects/' + firstGit.id + '/webhook' : '#/projects/new', 'webhook', 'Deploy on every push', 'Add the webhook to GitHub or GitLab.'],
          ['#/users', 'users', 'Invite your team', 'Operators can deploy; viewers can watch.'],
          ['#/servers', 'server', 'Add another server', 'Run stacks anywhere with the agent.'],
          ['#/account', 'key', 'Automate with the API', 'Create a token for CI pipelines.']]
          .map(([href, ic, t, d]) => h('a', { href, onclick: () => setDismissed(true) }, h('span', { class: 'ic' }, icon(ic)), h('div', {}, h('b', {}, t), h('span', {}, d))))),
      h('div', { class: 'ob-actions', style: { justifyContent: 'center', gap: '10px' } },
        h('a', { class: 'btn lg', href: '#/projects/' + p.id, onclick: () => setDismissed(true) }, 'Open project'),
        h('button', { class: 'btn primary lg', type: 'button', onclick: () => finish('#/') }, 'Go to dashboard', icon('right')))));
    deployBtn.replaceWith(h('span', { class: 'badge ok' }, icon('check'), 'deployed'));
  };
  const showFail = (b) => {
    clear(result).append(h('div', { class: 'callout bad mt16' }, icon('alert'), h('div', { class: 'grow' }, h('b', {}, 'The deploy failed. '), b.reason || 'See the log for details.',
      h('div', { class: 'flex mt8 wrap' }, h('a', { class: 'btn sm', href: '#/builds/' + b.id }, icon('terminal'), 'Open full log'),
        h('a', { class: 'btn sm', href: '#/projects/' + p.id + '/edit' }, icon('edit'), 'Edit project'),
        h('button', { class: 'btn sm primary', type: 'button', onclick: () => run() }, icon('refresh'), 'Try again'),
        h('button', { class: 'btn sm ghost', type: 'button', onclick: () => finish('#/') }, 'Finish later')))));
    deployBtn.disabled = false;
  };
  const run = async () => {
    deployBtn.disabled = true; clear(result);
    let b;
    try { b = await api('POST', '/api/projects/' + p.id + '/builds', { action: 'deploy' }); } catch (e) { fail(e); deployBtn.disabled = false; return; }
    const stages = h('div', { class: 'mini-stages' });
    const con = h('div', { class: 'console mini' });
    const status = h('span', { class: 'm faint' }, 'queued…');
    clear(area).append(h('div', { class: 'mt16' }, h('div', { class: 'flex between' }, h('div', { class: 'ob-h', style: { margin: 0 } }, 'Live deployment #' + b.number), status), stages,
      h('div', { class: 'console-wrap' }, con)));
    let n = 0, buf = '';
    const add = (raw) => {
      const { ts, tx } = parseLine(raw);
      if (tx.startsWith('##[stage] ')) return;
      n++;
      con.append(h('div', { class: 'ln ' + lineClass(tx) }, h('span', { class: 'no' }, n), h('span', { class: 'tx' }, tx || ' ')));
      con.scrollTop = con.scrollHeight;
    };
    const drawStages = (nb) => {
      clear(stages).append((nb.stages || []).map((s) => h('div', { class: 'stage ' + s.status },
        h('span', { class: 'n' }, icon(s.status === 'success' ? 'check' : s.status === 'running' ? 'loader' : 'x'), STAGE_LABEL[s.name] || s.name),
        h('span', { class: 'd' }, s.status === 'running' ? '…' : dur(stageDur(s))))));
      status.textContent = nb.status === 'running' ? 'running · ' + dur(buildDur(nb) || 0) : nb.status;
    };
    const es = new EventSource('/api/builds/' + b.id + '/stream');
    onCleanup(() => es.close());
    es.addEventListener('log', (e) => { buf += JSON.parse(e.data); const parts = buf.split('\n'); buf = parts.pop(); parts.forEach(add); });
    es.addEventListener('build', (e) => drawStages(JSON.parse(e.data)));
    es.addEventListener('end', async () => {
      es.close(); if (buf) add(buf);
      const nb = await api('GET', '/api/builds/' + b.id).catch(() => null);
      if (!nb) return;
      drawStages(nb);
      if (nb.status === 'success') showWin(nb); else showFail(nb);
    });
  };
  deployBtn.addEventListener('click', run);
  body.append(nav('project', h('span')));
}

// =====================================================================
//  router
// =====================================================================
let leaveGuard = null;
let lastHash = location.hash;
function go(hash) { if (location.hash === hash) route(); else location.hash = hash; }
async function route() {
  closeMenu();
  document.querySelectorAll('.overlay').forEach((o) => o.remove());
  runCleanups();
  if (!boot.user) return renderAuth();
  lastHash = location.hash;
  const parts = (location.hash.replace(/^#\/?/, '').split('?')[0] || '').split('/').filter(Boolean);
  window.scrollTo(0, 0);
  try {
    const [a, b, c] = parts;
    if (!a) await pageDashboard();
    else if (a === 'projects' && !b) await pageProjects();
    else if (a === 'projects' && b === 'new') await pageEdit(null);
    else if (a === 'projects' && c === 'edit') await pageEdit(b);
    else if (a === 'projects') await pageProject(b, c);
    else if (a === 'builds' && b) await pageBuild(b);
    else if (a === 'builds') await pageBuilds();
    else if (a === 'servers') await pageServers();
    else if (a === 'credentials') await pageCredentials();
    else if (a === 'users') await pageUsers();
    else if (a === 'settings') await pageSettings();
    else if (a === 'audit') await pageAudit();
    else if (a === 'account') await pageAccount();
    else if (a === 'welcome') await pageWelcome(b);
    else { shell('').append(emptyState('alert', 'Page not found', 'That page does not exist.', h('a', { class: 'btn', href: '#/' }, 'Back to dashboard'))); }
  } catch (e) { if (e && e.status !== 401) { console.error(e); fail(e); } }
}
window.addEventListener('hashchange', (e) => {
  if (leaveGuard) {
    const msg = leaveGuard();
    if (msg && !window.confirm(msg)) { history.replaceState(null, '', lastHash || '#/'); return; }
  }
  route();
});
window.addEventListener('beforeunload', (e) => { if (leaveGuard && leaveGuard()) { e.preventDefault(); e.returnValue = ''; } });

async function start() {
  try { boot = await api('GET', '/api/boot'); } catch (e) {
    clear($('#app')).append(h('div', { class: 'auth' }, h('div', { class: 'auth-card' }, h('div', { class: 'card' }, h('h3', {}, 'Cannot reach Dockup'), h('div', { class: 'sub' }, e.message),
      h('button', { class: 'btn primary', type: 'button', onclick: start }, 'Retry')))));
    return;
  }
  if (!boot.user) return renderAuth();
  await refreshSummary();
  route();
}
start();
})();

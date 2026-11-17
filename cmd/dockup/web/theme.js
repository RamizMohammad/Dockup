// Applied before first paint to avoid a theme flash. CSP forbids inline scripts.
(function () {
  var t = null;
  try { t = localStorage.getItem('dockup-theme'); } catch (e) { /* storage unavailable */ }
  if (t !== 'light' && t !== 'dark') {
    t = window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
  }
  document.documentElement.setAttribute('data-theme', t);
})();

(function () {
  'use strict';
  var ORDER = ['system', 'light', 'dark'];

  function readCookie() {
    var m = document.cookie.match(/(^| )fm_theme=([^;]+)/);
    return ORDER.indexOf(m ? m[2] : '') !== -1 ? m[2] : 'system';
  }

  function resolve(pref) {
    if (pref !== 'system') return pref;
    return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  // Loaded at bottom of <body>; document.body is available when this executes.
  document.body.addEventListener('theme-changed', function () {
    try {
      var pref = readCookie();
      var resolved = resolve(pref);
      var next = ORDER[(ORDER.indexOf(pref) + 1) % ORDER.length];
      document.documentElement.setAttribute('data-theme', pref);
      var btn = document.querySelector('[data-theme-toggle]');
      if (!btn) return;
      btn.setAttribute('aria-label', 'Theme: ' + pref + '; activate to cycle (next: ' + next + ')');
      btn.dataset.themeResolved = resolved;
      btn.setAttribute('hx-vals', JSON.stringify({ value: next }));
    } catch (e) {}
  });
})();

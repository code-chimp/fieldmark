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

  function apply(pref) {
    var resolved = resolve(pref);
    document.documentElement.setAttribute('data-theme', pref);
    document.documentElement.classList.toggle('dark', resolved === 'dark');
    document.querySelectorAll('[data-theme-choice]').forEach(function (btn) {
      btn.setAttribute('aria-pressed', btn.dataset.themeChoice === pref ? 'true' : 'false');
    });
  }

  document.body.addEventListener('theme-changed', function () {
    try { apply(readCookie()); } catch (e) {}
  });
})();

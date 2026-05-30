(function () {
  'use strict';
  function focusTab(tabs, i) {
    tabs.forEach(function (t, j) { t.setAttribute('tabindex', j === i ? '0' : '-1'); });
    tabs[i].focus();
  }
  function attach(strip) {
    if (strip._tabstripBound) return;
    strip._tabstripBound = true;
    var tabs = Array.prototype.slice.call(strip.querySelectorAll('button[role="tab"]'));
    strip.addEventListener('keydown', function (e) {
      var i = tabs.indexOf(document.activeElement);
      if (i < 0) return;
      if (e.key === 'ArrowLeft') { focusTab(tabs, (i - 1 + tabs.length) % tabs.length); e.preventDefault(); }
      else if (e.key === 'ArrowRight') { focusTab(tabs, (i + 1) % tabs.length); e.preventDefault(); }
      else if (e.key === 'Home') { focusTab(tabs, 0); e.preventDefault(); }
      else if (e.key === 'End') { focusTab(tabs, tabs.length - 1); e.preventDefault(); }
      else if (e.key === 'Enter' || e.key === ' ') { tabs[i].click(); e.preventDefault(); }
    });
  }
  function scan() { Array.prototype.forEach.call(document.querySelectorAll('nav[data-tabstrip]'), attach); }
  scan();
  document.addEventListener('htmx:after:swap', scan);
})();

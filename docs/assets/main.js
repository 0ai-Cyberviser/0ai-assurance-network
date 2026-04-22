/* ---- Smooth scroll highlight ---- */
(function () {
  'use strict';

  /* Active nav link on scroll */
  const sections = document.querySelectorAll('section[id]');
  const navLinks = document.querySelectorAll('.nav-links a[href^="#"]');

  function onScroll() {
    let current = '';
    sections.forEach(function (sec) {
      const top = sec.getBoundingClientRect().top;
      if (top <= 100) current = sec.id;
    });
    navLinks.forEach(function (a) {
      a.classList.toggle('active', a.getAttribute('href') === '#' + current);
    });
  }

  window.addEventListener('scroll', onScroll, { passive: true });

  /* Copy-to-clipboard for code blocks */
  document.querySelectorAll('.code-block').forEach(function (block) {
    var btn = document.createElement('button');
    btn.textContent = 'Copy';
    btn.className = 'copy-btn';
    btn.style.cssText = [
      'position:absolute', 'top:10px', 'right:10px',
      'background:rgba(79,142,247,.15)', 'color:#4f8ef7',
      'border:1px solid rgba(79,142,247,.3)', 'border-radius:6px',
      'padding:4px 10px', 'font-size:11px', 'cursor:pointer',
      'font-family:inherit', 'transition:opacity .2s'
    ].join(';');

    block.style.position = 'relative';

    btn.addEventListener('click', function () {
      var text = block.innerText.replace(/^Copy\n?/, '').replace(/\nCopied!?$/, '');
      navigator.clipboard.writeText(text).then(function () {
        btn.textContent = 'Copied!';
        setTimeout(function () { btn.textContent = 'Copy'; }, 1800);
      });
    });

    block.appendChild(btn);
  });

  /* Animate numbers in hero-meta on first view */
  var observed = false;
  var observer = new IntersectionObserver(function (entries) {
    if (observed) return;
    entries.forEach(function (e) {
      if (!e.isIntersecting) return;
      observed = true;
      document.querySelectorAll('[data-count]').forEach(function (el) {
        var target = parseInt(el.getAttribute('data-count'), 10);
        var suffix = el.getAttribute('data-suffix') || '';
        var start = 0;
        var duration = 900;
        var startTime = null;
        function step(ts) {
          if (!startTime) startTime = ts;
          var progress = Math.min((ts - startTime) / duration, 1);
          var val = Math.round(progress * target);
          el.textContent = val + suffix;
          if (progress < 1) requestAnimationFrame(step);
        }
        requestAnimationFrame(step);
      });
    });
  }, { threshold: 0.3 });

  var heroMeta = document.querySelector('.hero-meta');
  if (heroMeta) observer.observe(heroMeta);
})();

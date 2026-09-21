// app.js — authenticated pages: mobile sidebar, dropdowns, password
// visibility, confirm dialogs, reveal on scroll and stat count-ups.
(function () {
	var reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
	var shell = document.getElementById('shell');

	function closeSidebar() {
		if (shell) shell.classList.remove('sidebar-open');
		var open = document.querySelector('[data-sidebar-open]');
		if (open) open.setAttribute('aria-expanded', 'false');
	}

	// --- Mobile sidebar ---
	document.addEventListener('click', function (e) {
		if (e.target.closest('[data-sidebar-open]')) {
			if (shell) shell.classList.add('sidebar-open');
			var btn = e.target.closest('[data-sidebar-open]');
			btn.setAttribute('aria-expanded', 'true');
			return;
		}
		if (e.target.closest('[data-sidebar-close]')) {
			closeSidebar();
		}
	});

	// --- Dropdowns (profile menu, row actions) ---
	document.addEventListener('click', function (e) {
		var toggle = e.target.closest('[data-dropdown-toggle]');
		document.querySelectorAll('[data-dropdown].open').forEach(function (dd) {
			if (!toggle || !dd.contains(toggle)) {
				dd.classList.remove('open');
				var t = dd.querySelector('[data-dropdown-toggle]');
				if (t) t.setAttribute('aria-expanded', 'false');
			}
		});
		if (toggle) {
			var wrap = toggle.closest('[data-dropdown]');
			if (wrap) {
				var isOpen = wrap.classList.toggle('open');
				toggle.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
			}
		}
	});

	document.addEventListener('keydown', function (e) {
		if (e.key === 'Escape') {
			document.querySelectorAll('[data-dropdown].open').forEach(function (dd) {
				dd.classList.remove('open');
			});
			closeSidebar();
		}
	});

	// --- Password visibility toggle ---
	document.addEventListener('click', function (e) {
		var btn = e.target.closest('[data-password-toggle]');
		if (!btn) return;
		var wrap = btn.closest('.input-wrap');
		var input = wrap ? wrap.querySelector('[data-password-input]') : null;
		if (!input) return;
		var show = input.type === 'password';
		input.type = show ? 'text' : 'password';
		btn.classList.toggle('is-visible', show);
		btn.setAttribute('aria-label', show ? 'Hide password' : 'Show password');
	});

	// --- Confirm dialogs (capture phase so it runs before navigation) ---
	document.addEventListener('submit', function (e) {
		var form = e.target.closest('form[data-confirm]');
		if (form && !window.confirm(form.getAttribute('data-confirm'))) {
			e.preventDefault();
		}
	}, true);

	// --- Reveal on scroll ---
	var revealEls = document.querySelectorAll('[data-reveal]');
	if (!reduce && 'IntersectionObserver' in window) {
		var io = new IntersectionObserver(function (entries) {
			entries.forEach(function (entry) {
				if (entry.isIntersecting) {
					entry.target.classList.add('revealed');
					io.unobserve(entry.target);
				}
			});
		}, { threshold: 0.12, rootMargin: '0px 0px -32px 0px' });
		revealEls.forEach(function (el) { io.observe(el); });
	} else {
		revealEls.forEach(function (el) { el.classList.add('revealed'); });
	}

	// --- Count-up numbers on stat cards ---
	function countUp(el) {
		var to = parseInt(el.getAttribute('data-count-to'), 10);
		if (isNaN(to)) return;
		var start = null;
		function step(ts) {
			if (start === null) start = ts;
			var p = Math.min((ts - start) / 900, 1);
			el.textContent = Math.round(to * (1 - Math.pow(1 - p, 3))).toString();
			if (p < 1) window.requestAnimationFrame(step);
		}
		window.requestAnimationFrame(step);
	}

	var counters = document.querySelectorAll('[data-count-to]');
	counters.forEach(function (el) {
		if (!el.getAttribute('data-count-to')) return;
	});
	if (!reduce && 'IntersectionObserver' in window) {
		var cio = new IntersectionObserver(function (entries) {
			entries.forEach(function (entry) {
				if (entry.isIntersecting) {
					countUp(entry.target);
					cio.unobserve(entry.target);
				}
			});
		}, { threshold: 0.4 });
		counters.forEach(function (el) {
			if (el.getAttribute('data-count-to')) cio.observe(el);
		});
	}
})();

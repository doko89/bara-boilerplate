// motion.js — landing page motion: reveal on scroll (IntersectionObserver),
// cursor spotlight (--mx/--my CSS variables), 3D tilt and magnetic buttons.
// Everything honors prefers-reduced-motion and pointer:fine.
(function () {
	var reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
	var fine = window.matchMedia('(pointer: fine)').matches;

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

	// --- Spotlight: track the pointer as CSS variables ---
	function bindSpotlight(el) {
		el.addEventListener('pointermove', function (e) {
			var r = el.getBoundingClientRect();
			el.style.setProperty('--mx', (e.clientX - r.left) + 'px');
			el.style.setProperty('--my', (e.clientY - r.top) + 'px');
		});
	}

	document.querySelectorAll('[data-spotlight]').forEach(bindSpotlight);
	document.querySelectorAll('[data-tilt]').forEach(bindSpotlight);

	if (reduce || !fine) return;

	// --- 3D tilt on glass cards ---
	document.querySelectorAll('[data-tilt]').forEach(function (card) {
		var raf = null;
		card.addEventListener('pointermove', function (e) {
			if (raf) return;
			raf = window.requestAnimationFrame(function () {
				raf = null;
				var r = card.getBoundingClientRect();
				var px = (e.clientX - r.left) / r.width - 0.5;
				var py = (e.clientY - r.top) / r.height - 0.5;
				card.style.transform =
					'perspective(900px) rotateX(' + (-py * 5).toFixed(2) +
					'deg) rotateY(' + (px * 5).toFixed(2) + 'deg)';
			});
		});
		card.addEventListener('pointerleave', function () {
			card.style.transform = '';
		});
	});

	// --- Magnetic buttons ---
	document.querySelectorAll('[data-magnetic]').forEach(function (btn) {
		btn.addEventListener('pointermove', function (e) {
			var r = btn.getBoundingClientRect();
			var x = (e.clientX - r.left - r.width / 2) * 0.22;
			var y = (e.clientY - r.top - r.height / 2) * 0.32;
			btn.style.transform = 'translate(' + x.toFixed(1) + 'px,' + y.toFixed(1) + 'px)';
		});
		btn.addEventListener('pointerleave', function () {
			btn.style.transform = '';
		});
	});
})();

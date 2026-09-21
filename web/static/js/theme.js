// theme.js — runs on every page. The server already rendered the theme from
// the bara_theme cookie, so there is no flash of the wrong theme; this script
// only toggles. It also marks <html class="js"> so reveal-on-scroll styles
// activate progressively (without JS everything stays visible).
(function () {
	var root = document.documentElement;
	root.classList.add('js');

	function systemTheme() {
		return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
	}

	function currentTheme() {
		var t = root.getAttribute('data-theme');
		return t === 'light' || t === 'dark' ? t : systemTheme();
	}

	function apply(theme) {
		root.setAttribute('data-theme', theme);
		document.cookie =
			'bara_theme=' + theme + ';path=/;max-age=31536000;samesite=lax';
	}

	document.addEventListener('click', function (e) {
		var btn = e.target.closest('[data-theme-toggle]');
		if (!btn) return;
		apply(currentTheme() === 'dark' ? 'light' : 'dark');
	});
})();

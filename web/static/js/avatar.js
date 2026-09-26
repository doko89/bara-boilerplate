// avatar.js — settings page: drag-and-drop avatar upload. The chosen image
// is cropped square, resized to 512px and converted to WebP at 80% quality
// with the Canvas API, then posted to /settings/avatar. Only runs when a
// [data-avatar-dropzone] is on the page.
(function () {
	var MAX_SIDE = 512;
	var QUALITY = 0.8;

	var zone = document.querySelector('[data-avatar-dropzone]');
	if (!zone) return;

	var input = zone.querySelector('[data-avatar-input]');
	var status = zone.querySelector('[data-avatar-status]');
	var csrfInput = zone.querySelector('[data-avatar-csrf]');
	var maxBytes = parseInt(zone.getAttribute('data-max-bytes') || '0', 10);
	var busy = false;

	function say(msg, isError) {
		if (!status) return;
		status.textContent = msg;
		status.classList.toggle('avatar-status-error', !!isError);
	}

	function setSlots(url) {
		document.querySelectorAll('[data-avatar-slot]').forEach(function (slot) {
			while (slot.firstChild) slot.removeChild(slot.firstChild);
			var img = document.createElement('img');
			img.className = 'avatar-img';
			img.src = url;
			img.alt = '';
			slot.appendChild(img);
		});
	}

	function loadBitmap(file) {
		if ('createImageBitmap' in window) return window.createImageBitmap(file);
		return new Promise(function (resolve, reject) {
			var url = URL.createObjectURL(file);
			var img = new Image();
			img.onload = function () {
				URL.revokeObjectURL(url);
				resolve(img);
			};
			img.onerror = function () {
				URL.revokeObjectURL(url);
				reject(new Error('decode'));
			};
			img.src = url;
		});
	}

	function convert(file) {
		return loadBitmap(file).then(function (bmp) {
			var w = bmp.width || bmp.naturalWidth;
			var h = bmp.height || bmp.naturalHeight;
			if (!w || !h) throw new Error('decode');
			var side = Math.min(w, h);
			var target = Math.min(side, MAX_SIDE);
			var canvas = document.createElement('canvas');
			canvas.width = target;
			canvas.height = target;
			var ctx = canvas.getContext('2d');
			ctx.drawImage(bmp, (w - side) / 2, (h - side) / 2, side, side, 0, 0, target, target);
			if (bmp.close) bmp.close();
			return new Promise(function (resolve, reject) {
				canvas.toBlob(function (blob) {
					if (blob) resolve(blob);
					else reject(new Error('encode'));
				}, 'image/webp', QUALITY);
			});
		});
	}

	function upload(blob) {
		var form = new FormData();
		form.append('avatar', blob, 'avatar.webp');
		return fetch('/settings/avatar', {
			method: 'POST',
			headers: {
				'Accept': 'application/json',
				'X-CSRF-Token': csrfInput ? csrfInput.value : ''
			},
			body: form
		}).then(function (res) {
			return res.json().then(function (body) {
				if (!res.ok) throw new Error((body && body.error) || 'Upload failed.');
				return body.avatar_url;
			});
		});
	}

	function handle(file) {
		if (busy) return;
		if (!file || file.type.indexOf('image/') !== 0) {
			say('Choose an image file.', true);
			return;
		}
		busy = true;
		say('Converting…');
		convert(file).then(function (blob) {
			if (maxBytes > 0 && blob.size > maxBytes) {
				throw new Error('That image is too large even after compression.');
			}
			setSlots(URL.createObjectURL(blob));
			say('Uploading…');
			return upload(blob);
		}).then(function (url) {
			setSlots(url);
			say('Profile photo updated.');
		}).catch(function (err) {
			say(err && err.message ? err.message : 'Upload failed.', true);
		}).then(function () {
			busy = false;
			if (input) input.value = '';
		});
	}

	zone.addEventListener('click', function (e) {
		if (e.target.closest('[data-avatar-input]')) return;
		if (input) input.click();
	});
	zone.addEventListener('keydown', function (e) {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			if (input) input.click();
		}
	});
	if (input) {
		input.addEventListener('change', function () {
			if (input.files && input.files[0]) handle(input.files[0]);
		});
	}
	['dragenter', 'dragover'].forEach(function (name) {
		zone.addEventListener(name, function (e) {
			e.preventDefault();
			zone.classList.add('is-dragover');
		});
	});
	['dragleave', 'drop'].forEach(function (name) {
		zone.addEventListener(name, function (e) {
			e.preventDefault();
			if (name === 'drop' || e.target === zone) zone.classList.remove('is-dragover');
		});
	});
	zone.addEventListener('drop', function (e) {
		var files = e.dataTransfer ? e.dataTransfer.files : null;
		if (files && files[0]) handle(files[0]);
	});
})();

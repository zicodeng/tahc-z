$(document).ready(function() {
	var $form = $('#request-form');
	var $urlInput = $('#url-input');
	var $pageSum = $('#page-summary');
	var $error = $('#error');

	$form.submit(function(e) {
		e.preventDefault();

		// Remove previously displayed page summary if any.
		$pageSum.html('');
		$error.html('');

		var url = $urlInput.val();
		url = encodeURIComponent(url);

		$.ajax({
			url: 'v1/summary?q=' + url
		})
			.done(function(data) {
				var summary = (icon = images = videos = author = title = desc = type = url = '');
				for (prop in data) {
					switch (prop) {
						case 'icon':
							icon = '<img src="' + data[prop]['url'] + '" >';
							break;

						case 'images':
							data[prop].forEach(function(img) {
								if (img['url']) {
									images += '<img src="' + img['url'] + '"';
									if (img['width'] && img['height']) {
										images +=
											'width="' +
											img['width'] +
											'" height="' +
											img['height'] +
											'"';
									}
									images += '>';
								}
							});
							break;

						case 'videos':
							data[prop].forEach(function(video) {
								if (video['type'] && video['url']) {
									if (video['type'] === 'text/html') {
										videos += '<iframe src="' + video['url'] + '"';
									} else if (video['type'].startsWith('video')) {
										videos +=
											'<video src="' +
											video['url'] +
											'"' +
											'type="' +
											video['type'] +
											'" controls ';
									}

									// Add width and height.
									if (video['width'] && video['height']) {
										videos +=
											'width="' +
											video['width'] +
											'" height="' +
											video['height'] +
											'"';
									}

									if (video['type'] === 'text/html') {
										videos += '"></iframe>';
									} else if (video['type'].startsWith('video')) {
										videos += '"></video>';
									}
								}
							});
							break;

						case 'author':
							author = '<p>' + data[prop] + '</p>';
							break;

						case 'title':
							title = '<h3>' + data[prop] + '</h3>';
							break;

						case 'description':
							desc = '<p>' + data[prop] + '</p>';
							break;

						case 'type':
							type = '<span>' + data[prop] + ': </span>';
							break;

						case 'url':
							url = '<a href="' + data[prop] + '">' + data[prop] + '</a>';
							break;

						default:
							break;
					}
				}
				// Preserve order.
				summary = icon + images + videos + title + author + desc + type + url;

				// Display on page.
				$pageSum.html(summary);
				$pageSum.css('display', 'block');
			})
			.fail(function(error) {
				$error.text(error.responseText);
				$error.css('display', 'block');
			});
	});
});

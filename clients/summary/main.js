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
				var summary = (images = title = desc = type = url = '');
				for (prop in data) {
					switch (prop) {
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
				summary = images + title + desc + type + url;

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

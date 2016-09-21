// TODO: Lets wait for components to be ready
document.addEventListener('WebComponentsReady', function() {
	if ('addEventListener' in document) {
		document.addEventListener('DOMContentLoaded', function() {
			FastClick.attach(document.body);
		}, false);
	}
	var scope = document.getElementById('app-tmpl');
	console.warn('WebComponentsReady ready');
});

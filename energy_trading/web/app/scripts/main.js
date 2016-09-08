// TODO: Lets wait for components to be ready
document.addEventListener('WebComponentsReady', function() {
	if ('addEventListener' in document) {
		document.addEventListener('DOMContentLoaded', function() {
			FastClick.attach(document.body);
		}, false);
	}
	var scope = document.getElementById('app-tmpl');
	scope.config = {
		endpoint: '/api/v1',
		secureContext: null,
		chaincodeID: {
			report: '30268bf2818712b14161bd47db875bd5786b357641c2e09a218ff120dc2b072a15edc2e05a87bf5664debefab25880e91fa10ad0f62dde9ffb9ac47f91c8f73e',
			settle: '30268bf2818712b14161bd47db875bd5786b357641c2e09a218ff120dc2b072a15edc2e05a87bf5664debefab25880e91fa10ad0f62dde9ffb9ac47f91c8f73e'
		}
	};

	console.warn('WebComponentsReady ready');
});

document.addEventListener('WebComponentsReady', function() {
	var app = document.querySelector('#app');
	var socket = io.connect(window.location.href);
	socket.on('devices', function(data) {
		console.log('Updating data...', data);
		app.set('devices', data);
	});
});

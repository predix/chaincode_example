'use strict';
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

const express = require('express');
const serveStatic = require('serve-static');
const path = require('path');
const fs = require('fs');
const log = require('npmlog');
const pkg = require(path.resolve(__dirname, './package.json'));
const config = pkg.config;
const PORT = process.env.PORT || process.env.npm_package_config_port;

const app = express();

const BLOCKCHAIN_REFRESH_INTERVAL = process.env.BLOCKCHAIN_REFRESH_INTERVAL || 5000;



var program = {
	app: app,
	log: log,
	config: {
		path: './config.json',
		defaults: config,
		get: function(name) {
			return config[name] || null;
		}
	}
};

program.config.get('dirs').forEach(function(dir) {
	program.log.info('staticDir', dir);
	app.use(serveStatic(path.resolve(__dirname, dir), {index: ['index.html']}));
});

program.config.get('routes').forEach(function(route) {
	program.log.info('mounting', route);
	require(path.resolve(__dirname, route))(program, app);
});



app.use(require('morgan')('dev'));
app.set('port', process.env.PORT || process.env.npm_package_config_port || 9003);


const server = require('http').createServer(app);
const io = require('socket.io')(server);
server.listen(app.get('port'), function(){
  console.log('Express server listening on port ' + app.get('port'));
});


/*
app.listen(app.get('port'), function() {
	program.log.info('Open your browser to http://localhost:' + app.get('port'));
});*/



const remoteAttestation = require('./remote-attestation');

const pollDevices = (socket) =>{
	let refreshDevices = setInterval(() =>{
		log.info('refreshDevices');
		remoteAttestation.getDevices().then((resp) =>{
			socket.emit('devices', resp);
		});
	}, BLOCKCHAIN_REFRESH_INTERVAL);

	socket.on('disconnect', () =>{
		log.info('disconnect', 'socket disconnected clearing interval');
		clearInterval(refreshDevices);
	});
};


io.on('connection', (socket) => {
	pollDevices(socket);
	remoteAttestation.getDevices().then((resp) =>{
		socket.emit('devices', resp);
	});

});

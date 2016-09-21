'use strict';

const log = require('debug')('chaincode_example:socket');

/**
 * Socket.js is a web socket server that pushes updates to the client when the block height changes.
 */
module.exports = (program, app) => {
	log('socket mounted');

	//var app = require('express').createServer();
	const io = require('socket.io')(app);

	io.on('connection', function (socket) {
		socket.emit('news', {
			hello: 'world'
		});
		socket.on('my other event', function (data) {
			console.log(data);
		});
	});
};

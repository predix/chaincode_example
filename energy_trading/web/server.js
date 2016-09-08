'use strict';
// TODO: Simple http server to serve bower_components and www dir
var express = require('express'),
	serveStatic = require('serve-static'),
	log4js = require('log4js'),
	path = require('path'),
	fs = require('fs'),

	pkg = JSON.parse(fs.readFileSync(path.resolve(__dirname + path.sep + 'package.json'))),
	config = pkg.config,
	PORT = process.env.PORT || process.env.npm_package_config_port,
	app = express(),
	//server = livereload.createServer({port: config.livereload}),
	program = {};


var http = require('http');
var https = require('https');
var privateKey  = fs.readFileSync('sslcerts/key.pem', 'utf8');
var certificate = fs.readFileSync('sslcerts/cert.pem', 'utf8');

var credentials = {key: privateKey, cert: certificate};

//var httpServer = http.createServer(app);
//var httpsServer = https.createServer(credentials, app);


var log = require('debug')(`${pkg.name}:server`);

//ENVa
process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";
//process.env.JSON_PROXY_GATEWAY = process.env.http_proxy || process.env.https_proxy;

log('Setting up env');

program = {
	app: app,
	log: log4js.getLogger(pkg.name),
	config: {
		path: './config.json',
		defaults: config,
		get: function(name) {
			return this.defaults[name];
		}
	}
};

app.use(require('json-proxy').initialize(config));

program.config.get('dirs').forEach(function(dir) {
	app.use(express.static(path.resolve(__dirname, dir), {index: ['index.html']}));
	//server.watch(path.resolve(__dirname, dir));
});

//Dist dir
app.use('/bower_components', express.static(path.resolve(__dirname, './bower_components'), {index: ['index.html']}));
app.use('/dist', express.static(path.resolve(__dirname, './dist'), {index: ['index.html']}));

program.config.get('routes').forEach(function(route) {
	require(path.resolve(__dirname, route))(program, app);
});

app.listen(PORT, function() {
	program.log.debug('Open your browser to http://localhost:' + PORT);
});


//httpServer.listen(PORT);
//httpsServer.listen(8443);

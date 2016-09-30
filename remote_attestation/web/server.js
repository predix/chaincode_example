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

app.listen(PORT, function() {
	program.log.info('Open your browser to http://localhost:' + PORT);
});

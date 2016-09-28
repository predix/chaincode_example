'use strict';
const express = require('express');
const log = require('debug')('chaincode_example:api');
const request = require('request');
const bodyParser = require('body-parser');

module.exports = function (program, app) {
	var chaincodeID = process.env.BLOCKCHAIN_CHAINCODE_ID || '63d9383dd5b660303df7d1ff024b8b27b740c8608d83794469150892bdbd9b2e719109dab50f7fa13f75d401331cd408d7b02880871966194c1b7fc6072a0b03';
	var baseRequest = request.defaults({
		baseUrl: process.env.BLOCKCHAIN_ENDPOINT || 'https://blockchai-blockcha-1g7vft5n33q-1508885349.us-west-2.elb.amazonaws.com/',
		headers: {
			'Accept': 'application/json'
		}
	});

	var router = new express.Router();
	router.use(bodyParser.json());
	router.all('/*', function (req, res, next) {
		var options = {
			method: req.method,
			url: req.url.replace('api/v1/', ''),
			qs: req.query,
			json: req.body || null
		};

		if (req.method.toLowerCase() === 'post') {
			req.body.params.chaincodeID.name = chaincodeID;
			log(req.method, options.url, options.json.method, options.json.params.ctorMsg['function'], options.json.params.ctorMsg.args.toString());
		}

		options.json = req.body;

		baseRequest(options, function (err, resp, body) {
			if (!resp) {
				res.status(400).send({
					error: 'There was a problem with the request.',
					data: options
				});
			}
			log(resp.statusCode, options.url);
			if (resp.body && resp.body.error) {
				return res.status(400).send(resp.body.error);
			}
			if (err) {
				return res.status(resp.statusCode).send(resp.body);
			}
			return res.status(resp.statusCode).send(resp.body);
		});
	});

	app.use('/api/v1', router);
};

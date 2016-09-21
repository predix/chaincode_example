'use strict';
const express = require('express');
const log = require('debug')('chaincode_example:api');
const request = require('request');
const bodyParser = require('body-parser');

module.exports = function (program, app) {
	var chaincodeID = process.env.BLOCKCHAIN_CHAINCODE_ID || '30268bf2818712b14161bd47db875bd5786b357641c2e09a218ff120dc2b072a15edc2e05a87bf5664debefab25880e91fa10ad0f62dde9ffb9ac47f91c8f73e';
	var baseRequest = request.defaults({
		baseUrl: process.env.BLOCKCHAIN_ENDPOINT || 'https://blockchai-blockcha-sfvkghlrnmp2-1110560954.us-west-2.elb.amazonaws.com',
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

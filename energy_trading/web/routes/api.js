'use strict';
const express = require('express');
const log = require('debug')('poc-blockchain-ui:api');
const request = require('request');
const bodyParser = require('body-parser');

module.exports = function (program, app) {


	if(!process.env.BLOCKCHAIN_ENDPOINT){
		throw new Error('Must provide process.env.BLOCKCHAIN_ENDPOINT');
	}

	var baseRequest = request.defaults({
		baseUrl: process.env.BLOCKCHAIN_ENDPOINT,
		headers: {
			'Accept': 'application/json'
		}
	});

	function $http(options) {
		return new Promise(function (resolve, reject) {
			baseRequest(options, function (err, resp, body) {
				log('$http', options, resp, body);
				if (err) {
					reject(err);
				}
				try {
					resp.data = JSON.parse(resp.body);
				} catch (e) {
					console.error('Could not parse json', e);
				}
				resolve(resp);
			});
		});
	}


	var router = new express.Router();
	router.use(bodyParser.json());
	router.use(function (req, res, next) {
		log(req.method, req.url);
		next();
	});

	log('mounted');


	router.all('/api/v1/*', function (req, res, next) {
		var options = {
			method: req.method,
			url: req.url.replace('api/v1/', ''),
			qs: req.query,
			json: req.body || null
		};


		if(req.body && req.body.params && req.body.params.chaincodeID){
			if( process.env.CHAINCODE_ID){
				console.log('Replacing chaincodeID', process.env.CHAINCODE_ID);
				options.json.params.chaincodeID.name = process.env.CHAINCODE_ID;
			}

		}


		baseRequest(options, function (err, resp, body) {
			if(!resp){
				res.status(400).send({
					error_message: 'There was a problem.'
				});
			}
			log('response', resp.statusCode);
			if(err){
				res.status(resp.statusCode).send(body);
			}
			res.status(resp.statusCode).send(resp.body);
		});
	});



	app.use(router);
};

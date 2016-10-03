'use strict';
const request = require('request');
const moment = require('moment');
const log = require('npmlog');

process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

delete process.env.HTTP_PROXY;
delete process.env.HTTPS_PROXY;
delete process.env.http_proxy;
delete process.env.https_proxy;

const CHAINCODE_ID = process.env.CHAINCODE_ID || '48eedf6c2b7d5e83518795744ca9d6da9ddea3630599fc291974ce16de7309249df0f644469938a4fb747a4de2153d4db9c8878b72fd8f76f077d1e96b6380e3';
const BLOCKCHAIN_ENDPOINT = process.env.BLOCKCHAIN_ENDPOINT || 'https://blockchai-blockcha-sfvkghlrnmp2-1110560954.us-west-2.elb.amazonaws.com';

function sendBlockchainRequest(){
	return new Promise((resolve, reject) =>{
		var options = {
			method: 'POST',
		  baseUrl: BLOCKCHAIN_ENDPOINT,
			url: '/chaincode',
		  headers:{
		     'Content-Type': 'application/x-www-form-urlencoded'
			 },
		  json: {
				"id": 0,
			  "jsonrpc": "2.0",
			  "method": "query",
			  "params": {
			    "type": 1,
			    "chaincodeID": {
			      "name": CHAINCODE_ID
			    },
			    "ctorMsg": {
			      "function": "attestationRecords",
			      "args": [
			      ]
			    }
			  }
			}
		};
		request(options, (error, response, body) => {
		  if(error){
				reject(error);
			}
			try {
				body.data = JSON.parse(body.result.message);
				resolve(body.data);
			} catch (e) {
				log.error('parse', e);
				reject(e);
			}
		});
	});
}

function parseDates(resp){
	if(resp && resp.length){
		resp.forEach((device) =>{
			device.datetime = moment(new Date(device.time / 1000000)).format('llll');
		});
	}
	return resp;
}

module.exports = {
	getDevices: () =>{
		return sendBlockchainRequest().then(parseDates).then((resp) =>{
			return resp;
		});
	}
};

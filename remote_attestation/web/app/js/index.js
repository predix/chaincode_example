document.addEventListener('WebComponentsReady', function() {

	Polymer({
		is: 'x-app',
		properties: {
			listData: {
				type: Array
			}
		},
		ready: function() {
			console.warn(this.tagName, ' created!');
		},
		attached: function() {
			console.warn(this.tagName, ' attached!');
		},
		detached: function() {
			console.warn(this.tagName, ' detached!');
		},
		formatDate: function(d){
			var date = moment(d / 1000000).format('llll');
			console.log('formatDate', d, date);
			return date;
		}
	});

/**
html2canvas(document.body, {
  onrendered: function(canvas) {
    document.body.appendChild(canvas);
  }
});*/
	var app = document.querySelector('#app');
var data = {
    "jsonrpc": "2.0",
    "result": {
        "status": "OK",
        "message": "[{\"device_id\":\"client-1\",\"attestation_serve_id\":\"server-1\",\"status\":2,\"validation_hash\":\"abc13f5ed82b868699041677f80a5e879af2fc78\",\"time\":1475083802894940886},{\"device_id\":\"device1\",\"attestation_serve_id\":\"server1\",\"status\":1,\"validation_hash\":\"0osdfgs9uodmvnp4wcw7\",\"time\":1475012947665151410},{\"device_id\":\"device2\",\"attestation_serve_id\":\"server2\",\"status\":2,\"validation_hash\":\"0osdfgsrtty678mvnp4wcw7\",\"time\":1475084648581502932}]"
    },
    "id": 0
};

var devices = JSON.parse(data.result.message);
	app.data = devices;
	console.log(devices);
});
# Remote Attestation Example App
This repository contains an example application that talks to a blockchain backend.

## Get Started
First clone this repo and install the dependencies.

1. Clone the repo and go into web directory

```
$ git clone https://github.com/atulkc/chaincode_example.git && cd remote_attestation/web
```

2. Install dependencies (npm)

```
$ npm install
```

3. Start application locally

```
$ npm start
```

> Open web browser to http://localhost:9002

4. Deploy application to Cloud Foundry

```
$ npm run deploy
```


## Customizing

5. To change port, chaincodeID and blockchain endpoint, simply sent the following env vars.

```
$ export PORT=9049;
$ export BLOCKCHAIN_ENDPOINT=https://endpoint
```


## Usage
The web application demonstrates using the blockchain REST api to invoke and query nodes.

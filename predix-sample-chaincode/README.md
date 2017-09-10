# Sample Chaincode
This repo holds the sample chaincode that can be deployed to hyperledger blockchain network, which is the underlying blockchain implementation used in Predix Blockchain-as-a-service.

This chaincode shows how you can use zone attribute, which is set for each user to enforce zone based access control in chaincode itself.

Use this chaincode as a reference for enforcing zone based access control as well as hyperledger dependencies.

## How to use
1. Make sure GO environment is setup (`GOPATH`, `GOROOT`)
1. Go get this repo

	```
	go get github.devtools.predix.io/predix-security/sample-hyperledger-chaincode
	```
1. Make sure it compiles

	```
	cd $GOPATH/src/github.devtools.predix.io/predix-security/sample-hyperledger-chaincode
	go build example.go
	```
1. Make required changes to `example.go` as per requirement for your smart contract.
1. Make sure the changes compile
1. Create a tar.gz of the chaincode and dependencies and upload it to predix blockchain service

	```
	tar czvf mychaincode.tgz *
	```
	Use `PUT /v1/chaincodes/:chaincodename` endpoint of Predix blockchain service to deploy this chaincode to predix blockchain.

# Smart contract for device maintenance checklist
Smart contract to demonstrate multi party collaboration during device maintenance. It borrows concepts of access control from [here](https://github.com/hyperledger/fabric/tree/master/examples/chaincode/go/asset_management).

For a high value asset the maintainenance goes through multiple stages and is done by multiple service providers. This smart contract allows each service provider to record the completion of their respective maintenance part indicating if the maintenance was OK and asset is OK to be in service again. Owner of the contract then signs off on completion of maintenance and asset is ready to be in service again.

All the dependencies required to execute this smart contract are not currently captured. Because of that in order to execute this test copy this directory to `examples/chaincode/go` under `github.com/hyperledger/fabric` go repository and execute this from the vagrant vm as specified in the hyperledger documentation.

To deploy and execute this smart contract do following:

1. Build the client

	```
	cd client
	go build
	```
1. Update `core.yaml` to point to your blockchain deployment
1. Execute the binary
	```
	./client
	```
1. Program will output failure or success of the test

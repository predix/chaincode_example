# Smart contract for device maintenance checklist
Smart contract to demonstrate multi party collaboration during device maintenance. It borrows concepts of access control heavily from [here](https://github.com/hyperledger/fabric/tree/master/examples/chaincode/go/asset_management).

For a high value asset the maintainenance goes through multiple stages and is done by multiple service providers. This smart contract allows each service provider to record the completion of their respective maintenance part indicating if the maintenance was OK and asset is OK to be in service again. Owner of the contract then signs off on completion of maintenance and asset is ready to be in service again.

This is still Work in Progress and not completed yet

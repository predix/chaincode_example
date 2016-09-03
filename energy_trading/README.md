# Energy trading smart contract
Smart contract to demonstrate peer to peer energy trading. We are using [this](https://github.com/olegabu/decentralized-energy-utility) as our starting point. It was a project done for hackathon on hyperledger and we are changing it to demonstrate peer to peer energy trading. Following changes have been made:

1. Deployer of chaincode specifies exchange rate per kwh and specifies the commission to be charged per kwh
1. Exchange account (grid maintainer) is initialized at time of deploy
1. Provide method to register a new meter
1. Store the data on meter id, reported kwh and account balance in a table
1. Iterate over all the rows in table when doing settlement
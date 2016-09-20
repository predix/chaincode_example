# Remote attestation smart contract
Smart contract to record the remote attestation status of device. This smart contract provides a way for attestors/verifiers to record the status and validation hash on block chain to guarantee integrity and allow multiple attestors to record the status.

## Steps to deploy and use this smart contract
1. Deploy chaincode

    ```
    curl -k -XPOST -d @scripts/deploy_chaincode.txt https://<blockchain ip>/chaincode
    ```
1. Record attestation status

    ```
    curl -k -XPOST -d @scripts/record_attestation_status.txt https://<blockchain ip>/chaincode
    ```
1. Query all attestation records

    ```
    curl -k -XPOST -d @scripts/attestation_records.txt https://<blockchain ip>/chaincode
    ```
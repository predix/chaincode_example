Deploy chaincode

```
curl -X POST 10.244.13.2:5000/chaincode -d '
{
    "jsonrpc": "2.0",
    "method": "deploy",
    "params": {
        "type": 1,
        "chaincodeID":{
            "path":"https://github.com/vbanga/chaincode_example/infinite_chaincode"
        },
        "ctorMsg": {
            "function":"init",
            "args":["c", "500"]
        }
    },
    "id": 1
}'
```

Invoke the chaincode which triggers the infinite loop

```
curl -X POST 10.244.13.2:5000/chaincode -d '
{
    "jsonrpc": "2.0",
    "method": "invoke",
    "params": {
        "type": 1,
        "chaincodeID":{
            "name":"1f94386ed6053eea263921c9513f371606bbfdf31abeda0542994c220c0e24d5af80547b09c921c43d454ffd4e07f07f64515bb89c50de228bcab81b5e3b86c4"
        },
        "ctorMsg": {
            "function":"invoke",
            "args":["c", "100"]
        }
    },
    "id": 3
}'
```

CURL to invoke 'query' method on the chaincode for permissioned blockchain
```
curl -X POST 10.244.13.2:5000/chaincode -d '
{
    "jsonrpc": "2.0",
    "method": "query",
    "params": {
        "type": 1,
        "chaincodeID":{
            "name":"1f94386ed6053eea263921c9513f371606bbfdf31abeda0542994c220c0e24d5af80547b09c921c43d454ffd4e07f07f64515bb89c50de228bcab81b5e3b86c4"
        },
        "ctorMsg": {
            "function":"query",
            "args":["c"]
        }
    },
    "id": 4
}'
```

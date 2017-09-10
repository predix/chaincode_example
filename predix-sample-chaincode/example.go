package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ZoneBasedCounter is an example that use Attribute Based Access Control to control the access to a
// counter by users belonging to a specific zone. Zone is a concept in Predix that identifies tenancy
// within a service or application. Predix Blockchain service adds zone attribute to each user that
// it creates on hyperledger network. This attribute can be used by chaincode to enforce zone based
// access control.
type ZoneBasedCounter struct {
}

type fn func() ([]byte, error)

//Init the chaincode asigned the value "0" to the counter in the state.
func (t *ZoneBasedCounter) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	err := stub.PutState("zoneID", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	err = stub.PutState("initVal", []byte(args[1]))
	if err != nil {
		return nil, err
	}
	err = stub.PutState("counter", []byte("0"))
	return nil, err
}

//Invoke Transaction makes increment counter
func (t *ZoneBasedCounter) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "increment" {
		f := func() ([]byte, error) {
			counter, err := stub.GetState("counter")
			if err != nil {
				return nil, err
			}
			var cInt int
			cInt, err = strconv.Atoi(string(counter))
			if err != nil {
				return nil, err
			}
			cInt = cInt + 1
			counter = []byte(strconv.Itoa(cInt))
			err = stub.PutState("counter", counter)
			return nil, err
		}
		return t.verifyAndExecute(stub, f)
	} else if function == "reset" {
		f := func() ([]byte, error) {
			err := stub.PutState("counter", []byte("0"))
			return nil, err
		}
		return t.verifyAndExecute(stub, f)
	}

	return nil, errors.New("Invalid invoke function name. Expecting \"increment\" or \"reset\"")
}

func (t *ZoneBasedCounter) getAttributeFromCertificate(stub shim.ChaincodeStubInterface) (string, []byte, error) {
	attributeName := "zone"
	val, err := stub.ReadCertAttribute(attributeName)
	fmt.Printf("%s => %v error %v \n", attributeName, string(val), err)
	if err != nil {
		return "", nil, errors.New("Error reading cert attribute:" + err.Error())
	}
	return attributeName, val, nil
}

func (t *ZoneBasedCounter) verifyAndExecute(stub shim.ChaincodeStubInterface, f fn) ([]byte, error) {
	zoneID, err := stub.GetState("zoneID")
	if err != nil {
		return nil, err
	}
	// Here the ABAC API is called to verify the attribute, just if the value is verified the counter will be incremented.
	isOK, _ := stub.VerifyAttribute("zone", zoneID)
	if isOK {
		return f()
	}
	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *ZoneBasedCounter) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "read" {
		var err error
		// Get the state from the ledger
		Avalbytes, err := stub.GetState("counter")
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get state for counter\"}"
			return nil, errors.New(jsonResp)
		}

		if Avalbytes == nil {
			jsonResp := "{\"Error\":\"Nil amount for counter\"}"
			return nil, errors.New(jsonResp)
		}

		jsonResp := "{\"Name\":\"counter\",\"Amount\":\"" + string(Avalbytes) + "\"}"
		fmt.Printf("Query Response:%s\n", jsonResp)
		return Avalbytes, nil
	} else if function == "attributes" {
		_, val, err := t.getAttributeFromCertificate(stub)
		return val, err
	}
	return nil, errors.New("Invalid query function name. Expecting \"read\"")
}

func main() {
	err := shim.Start(new(ZoneBasedCounter))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

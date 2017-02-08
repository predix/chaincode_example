package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("remote_attestation")

const (
	tableName = "DeviceAttestation"
)

type DeviceAttestationInfo struct {
	DeviceId            string `json:"device_id"`
	AttestationServerId string `json:"attestation_serve_id"`
	Status              uint64 `json:"status"`
	ValidationHash      string `json:"validation_hash"`
	Time                int64  `json:"time"`
}

// RemoteDeviceAttestation implementation. This smart contract enables multiple attestors
// to perform remote attestation of device and verify that the device is running authentic
// and valid software/hardware configuration
type RemoteDeviceAttestation struct {
}

func (t *RemoteDeviceAttestation) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error

	if len(args) != 0 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. No arguments required for deploying this contract.")
	}

	_, err = stub.GetTable(tableName)
	if err == shim.ErrTableNotFound {
		err = stub.CreateTable(tableName, []*shim.ColumnDefinition{
			&shim.ColumnDefinition{Name: "DeviceId", Type: shim.ColumnDefinition_STRING, Key: true},
			&shim.ColumnDefinition{Name: "AttestationServerId", Type: shim.ColumnDefinition_STRING, Key: true},
			&shim.ColumnDefinition{Name: "Status", Type: shim.ColumnDefinition_UINT64, Key: false},
			&shim.ColumnDefinition{Name: "ValidationHash", Type: shim.ColumnDefinition_STRING, Key: false},
			&shim.ColumnDefinition{Name: "Time", Type: shim.ColumnDefinition_INT64, Key: false},
		})
		if err != nil {
			logger.Errorf("Error creating table:%s", err.Error())
			return nil, errors.New("Failed creating DeviceAttestation table.")
		}
	} else {
		logger.Info("Table already exists")
	}

	logger.Info("Successfully deployed chain code")

	return nil, nil
}

func (t *RemoteDeviceAttestation) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "deviceAttestationStatus" {
		return t.deviceAttestationStatus(stub, args)
	}

	logger.Errorf("Unimplemented method :%s called", function)

	return nil, errors.New("Unimplemented '" + function + "' invoked")
}

// Enrolls a new meter
func (t *RemoteDeviceAttestation) deviceAttestationStatus(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In deviceAttestationStatus function")
	if len(args) < 4 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. Specify device id, attestation server id, status and validation hash")
	}

	deviceId := args[0]
	serverId := args[1]
	statusStr := args[2]
	validationHash := args[3]
	status, err := strconv.ParseUint(string(statusStr), 10, 64)
	if err != nil {
		logger.Errorf("Error in converting to int:%s", err.Error())
		return nil, fmt.Errorf("Invalid value of status:%s", statusStr)
	}

	logger.Infof("Registering attestation status of device:%s, status:%d, server:%s and validation hash:%s", deviceId, status, serverId, validationHash)
	now := time.Now().UnixNano()
	ok, err := stub.InsertRow(tableName, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: deviceId}},
			&shim.Column{Value: &shim.Column_String_{String_: serverId}},
			&shim.Column{Value: &shim.Column_Uint64{Uint64: status}},
			&shim.Column{Value: &shim.Column_String_{String_: validationHash}},
			&shim.Column{Value: &shim.Column_Int64{Int64: now}},
		},
	})

	if err != nil {
		logger.Errorf("Error in registering device attestation status:%s", err)
		return nil, errors.New("Error in registering device attestation status")
	}

	if !ok {
		logger.Debugf("Row for device :%s and server: %s already exists, just updating the entry", deviceId, serverId)
		row, err := t.getRow(stub, deviceId, serverId)
		if err != nil {
			logger.Errorf("Failed retrieving record for device [%s] and server [%s] combination: [%s]", deviceId, serverId, err)
			return nil, fmt.Errorf("Failed retrieving record for device [%s] and server [%s] combination: [%s]", deviceId, serverId, err)
		}
		row.Columns[2] = &shim.Column{Value: &shim.Column_Uint64{Uint64: status}}
		row.Columns[3] = &shim.Column{Value: &shim.Column_String_{String_: validationHash}}
		row.Columns[4] = &shim.Column{Value: &shim.Column_Int64{Int64: now}}
		_, err = t.updateRow(stub, row)
		if err != nil {
			logger.Errorf("Error in updating attestation record for device:%s with status:%d and validation hash:%s by server %s", deviceId, status, validationHash, serverId)
			return nil, errors.New("Error in updating attestation record")
		}
	}
	logger.Infof("Registered attestation status:%d for device: %s", status, deviceId)

	return nil, nil
}

func (t *RemoteDeviceAttestation) getRow(stub shim.ChaincodeStubInterface, deviceId, serverId string) (shim.Row, error) {
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: deviceId}}
	col2 := shim.Column{Value: &shim.Column_String_{String_: serverId}}
	columns = append(columns, col1)
	columns = append(columns, col2)

	return stub.GetRow(tableName, columns)
}

func (t *RemoteDeviceAttestation) updateRow(stub shim.ChaincodeStubInterface, row shim.Row) (bool, error) {
	return stub.ReplaceRow(tableName, row)
}

// Query callback representing the query of a chaincode
func (t *RemoteDeviceAttestation) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "attestationRecords" {
		return t.attestationRecords(stub, args)
	}

	return nil, errors.New("Invalid query function name")
}

// Return all attestation records
func (t *RemoteDeviceAttestation) attestationRecords(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In attestationRecords function")
	if len(args) > 0 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. No arguments required")
	}

	var columns []shim.Column

	rowChannel, err := stub.GetRows(tableName, columns)
	if err != nil {
		logger.Errorf("Error in getting rows:%s", err.Error())
		return nil, errors.New("Error in fetching rows")
	}
	attestnRecords := make([]DeviceAttestationInfo, 0)
	for row := range rowChannel {
		attestnRecord := DeviceAttestationInfo{
			DeviceId:            row.Columns[0].GetString_(),
			AttestationServerId: row.Columns[1].GetString_(),
			Status:              row.Columns[2].GetUint64(),
			ValidationHash:      row.Columns[3].GetString_(),
			Time:                row.Columns[4].GetInt64(),
		}
		attestnRecords = append(attestnRecords, attestnRecord)
	}

	payload, err := json.Marshal(attestnRecords)
	if err != nil {
		logger.Errorf("Failed marshalling payload")
		return nil, fmt.Errorf("Failed marshalling payload [%s]", err)
	}

	return payload, nil
}

func main() {
	err := shim.Start(new(RemoteDeviceAttestation))
	if err != nil {
		fmt.Printf("Error starting Remote device attestation chaincode: %s", err)
	}
}

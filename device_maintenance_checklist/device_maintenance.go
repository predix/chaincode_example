package main

import (
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"

    "github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("device_maintenance")

const (
    deviceChecksOwnerMapTable = "DeviceChecksOwnerMap"
    deviceServiceTable        = "DeviceService"
)

type Device struct {
    Id string `json:"id"`
}

// DeviceMaintenance chaincode that provides a way to record maintenance checklist
// for high value assets being worked on by multiple parties
type DeviceMaintenanceChaincode struct {
}

func (t *DeviceMaintenanceChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    var err error

    if len(args) > 0 {
        logger.Error("Incorrect number of arguments")
        return nil, errors.New("Incorrect number of arguments. No arguments required.")
    }

    _, err = stub.GetTable(deviceChecksOwnerMapTable)
    if err == shim.ErrTableNotFound {
        err = stub.CreateTable(deviceChecksOwnerMapTable, []*shim.ColumnDefinition{
            &shim.ColumnDefinition{Name: "DeviceId", Type: shim.ColumnDefinition_STRING, Key: true},
            &shim.ColumnDefinition{Name: "PublicKey", Type: shim.ColumnDefinition_BYTES, Key: false},
            &shim.ColumnDefinition{Name: "Owner", Type: shim.ColumnDefinition_BYTES, Key: false},
            &shim.ColumnDefinition{Name: "Check1", Type: shim.ColumnDefinition_BYTES, Key: false},
            &shim.ColumnDefinition{Name: "Check2", Type: shim.ColumnDefinition_BYTES, Key: false},
            &shim.ColumnDefinition{Name: "Check3", Type: shim.ColumnDefinition_BYTES, Key: false},
        })
        if err != nil {
            logger.Errorf("Error creating table:%s - %s", deviceChecksOwnerMapTable, err.Error())
            return nil, errors.New("Failed creating DeviceChecksOwnerMap table.")
        }
    } else {
        logger.Info("Table already exists")
    }

    _, err = stub.GetTable(deviceServiceTable)
    if err == shim.ErrTableNotFound {
        err = stub.CreateTable(deviceServiceTable, []*shim.ColumnDefinition{
            &shim.ColumnDefinition{Name: "DeviceId", Type: shim.ColumnDefinition_STRING, Key: true},
            &shim.ColumnDefinition{Name: "ServiceId", Type: shim.ColumnDefinition_BYTES, Key: true},
            &shim.ColumnDefinition{Name: "Check1", Type: shim.ColumnDefinition_BOOL, Key: false},
            &shim.ColumnDefinition{Name: "Check2", Type: shim.ColumnDefinition_BOOL, Key: false},
            &shim.ColumnDefinition{Name: "Check3", Type: shim.ColumnDefinition_BOOL, Key: false},
            &shim.ColumnDefinition{Name: "SignOff", Type: shim.ColumnDefinition_BOOL, Key: false},
        })
        if err != nil {
            logger.Errorf("Error creating table:%s - %s", deviceServiceTable, err.Error())
            return nil, errors.New("Failed creating DeviceService table.")
        }
    } else {
        logger.Info("Table already exists")
    }

    // Set the admin
    // The metadata will contain the certificate of the administrator
    adminCert, err := stub.GetCallerMetadata()
    if err != nil {
        logger.Debug("Failed getting metadata")
        return nil, errors.New("Failed getting metadata.")
    }
    if len(adminCert) == 0 {
        logger.Debug("Invalid admin certificate. Empty.")
        return nil, errors.New("Invalid admin certificate. Empty.")
    }

    logger.Debug("The administrator is [%x]", adminCert)

    stub.PutState("admin", adminCert)

    logger.Info("Successfully deployed chain code")

    return nil, nil
}

func (t *DeviceMaintenanceChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

    if function == "enroll" {
        return t.enroll(stub, args)
    }

    if function == "delete" {
        return t.delete(stub, args)

    }

    logger.Errorf("Unimplemented method :%s called", function)

    return nil, errors.New("Unimplemented '" + function + "' invoked")
}

// Enrolls a new meter
func (t *DeviceMaintenanceChaincode) enroll(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    logger.Info("In enroll function")
    if len(args) < 6 {
        logger.Error("Incorrect number of arguments")
        return nil, errors.New("Incorrect number of arguments. Specify device id, public key, owner, owner for check1, check2 and check3.")
    }

    deviceId := args[0]
    devicePubKey, err := base64.StdEncoding.DecodeString(args[1])
    if err != nil {
        logger.Error("Failed decoding public key")
        return nil, errors.New("Failed decoding public key")
    }
    owner, err := base64.StdEncoding.DecodeString(args[2])
    if err != nil {
        logger.Error("Failed decoding owner certificate")
        return nil, errors.New("Failed decoding owner")
    }
    check1Owner, err := base64.StdEncoding.DecodeString(args[3])
    if err != nil {
        logger.Error("Failed decoding check1 owner certificate")
        return nil, errors.New("Failed decoding check1 owner")
    }
    check2Owner, err := base64.StdEncoding.DecodeString(args[4])
    if err != nil {
        logger.Error("Failed decoding check2 owner certificate")
        return nil, errors.New("Failed decoding check1 owner")
    }
    check3Owner, err := base64.StdEncoding.DecodeString(args[5])
    if err != nil {
        logger.Error("Failed decoding check3 owner certificate")
        return nil, errors.New("Failed decoding check1 owner")
    }

    logger.Infof("Enrolling device with id:%s", deviceId)

    adminCertificate, err := stub.GetState("admin")
    if err != nil {
        return nil, errors.New("Failed fetching admin identity")
    }
    // Only admin can enroll a device
    ok, err := t.isCaller(stub, adminCertificate)
    if err != nil {
        logger.Error("Failed checking admin identity")
        return nil, errors.New("Failed checking admin identity")
    }
    if !ok {
        logger.Error("Caller is not administrator, cannot enroll a new device")
        return nil, errors.New("The caller is not an administrator")
    }

    ok, err = stub.InsertRow(deviceChecksOwnerMapTable, shim.Row{
        Columns: []*shim.Column{
            &shim.Column{Value: &shim.Column_String_{String_: deviceId}},
            &shim.Column{Value: &shim.Column_Bytes{Bytes: devicePubKey}},
            &shim.Column{Value: &shim.Column_Bytes{Bytes: owner}},
            &shim.Column{Value: &shim.Column_Bytes{Bytes: check1Owner}},
            &shim.Column{Value: &shim.Column_Bytes{Bytes: check2Owner}},
            &shim.Column{Value: &shim.Column_Bytes{Bytes: check3Owner}},
        },
    })

    if !ok && err == nil {
        logger.Errorf("Error in enrolling a new device:%s", err)
        return nil, errors.New("Error in enrolling a new device")
    }
    logger.Infof("Enrolled device %s", deviceId)

    return nil, nil
}

func (t *DeviceMaintenanceChaincode) isCaller(stub *shim.ChaincodeStub, certificate []byte) (bool, error) {
    logger.Debug("Checking caller...")

    // In order to enforce access control, we require that the
    // metadata contains the signature under the signing key corresponding
    // to the verification key inside certificate of
    // the payload of the transaction (namely, function name and args) and
    // the transaction binding (to avoid copying attacks)

    // Verify \sigma=Sign(certificate.sk, tx.Payload||tx.Binding) against certificate.vk
    // \sigma is in the metadata

    sigma, err := stub.GetCallerMetadata()
    if err != nil {
        return false, errors.New("Failed getting metadata")
    }
    payload, err := stub.GetPayload()
    if err != nil {
        return false, errors.New("Failed getting payload")
    }
    binding, err := stub.GetBinding()
    if err != nil {
        return false, errors.New("Failed getting binding")
    }

    logger.Debugf("passed certificate [% x]", certificate)
    logger.Debugf("passed sigma [% x]", sigma)
    logger.Debugf("passed payload [% x]", payload)
    logger.Debugf("passed binding [% x]", binding)

    ok, err := stub.VerifySignature(
        certificate,
        sigma,
        append(payload, binding...),
    )
    if err != nil {
        logger.Errorf("Failed checking signature [%s]", err)
        return ok, err
    }
    if !ok {
        logger.Error("Invalid signature")
    }

    logger.Debug("Check caller...Verified!")

    return ok, err
}

// Deletes an existing device and its entry
func (t *DeviceMaintenanceChaincode) delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    logger.Info("In delete function")
    if len(args) != 1 {
        logger.Error("Incorrect number of arguments")
        return nil, errors.New("Incorrect number of arguments. Specify device ID to be deleted")
    }

    deviceId := args[0]

    logger.Infof("Deleting device with id:%s", deviceId)

    var columns []shim.Column
    col1 := shim.Column{Value: &shim.Column_String_{String_: deviceId}}
    columns = append(columns, col1)
    err := stub.DeleteRow(deviceServiceTable, columns)
    if err != nil {
        logger.Errorf("Error in deleting an service records for device:%s", err)
        return nil, errors.New("Error in deleting service records")
    }
    err = stub.DeleteRow(deviceChecksOwnerMapTable, columns)
    if err != nil {
        logger.Errorf("Error in deleting device:%s", err)
        return nil, errors.New("Error in deleting device")
    }
    logger.Infof("Deleted device %s", deviceId)

    return nil, nil
}

// Query callback representing the query of a chaincode
func (t *DeviceMaintenanceChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    if function == "devices" {
        return t.devices(stub, args)
    }

    return nil, errors.New("Invalid query function name")
}

// Return all meters
func (t *DeviceMaintenanceChaincode) devices(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    logger.Info("In devices function")
    if len(args) > 0 {
        logger.Error("Incorrect number of arguments")
        return nil, errors.New("Incorrect number of arguments. No arguments required")
    }

    var columns []shim.Column

    rowChannel, err := stub.GetRows(deviceChecksOwnerMapTable, columns)
    if err != nil {
        logger.Errorf("Error in getting rows:%s", err.Error())
        return nil, errors.New("Error in fetching rows")
    }
    devices := make([]Device, 0)
    for row := range rowChannel {
        device := Device{
            Id: row.Columns[0].GetString_(),
        }
        devices = append(devices, device)
    }

    payload, err := json.Marshal(devices)
    if err != nil {
        logger.Errorf("Failed marshalling payload: [%s]", err)
        return nil, fmt.Errorf("Failed marshalling payload [%s]", err)
    }

    return payload, nil
}

func main() {
    err := shim.Start(new(DeviceMaintenanceChaincode))
    if err != nil {
        fmt.Printf("Error starting Energy trading chaincode: %s", err)
    }
}


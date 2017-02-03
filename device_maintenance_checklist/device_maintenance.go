package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/crypto/primitives"
)

var logger = shim.NewLogger("device_maintenance")

const (
	deviceChecksOwnerMapTable = "DeviceChecksOwnerMap"
	deviceServiceTable        = "DeviceService"
)

type Devices []Device

type Device struct {
	Id        string `json:"id"`
	PublicKey []byte
	Owner     []byte
	Check1    []byte
	Check2    []byte
	Check3    []byte
}

type DeviceServiceRecords []DeviceServiceRecord

type DeviceServiceRecord struct {
	DeviceId  string `json:"device_id"`
	ServiceId string `json:"service_id"`
	Check1    bool   `json:"check1"`
	Check2    bool   `json:"check2"`
	Check3    bool   `json:"check3"`
	Signoff   bool   `json:"signoff"`
}

// DeviceMaintenance chaincode that provides a way to record maintenance checklist
// for high value assets being worked on by multiple parties.
type DeviceMaintenanceChaincode struct {
}

func (t *DeviceMaintenanceChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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
			&shim.ColumnDefinition{Name: "ServiceId", Type: shim.ColumnDefinition_STRING, Key: true},
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

	return adminCert, nil
}

func (t *DeviceMaintenanceChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	if function == "enroll" {
		return t.enroll(stub, args)
	}

	if function == "startServiceCycle" {
		return t.startServiceCycle(stub, args)
	}

	if function == "markCheckComplete" {
		return t.markCheckComplete(stub, args)
	}

	if function == "signoff" {
		return t.signoff(stub, args)
	}

	if function == "delete" {
		return t.delete(stub, args)

	}

	logger.Errorf("Unimplemented method :%s called", function)

	return nil, errors.New("Unimplemented '" + function + "' invoked")
}

// Enrolls a new device
func (t *DeviceMaintenanceChaincode) enroll(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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
		return nil, fmt.Errorf("Failed getting admin certificate:%s", err.Error())
	}
	// Only admin can enroll a device
	ok, err := t.isCaller(stub, adminCertificate)
	if err != nil {
		logger.Error("Failed checking admin identity")
		return nil, fmt.Errorf("Failed checking admin identity:%s", err.Error())
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

	if !ok || err != nil {
		logger.Errorf("Error in enrolling a new device:%s", err)
		return nil, errors.New("Error in enrolling a new device")
	}
	logger.Infof("Enrolled device %s", deviceId)

	return nil, nil
}

// Starts a new service cycle. Allowed only if called by the owner of the device
func (t *DeviceMaintenanceChaincode) startServiceCycle(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In startServiceCycle function")
	if len(args) != 2 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. Specify device id and service id.")
	}

	deviceId := args[0]
	serviceId := args[1]
	device, err := t.getDevice(stub, deviceId)
	if err != nil {
		logger.Errorf("Failed fetching device: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device [%s]", err)
	}

	logger.Infof("Starting new service cycle for device:%s", deviceId)
	// We could add check to see if there are any active service cycles
	// going on for this device and fail if there is any active service cycle.
	// But for now we will skip that check...

	// Only owner can start a service cycle
	ok, err := t.isCaller(stub, device.Owner)
	if err != nil {
		logger.Error("Failed checking owner identity")
		return nil, errors.New("Failed checking owner identity")
	}
	if !ok {
		logger.Error("Caller is not the owner, cannot start a new service cycle")
		return nil, errors.New("The caller is not the owner of device")
	}

	ok, err = stub.InsertRow(deviceServiceTable, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: deviceId}},
			&shim.Column{Value: &shim.Column_String_{String_: serviceId}},
			&shim.Column{Value: &shim.Column_Bool{Bool: false}},
			&shim.Column{Value: &shim.Column_Bool{Bool: false}},
			&shim.Column{Value: &shim.Column_Bool{Bool: false}},
			&shim.Column{Value: &shim.Column_Bool{Bool: false}},
		},
	})

	if !ok || err != nil {
		logger.Errorf("Error in starting a new service cycle:%s", err)
		return nil, errors.New("Error in starting a new service cycle")
	}
	logger.Infof("New service cycle %s started for device id: %s", serviceId, deviceId)

	return nil, nil
}

// Marks a check complete. Allowed only if the caller is owner of that check
func (t *DeviceMaintenanceChaincode) markCheckComplete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In markCheckComplete function")
	if len(args) != 3 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. Specify device id, service id and the check being performed.")
	}

	deviceId := args[0]
	serviceId := args[1]
	check := args[2]
	device, err := t.getDevice(stub, deviceId)
	if err != nil {
		logger.Errorf("Failed fetching device: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device [%s]", err)
	}

	deviceServiceRecord, err := t.getDeviceServiceRecord(stub, deviceId, serviceId)
	if err != nil {
		logger.Errorf("Failed fetching device service record: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device service record [%s]", err)
	}

	if deviceServiceRecord.ServiceId == "" {
		logger.Errorf("No service record with id [%s] found for device %s", serviceId, deviceId)
		return nil, fmt.Errorf("No service record with id [%s] found for device %s", serviceId, deviceId)
	}

	logger.Infof("Marking check %s completed for device:%s", check, deviceId)
	var checkOwner []byte
	switch check {
	case "check1":
		logger.Debug("Mark Check1 completed")
		checkOwner = device.Check1
		deviceServiceRecord.Check1 = true
	case "check2":
		logger.Debug("Mark Check2 completed")
		checkOwner = device.Check2
		deviceServiceRecord.Check2 = true
	case "check3":
		logger.Debug("Mark Check3 completed")
		checkOwner = device.Check3
		deviceServiceRecord.Check3 = true
	default:
		logger.Errorf("Invalid check specified %s", check)
		return nil, fmt.Errorf("Invalid check specified %s", check)
	}

	// Only check owner can mark check complete
	ok, err := t.isCaller(stub, checkOwner)
	if err != nil {
		logger.Error("Failed checking owner identity")
		return nil, errors.New("Failed checking owner identity")
	}
	if !ok {
		logger.Error("Caller is not the owner for this check, cannot mark it complete")
		return nil, errors.New("Caller is not the owner for this check, cannot mark it complete")
	}

	ok, err = stub.ReplaceRow(deviceServiceTable, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: deviceId}},
			&shim.Column{Value: &shim.Column_String_{String_: serviceId}},
			&shim.Column{Value: &shim.Column_Bool{Bool: deviceServiceRecord.Check1}},
			&shim.Column{Value: &shim.Column_Bool{Bool: deviceServiceRecord.Check2}},
			&shim.Column{Value: &shim.Column_Bool{Bool: deviceServiceRecord.Check3}},
			&shim.Column{Value: &shim.Column_Bool{Bool: deviceServiceRecord.Signoff}},
		},
	})

	if !ok || err != nil {
		logger.Errorf("Error in marking the check complete:%s", err)
		return nil, errors.New("Error in marking the check complete")
	}
	logger.Infof("Check %s completed for device %s", check, deviceId)

	return nil, nil
}

// Sign off the service. Only owner can do it.
func (t *DeviceMaintenanceChaincode) signoff(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In signoff function")
	if len(args) != 2 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. Specify device id and service id.")
	}

	deviceId := args[0]
	device, err := t.getDevice(stub, deviceId)
	if err != nil {
		logger.Errorf("Failed fetching device: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device [%s]", err)
	}

	serviceId := args[1]
	deviceServiceRecord, err := t.getDeviceServiceRecord(stub, deviceId, serviceId)
	if err != nil {
		logger.Errorf("Failed fetching device service record: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device service record [%s]", err)
	}

	if deviceServiceRecord.ServiceId == "" {
		logger.Errorf("No service record with id [%s] found for device %s", serviceId, deviceId)
		return nil, fmt.Errorf("No service record with id [%s] found for device %s", serviceId, deviceId)
	}

	logger.Infof("Marking the service %s complete for device %s", serviceId, deviceId)

	// Only owner can signoff or close the service
	ok, err := t.isCaller(stub, device.Owner)
	if err != nil {
		logger.Error("Failed checking owner identity")
		return nil, errors.New("Failed checking owner identity")
	}
	if !ok {
		logger.Error("Caller is not the owner, cannot start a new service cycle")
		return nil, errors.New("The caller is not the owner of device")
	}

	if !deviceServiceRecord.Check1 || !deviceServiceRecord.Check2 || !deviceServiceRecord.Check3 {
		logger.Error("All checks not completed, cannot close the service cycle")
		return nil, errors.New("All checks not completed, cannot close the service cycle")
	}

	ok, err = stub.ReplaceRow(deviceServiceTable, shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: deviceId}},
			&shim.Column{Value: &shim.Column_String_{String_: serviceId}},
			&shim.Column{Value: &shim.Column_Bool{Bool: deviceServiceRecord.Check1}},
			&shim.Column{Value: &shim.Column_Bool{Bool: deviceServiceRecord.Check2}},
			&shim.Column{Value: &shim.Column_Bool{Bool: deviceServiceRecord.Check3}},
			&shim.Column{Value: &shim.Column_Bool{Bool: true}},
		},
	})

	if !ok || err != nil {
		logger.Errorf("Error in signing off the service cycle:%s", err)
		return nil, errors.New("Error in signing off the service cycle")
	}
	logger.Infof("Service %s completed for device %s", serviceId, deviceId)
	return nil, nil
}

func (t *DeviceMaintenanceChaincode) extractDevice(row shim.Row) Device {
	return Device{
		Id:        row.Columns[0].GetString_(),
		PublicKey: row.Columns[1].GetBytes(),
		Owner:     row.Columns[2].GetBytes(),
		Check1:    row.Columns[3].GetBytes(),
		Check2:    row.Columns[4].GetBytes(),
		Check3:    row.Columns[5].GetBytes(),
	}
}

func (t *DeviceMaintenanceChaincode) getDevice(stub shim.ChaincodeStubInterface, deviceId string) (Device, error) {
	var columns []shim.Column
	col := shim.Column{Value: &shim.Column_String_{String_: deviceId}}
	columns = append(columns, col)

	row, err := stub.GetRow(deviceChecksOwnerMapTable, columns)
	if err != nil {
		logger.Errorf("Error in getting device:%s", err.Error())
		return Device{}, errors.New("Error in fetching device")
	}
	device := t.extractDevice(row)
	return device, nil
}

func (t *DeviceMaintenanceChaincode) getDevices(stub shim.ChaincodeStubInterface) (Devices, error) {
	var columns []shim.Column

	rowChannel, err := stub.GetRows(deviceServiceTable, columns)
	if err != nil {
		logger.Errorf("Error in getting rows:%s", err.Error())
		return nil, errors.New("Error in fetching rows")
	}
	devices := Devices{}
	for row := range rowChannel {
		device := t.extractDevice(row)
		devices = append(devices, device)
	}
	return devices, nil
}

func (t *DeviceMaintenanceChaincode) extractServiceRecord(row shim.Row) DeviceServiceRecord {
	return DeviceServiceRecord{
		DeviceId:  row.Columns[0].GetString_(),
		ServiceId: row.Columns[1].GetString_(),
		Check1:    row.Columns[2].GetBool(),
		Check2:    row.Columns[3].GetBool(),
		Check3:    row.Columns[4].GetBool(),
		Signoff:   row.Columns[5].GetBool(),
	}
}

func (t *DeviceMaintenanceChaincode) getDeviceServiceRecord(stub shim.ChaincodeStubInterface, deviceId, serviceId string) (DeviceServiceRecord, error) {
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: deviceId}}
	col2 := shim.Column{Value: &shim.Column_String_{String_: serviceId}}
	columns = append(columns, col1, col2)

	row, err := stub.GetRow(deviceServiceTable, columns)
	if err != nil {
		logger.Errorf("Error in getting device service record:%s", err.Error())
		return DeviceServiceRecord{}, errors.New("Error in fetching device service record")
	}
	deviceServiceRecord := t.extractServiceRecord(row)
	return deviceServiceRecord, nil
}

func (t *DeviceMaintenanceChaincode) getDeviceServiceRecords(stub shim.ChaincodeStubInterface, deviceId string) (DeviceServiceRecords, error) {
	var columns []shim.Column
	if deviceId != "" {
		col := shim.Column{Value: &shim.Column_String_{String_: deviceId}}
		columns = append(columns, col)
	}

	rowChannel, err := stub.GetRows(deviceServiceTable, columns)
	if err != nil {
		logger.Errorf("Error in getting rows:%s", err.Error())
		return nil, errors.New("Error in fetching rows")
	}
	deviceServiceRecords := DeviceServiceRecords{}
	for row := range rowChannel {
		deviceServiceRecord := t.extractServiceRecord(row)
		deviceServiceRecords = append(deviceServiceRecords, deviceServiceRecord)
	}
	return deviceServiceRecords, nil
}

func (t *DeviceMaintenanceChaincode) isCaller(stub shim.ChaincodeStubInterface, certificate []byte) (bool, error) {
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
		return ok, fmt.Errorf("Failed checking signature [%s]", err)
	}
	if !ok {
		logger.Error("Invalid signature")
	}

	logger.Debug("Check caller...Verified!")

	return ok, err
}

// Deletes an existing device and its entry
func (t *DeviceMaintenanceChaincode) delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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
func (t *DeviceMaintenanceChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "devices" {
		return t.devices(stub, args)
	}
	if function == "device" {
		return t.device(stub, args)
	}
	if function == "deviceServiceRecords" {
		return t.deviceServiceRecords(stub, args)
	}
	if function == "deviceServiceRecord" {
		return t.deviceServiceRecord(stub, args)
	}
	if function == "allServiceRecords" {
		return t.allServiceRecords(stub, args)
	}

	return nil, errors.New("Invalid query function name")
}

// Return a device
func (t *DeviceMaintenanceChaincode) device(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In device function")
	if len(args) != 1 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. Specify device id.")
	}
	deviceId := args[0]
	device, err := t.getDevice(stub, deviceId)
	if err != nil {
		logger.Errorf("Failed fetching device: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device [%s]", err)
	}

	payload, err := json.Marshal(device)
	if err != nil {
		logger.Errorf("Failed marshalling payload: [%s]", err)
		return nil, fmt.Errorf("Failed marshalling payload [%s]", err)
	}

	return payload, nil
}

// Return all devices
func (t *DeviceMaintenanceChaincode) devices(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In devices function")
	if len(args) > 0 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. No arguments required")
	}
	devices, err := t.getDevices(stub)
	if err != nil {
		logger.Errorf("Failed fetching devices: [%s]", err)
		return nil, fmt.Errorf("Failed fetching devices [%s]", err)
	}

	payload, err := json.Marshal(devices)
	if err != nil {
		logger.Errorf("Failed marshalling payload: [%s]", err)
		return nil, fmt.Errorf("Failed marshalling payload [%s]", err)
	}

	return payload, nil
}

// Return a device service record
func (t *DeviceMaintenanceChaincode) deviceServiceRecord(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In deviceServiceRecord function")
	if len(args) != 2 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. Specify device id and service id.")
	}
	deviceId := args[0]
	serviceId := args[1]
	deviceServiceRecord, err := t.getDeviceServiceRecord(stub, deviceId, serviceId)
	if err != nil {
		logger.Errorf("Failed fetching device service record: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device service record [%s]", err)
	}

	payload, err := json.Marshal(deviceServiceRecord)
	if err != nil {
		logger.Errorf("Failed marshalling payload: [%s]", err)
		return nil, fmt.Errorf("Failed marshalling payload [%s]", err)
	}

	return payload, nil
}

// Return service records for a device
func (t *DeviceMaintenanceChaincode) deviceServiceRecords(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In deviceServiceRecord function")
	if len(args) != 1 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. Specify device id.")
	}
	deviceId := args[0]
	deviceServiceRecords, err := t.getDeviceServiceRecords(stub, deviceId)
	if err != nil {
		logger.Errorf("Failed fetching device service records: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device service records [%s]", err)
	}

	payload, err := json.Marshal(deviceServiceRecords)
	if err != nil {
		logger.Errorf("Failed marshalling payload: [%s]", err)
		return nil, fmt.Errorf("Failed marshalling payload [%s]", err)
	}

	return payload, nil
}

// Return service records for a device
func (t *DeviceMaintenanceChaincode) allServiceRecords(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Info("In deviceServiceRecord function")
	if len(args) != 0 {
		logger.Error("Incorrect number of arguments")
		return nil, errors.New("Incorrect number of arguments. No arguments required")
	}
	deviceId := ""
	deviceServiceRecords, err := t.getDeviceServiceRecords(stub, deviceId)
	if err != nil {
		logger.Errorf("Failed fetching device service records: [%s]", err)
		return nil, fmt.Errorf("Failed fetching device service records [%s]", err)
	}

	payload, err := json.Marshal(deviceServiceRecords)
	if err != nil {
		logger.Errorf("Failed marshalling payload: [%s]", err)
		return nil, fmt.Errorf("Failed marshalling payload [%s]", err)
	}

	return payload, nil
}

func main() {
	primitives.SetSecurityLevel("SHA3", 256)
	logger.SetLevel(shim.LogDebug)
	err := shim.Start(new(DeviceMaintenanceChaincode))
	if err != nil {
		fmt.Printf("Error starting Energy trading chaincode: %s", err)
	}
}

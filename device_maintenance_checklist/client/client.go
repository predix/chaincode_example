package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/crypto"
	pb "github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
)

var (
	// Logging
	logger = logging.MustGetLogger("app")

	deviceId  string
	serviceId string

	// NVP related objects
	peerClientConn *grpc.ClientConn
	serverClient   pb.PeerClient

	// Alice is the deployer
	alice     crypto.Client
	aliceCert crypto.CertificateHandler

	// Bob is the owner of the device
	bob     crypto.Client
	bobCert crypto.CertificateHandler

	// Carol, Dave and Finn are owners for check1, check2 and check3 respectively
	carol     crypto.Client
	carolCert crypto.CertificateHandler

	dave     crypto.Client
	daveCert crypto.CertificateHandler

	finn     crypto.Client
	finnCert crypto.CertificateHandler
)

type Device struct {
	Id string `json:"id"`
}

type DeviceServiceRecord struct {
	DeviceId  string `json:"device_id"`
	ServiceId string `json:"service_id"`
	Check1    bool   `json:"check1"`
	Check2    bool   `json:"check2"`
	Check3    bool   `json:"check3"`
	Signoff   bool   `json:"signoff"`
}

func deploy() (err error) {
	logger.Debug("Alice is deployer and deploys the chaincode")

	aliceCert, err = alice.GetEnrollmentCertificateHandler()
	if err != nil {
		logger.Errorf("Failed getting Alice TCert [%s]", err)
		return
	}

	resp, err := deployInternal(alice, aliceCert)
	if err != nil {
		logger.Errorf("Failed deploying [%s]", err)
		return
	}
	logger.Debugf("Resp [%s]", resp.String())
	logger.Debugf("Chaincode NAME: [%s]-[%s]", chaincodeName, string(resp.Msg))

	logger.Debug("Wait 30 seconds")
	time.Sleep(30 * time.Second)

	logger.Debug("------------- Done!")
	return
}

func enroll() (err error) {
	logger.Debug("Alice enrolls a new device by assigning bob as owner")

	// 1. Alice is the administrator of the chaincode;
	// 2. Alice enrolls a new device and assigns ownership to Bob.
	// 3. Alice also assigns ownership of check1, check2 and check3 to
	//    carol, dave and finn respectively
	deviceId = "Device6"

	bobCert, err = bob.GetEnrollmentCertificateHandler()
	if err != nil {
		logger.Errorf("Failed getting Bob TCert [%s]", err)
		return
	}
	carolCert, err = carol.GetEnrollmentCertificateHandler()
	if err != nil {
		logger.Errorf("Failed getting Carol TCert [%s]", err)
		return
	}
	daveCert, err = dave.GetEnrollmentCertificateHandler()
	if err != nil {
		logger.Errorf("Failed getting Dave TCert [%s]", err)
		return
	}
	finnCert, err = finn.GetEnrollmentCertificateHandler()
	if err != nil {
		logger.Errorf("Failed getting Finn TCert [%s]", err)
		return
	}

	resp, err := enrollInternal(alice, aliceCert, deviceId, bobCert, carolCert, daveCert, finnCert)
	if err != nil {
		logger.Errorf("Failed enrolling a new device [%s]", err)
		return
	}
	logger.Debugf("Resp [%s]", resp.String())

	logger.Debug("Device enrollment transaction submitted")
	logger.Debug("Wait 10 seconds...")
	time.Sleep(10 * time.Second)

	// Now query the device and see if it is enrolled
	device, err := getDevice(bob, deviceId)
	if err != nil {
		logger.Errorf("Failed querying device record [%s]", err)
		return
	}
	if device.Id != deviceId {
		err = fmt.Errorf("Device not enrolled!")
		logger.Errorf("Device not enrolled!")
		return
	}

	logger.Debug("Enrollment of device Done!")
	return
}

func startServiceCycle() (err error) {
	logger.Debug("Starting service cycle for device")
	v := time.Now().UnixNano()
	serviceId = strconv.FormatInt(v, 16)

	// Bob is the owner. So only bob should be able to start the cycle

	// First try with alice and verify that cycle cannot be started
	resp, err := startServiceCycleInternal(deviceId, serviceId, aliceCert)
	if err != nil {
		logger.Errorf("Failed invoking start service cycle [%s]", err)
		return
	}
	logger.Debugf("Resp [%s]", resp.String())

	logger.Debug("Wait 10 seconds")
	time.Sleep(10 * time.Second)

	// Now query the device service record...it shouldn't exist
	deviceServiceRecords, err := getAllServiceRecords(bob)
	if err != nil {
		logger.Errorf("Failed querying device service record [%s]", err)
		return
	}
	for _, deviceServiceRecord := range deviceServiceRecords {
		if deviceServiceRecord.DeviceId == deviceId &&
			deviceServiceRecord.ServiceId == serviceId {
			err = fmt.Errorf("Service cycle started by non-owner!")
			logger.Errorf("Service cycle started by non-owner!")
			return
		}
	}

	// Now try with bob
	resp, err = startServiceCycleInternal(deviceId, serviceId, bobCert)
	if err != nil {
		logger.Errorf("Failed invoking start service cycle [%s]", err)
		return
	}
	logger.Debugf("Resp [%s]", resp.String())

	logger.Debug("Wait 10 seconds")
	time.Sleep(10 * time.Second)

	// Now query the device service record...it should exist
	deviceServiceRecord, err := getDeviceServiceRecord(bob, deviceId, serviceId)
	if err != nil {
		logger.Errorf("Failed querying device service record [%s]", err)
		return
	}
	if deviceServiceRecord.ServiceId == "" {
		err = fmt.Errorf("Service cycle not started!")
		logger.Errorf("Service cycle not started!")
		return
	}
	logger.Infof("StartServiceCycle test successful-------")
	return
}

func markCheckComplete(check string, ownerCert, nonOwnerCert crypto.CertificateHandler) (err error) {
	logger.Debugf("Marking %s as complete", check)

	// First try with non owner and verify that check is not marked complete
	logger.Debugf("First trying with non-owner")
	resp, err := markCheckCompleteInternal(deviceId, serviceId, check, nonOwnerCert)
	if err != nil {
		logger.Errorf("Failed invoking markCheckComplete [%s]", err)
		return
	}
	logger.Debugf("Resp [%s]", resp.String())

	logger.Debug("Wait 10 seconds")
	time.Sleep(10 * time.Second)

	// Now query the device service record...check1 should be false
	deviceServiceRecord, err := getDeviceServiceRecord(bob, deviceId, serviceId)
	if err != nil {
		logger.Errorf("Failed querying device service record [%s]", err)
		return
	}
	switch check {
	case "check1":
		if deviceServiceRecord.Check1 {
			logger.Errorf("Check: %s marked complete by non-owner", check)
			err = fmt.Errorf("Check: %s marked complete by non-owner", check)
			return
		}
	case "check2":
		if deviceServiceRecord.Check2 {
			logger.Errorf("Check: %s marked complete by non-owner", check)
			err = fmt.Errorf("Check: %s marked complete by non-owner", check)
			return
		}
	case "check3":
		if deviceServiceRecord.Check3 {
			logger.Errorf("Check: %s marked complete by non-owner", check)
			err = fmt.Errorf("Check: %s marked complete by non-owner", check)
			return
		}
	}
	logger.Debugf("Now trying with owner")

	// Now try with owner
	resp, err = markCheckCompleteInternal(deviceId, serviceId, check, ownerCert)
	if err != nil {
		logger.Errorf("Failed invoking markCheckComplete [%s]", err)
		return
	}
	logger.Debugf("Resp [%s]", resp.String())

	logger.Debug("Wait 10 seconds")
	time.Sleep(10 * time.Second)

	// Now query the device service record...check1 should be false
	deviceServiceRecord, err = getDeviceServiceRecord(bob, deviceId, serviceId)
	if err != nil {
		logger.Errorf("Failed querying device service record [%s]", err)
		return
	}
	switch check {
	case "check1":
		if !deviceServiceRecord.Check1 {
			logger.Errorf("Check: %s marked complete by non-owner", check)
			err = fmt.Errorf("Check: %s marked complete by non-owner", check)
			return
		}
	case "check2":
		if !deviceServiceRecord.Check2 {
			logger.Errorf("Check: %s marked complete by non-owner", check)
			err = fmt.Errorf("Check: %s marked complete by non-owner", check)
			return
		}
	case "check3":
		if !deviceServiceRecord.Check3 {
			logger.Errorf("Check: %s marked complete by non-owner", check)
			err = fmt.Errorf("Check: %s marked complete by non-owner", check)
			return
		}
	}
	logger.Infof("Mark check:%s complete test successful-------", check)
	return
}

func signOff() (err error) {
	logger.Debug("Signing off the service")
	// Bob is the owner. So only bob should be able to signoff

	// First try with alice and verify that signoff is not allowed
	resp, err := signoffInternal(deviceId, serviceId, aliceCert)
	if err != nil {
		logger.Errorf("Failed invoking signoff [%s]", err)
		return
	}
	logger.Debugf("Resp [%s]", resp.String())

	logger.Debug("Wait 10 seconds")
	time.Sleep(10 * time.Second)
	// Now query the device service record...check1 should be false
	deviceServiceRecord, err := getDeviceServiceRecord(bob, deviceId, serviceId)
	if err != nil {
		logger.Errorf("Failed querying device service record [%s]", err)
		return
	}
	if deviceServiceRecord.Signoff {
		logger.Errorf("Signoff allowed by non-owner!")
		err = fmt.Errorf("Signoff allowed by non-owner!")
		return
	}
	// Now try with bob and verify that signoff is allowed
	resp, err = signoffInternal(deviceId, serviceId, bobCert)
	if err != nil {
		logger.Errorf("Failed invoking signoff [%s]", err)
		return
	}
	logger.Debugf("Resp [%s]", resp.String())

	logger.Debug("Wait 10 seconds")
	time.Sleep(10 * time.Second)
	// Now query the device service record...check1 should be false
	deviceServiceRecord, err = getDeviceServiceRecord(bob, deviceId, serviceId)
	if err != nil {
		logger.Errorf("Failed querying device service record [%s]", err)
		return
	}
	if !deviceServiceRecord.Signoff {
		logger.Errorf("Signoff not allowed by owner!")
		err = fmt.Errorf("Signoff not allowed by owner!")
		return
	}

	logger.Infof("Signoff test successful-------")
	return
}

func testDeviceMaintenanceChecklistChaincode() (err error) {
	// Deploy
	err = deploy()
	if err != nil {
		logger.Errorf("Failed deploying [%s]", err)
		return
	}

	// Enroll a new device
	err = enroll()
	if err != nil {
		logger.Errorf("Failed assigning ownership [%s]", err)
		return
	}

	err = startServiceCycle()
	if err != nil {
		logger.Errorf("Failed starting service cycle [%s]", err)
		return
	}

	err = markCheckComplete("check1", carolCert, daveCert)
	if err != nil {
		logger.Errorf("Failed marking check1 complete [%s]", err)
		return
	}

	err = markCheckComplete("check2", daveCert, finnCert)
	if err != nil {
		logger.Errorf("Failed marking check2 complete [%s]", err)
		return
	}

	err = markCheckComplete("check3", finnCert, aliceCert)
	if err != nil {
		logger.Errorf("Failed marking check3 complete [%s]", err)
		return
	}

	err = signOff()
	if err != nil {
		logger.Errorf("Failed signoff [%s]", err)
		return
	}

	logger.Infof("-------------ALL SUCCESS-----------")

	return
}

func main() {
	// Acts as non validating client that submits transactions to block chain.
	// A 'core.yaml' file is assumed to be available in the working directory.
	if err := initNVP(); err != nil {
		logger.Debugf("Failed initiliazing NVP [%s]", err)
		os.Exit(-1)
	}

	// Enable fabric 'confidentiality'
	confidentiality(false)

	// Exercise the 'device maintenance checklist' chaincode
	if err := testDeviceMaintenanceChecklistChaincode(); err != nil {
		logger.Debugf("Failed testing device maintenance checklist chaincode [%s]", err)
		os.Exit(-2)
	}
}

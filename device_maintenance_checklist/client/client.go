package main

import (
    "os"
    "time"

    "github.com/hyperledger/fabric/core/crypto"
    pb "github.com/hyperledger/fabric/protos"
    "github.com/op/go-logging"
    "google.golang.org/grpc"
)

var (
    // Logging
    logger = logging.MustGetLogger("app")

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

func deploy() (err error) {
    logger.Debug("Alice is deployer and deploys the chaincode")

    aliceCert, err = alice.GetTCertificateHandlerNext()
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

    bobCert, err = bob.GetTCertificateHandlerNext()
    if err != nil {
        logger.Errorf("Failed getting Bob TCert [%s]", err)
        return
    }
    carolCert, err = carol.GetTCertificateHandlerNext()
    if err != nil {
        logger.Errorf("Failed getting Carol TCert [%s]", err)
        return
    }
    daveCert, err = dave.GetTCertificateHandlerNext()
    if err != nil {
        logger.Errorf("Failed getting Dave TCert [%s]", err)
        return
    }
    finnCert, err = finn.GetTCertificateHandlerNext()
    if err != nil {
        logger.Errorf("Failed getting Finn TCert [%s]", err)
        return
    }

    resp, err := enrollInternal(alice, aliceCert, "Device1", bobCert, carolCert, daveCert, finnCert)
    if err != nil {
        logger.Errorf("Failed enrolling a new device [%s]", err)
        return
    }
    logger.Debugf("Resp [%s]", resp.String())

    logger.Debug("Wait 30 seconds")
    time.Sleep(30 * time.Second)

    logger.Debug("------------- Done!")
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


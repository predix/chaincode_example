package main

import (
    "encoding/base64"
    "errors"
    "fmt"

    "github.com/golang/protobuf/proto"
    "github.com/hyperledger/fabric/core/chaincode"
    "github.com/hyperledger/fabric/core/chaincode/platforms"
    "github.com/hyperledger/fabric/core/config"
    "github.com/hyperledger/fabric/core/container"
    "github.com/hyperledger/fabric/core/crypto"
    "github.com/hyperledger/fabric/core/peer"
    "github.com/hyperledger/fabric/core/util"
    pb "github.com/hyperledger/fabric/protos"
    "github.com/op/go-logging"
    "github.com/spf13/viper"
    "golang.org/x/net/context"
)

var (
    confidentialityOn bool

    confidentialityLevel pb.ConfidentialityLevel
    chaincodeName        string
)

func initNVP() (err error) {
    if err = initPeerClient(); err != nil {
        logger.Debugf("Failed deploying [%s]", err)
        return

    }
    if err = initCryptoClients(); err != nil {
        logger.Debugf("Failed deploying [%s]", err)
        return
    }

    return
}

func initPeerClient() (err error) {
    config.SetupTestConfig(".")
    viper.Set("ledger.blockchain.deploy-system-chaincode", "false")
    viper.Set("peer.validator.validity-period.verification", "false")

    peerClientConn, err = peer.NewPeerClientConnection()
    if err != nil {
        fmt.Printf("error connection to server at host:port = %s\n", viper.GetString("peer.address"))
        return
    }
    serverClient = pb.NewPeerClient(peerClientConn)

    // Logging
    var formatter = logging.MustStringFormatter(
        `%{color}[%{module}] %{shortfunc} [%{shortfile}] -> %{level:.4s} %{id:03x}%{color:reset} %{message}`,
    )
    logging.SetFormatter(formatter)

    return
}

func initCryptoClient(name, secret string) (crypto.Client, error) {
    if err := crypto.RegisterClient(name, nil, name, secret); err != nil {
        return nil, err
    }
    logger.Debugf("Completed registering client %s", name)
    client, err := crypto.InitClient(name, nil)
    if err != nil {
        return nil, err
    }
    logger.Debugf("Completed initializing client %s", name)
    return client, nil
}

func initCryptoClients() error {
    var err error
    crypto.Init()

    // Initialize the clients mapping alice, bob, carol, dave and finn
    // to identities already defined in 'membersrvc.yaml'

    alice, err = initCryptoClient("alice", "NPKYL39uKbkj")
    if err != nil {
        return err
    }

    bob, err = initCryptoClient("bob", "DRJ20pEql15a")
    if err != nil {
        return err
    }

    carol, err = initCryptoClient("carol", "DRJ23pEQl16a")
    if err != nil {
        return err
    }

    dave, err = initCryptoClient("dave", "7avZQLwcUe9q")
    if err != nil {
        return err
    }

    finn, err = initCryptoClient("finn", "6avZQLwcUe9b")
    if err != nil {
        return err
    }

    logger.Info("Completed initializing required clients for this test")

    return nil
}

func processTransaction(tx *pb.Transaction) (*pb.Response, error) {
    return serverClient.ProcessTransaction(context.Background(), tx)
}

func confidentiality(enabled bool) {
    confidentialityOn = enabled

    if confidentialityOn {
        confidentialityLevel = pb.ConfidentialityLevel_CONFIDENTIAL
    } else {
        confidentialityLevel = pb.ConfidentialityLevel_PUBLIC
    }
}

func deployInternal(deployer crypto.Client, adminCert crypto.CertificateHandler) (resp *pb.Response, err error) {
    // Prepare the spec. The metadata includes the identity of the administrator
    spec := &pb.ChaincodeSpec{
        Type:                 1,
        ChaincodeID:          &pb.ChaincodeID{Path: "github.com/atulkc/chaincode_example/device_maintenance_checklist"},
        CtorMsg:              &pb.ChaincodeInput{Function: "init", Args: []string{}},
        Metadata:             adminCert.GetCertificate(),
        ConfidentialityLevel: confidentialityLevel,
    }

    // First build the deployment spec
    cds, err := getChaincodeBytes(spec)
    if err != nil {
        return nil, fmt.Errorf("Error getting deployment spec: %s ", err)
    }

    // Now create the Transactions message and send to Peer.
    transaction, err := deployer.NewChaincodeDeployTransaction(cds, cds.ChaincodeSpec.ChaincodeID.Name)
    if err != nil {
        return nil, fmt.Errorf("Error deploying chaincode: %s ", err)
    }

    resp, err = processTransaction(transaction)

    logger.Debugf("resp [%s]", resp.String())

    
     chaincodeName = cds.ChaincodeSpec.ChaincodeID.Name
    logger.Debugf("ChaincodeName [%s]", chaincodeName)

    return
}

func enrollInternal(invoker crypto.Client, invokerCert crypto.CertificateHandler, deviceId string, ownerCert, check1Cert, check2Cert, check3Cert crypto.CertificateHandler) (resp *pb.Response, err error) {
    // Get a transaction handler to be used to submit the execute transaction
    // and bind the chaincode access control logic using the binding
    submittingCertHandler, err := invoker.GetTCertificateHandlerNext()
    if err != nil {
        return nil, err
    }
    txHandler, err := submittingCertHandler.GetTransactionHandler()
    if err != nil {
        return nil, err
    }
    binding, err := txHandler.GetBinding()
    if err != nil {
        return nil, err
    }

    pubKey := base64.StdEncoding.EncodeToString([]byte("public key"))
    owner := base64.StdEncoding.EncodeToString(ownerCert.GetCertificate())
    check1 := base64.StdEncoding.EncodeToString(check1Cert.GetCertificate())
    check2 := base64.StdEncoding.EncodeToString(check2Cert.GetCertificate())
    check3 := base64.StdEncoding.EncodeToString(check3Cert.GetCertificate())

    chaincodeInput := &pb.ChaincodeInput{
        Function: "enroll",
        Args:     []string{deviceId, pubKey, owner, check1, check2, check3},
    }
    chaincodeInputRaw, err := proto.Marshal(chaincodeInput)
    if err != nil {
        return nil, err
    }

    // Access control. Administrator signs chaincodeInputRaw || binding to confirm his identity
    sigma, err := invokerCert.Sign(append(chaincodeInputRaw, binding...))
    if err != nil {
        return nil, err
    }

    // Prepare spec and submit
    spec := &pb.ChaincodeSpec{
        Type:                 1,
        ChaincodeID:          &pb.ChaincodeID{Name: chaincodeName},
        CtorMsg:              chaincodeInput,
        Metadata:             sigma, // Proof of identity
        ConfidentialityLevel: confidentialityLevel,
    }

    chaincodeInvocationSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

    // Now create the Transactions message and send to Peer.
    transaction, err := txHandler.NewChaincodeExecute(chaincodeInvocationSpec, util.GenerateUUID())
    if err != nil {
        return nil, fmt.Errorf("Error invoking chaincode: %s ", err)
    }

    return processTransaction(transaction)
}

func getChaincodeBytes(spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
    mode := viper.GetString("chaincode.mode")
    var codePackageBytes []byte
    if mode != chaincode.DevModeUserRunsChaincode {
        logger.Debugf("Received build request for chaincode spec: %v", spec)
        var err error
        if err = checkSpec(spec); err != nil {
            return nil, err
        }

        codePackageBytes, err = container.GetChaincodePackageBytes(spec)
        if err != nil {
            err = fmt.Errorf("Error getting chaincode package bytes: %s", err)
            logger.Errorf("%s", err)
            return nil, err
        }
    }
    chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
    return chaincodeDeploymentSpec, nil
}

func checkSpec(spec *pb.ChaincodeSpec) error {
    // Don't allow nil value
    if spec == nil {
        return errors.New("Expected chaincode specification, nil received")
    }

    platform, err := platforms.Find(spec.Type)
    if err != nil {
        return fmt.Errorf("Failed to determine platform type: %s", err)
    }

    return platform.ValidateSpec(spec)
}


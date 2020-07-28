package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

const (
	LAYOUT_TIMESTAMP = "2006-01-02T15:04:05"
)

func GetOrg(stub shim.ChaincodeStubInterface) (string, error) {
	org, ok, err := cid.GetAttributeValue(stub, "hf.Affiliation")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("Org Not Found")
	}
	return org, nil
}

func SetupArgsArray(funcName string, args ...string) [][]byte {
	// Create an args array with 1 additional element for the funcName
	ccArgs := make([][]byte, 1+len(args))

	// Setup the function name
	ccArgs[0] = []byte(funcName)

	// Set up the args array
	for i, arg := range args {
		ccArgs[i+1] = []byte(arg)
	}

	return ccArgs
}

func main() {
	airplaneCC := new(AirplaneCC)
	err := shim.Start(airplaneCC)
	if err != nil {
		panic(err.Error())
	}
}

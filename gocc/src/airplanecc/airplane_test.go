package main

import (
	"encoding/json"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/msp"
)

var (
	scc  = new(AirplaneCC)
	stub = shimtest.NewMockStub("planeStub", scc)
)

const boeingCert = `
-----BEGIN CERTIFICATE-----
MIICrzCCAlWgAwIBAgIUPoHFsm8QlslqR1ct1jUiZDh9XEowCgYIKoZIzj0EAwIw
XDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMQ8wDQYDVQQH
EwZEdXJoYW0xDzANBgNVBAoTBmJvZWluZzESMBAGA1UEAxMJY2EtYm9laW5nMB4X
DTIwMDcxNzA1MzYwMFoXDTIxMDcxNzA1NDEwMFowbDELMAkGA1UEBhMCVVMxFzAV
BgNVBAgTDk5vcnRoIENhcm9saW5hMRQwEgYDVQQKEwtIeXBlcmxlZGdlcjEeMA0G
A1UECxMGY2xpZW50MA0GA1UECxMGYm9laW5nMQ4wDAYDVQQDEwV1c2VyMTBZMBMG
ByqGSM49AgEGCCqGSM49AwEHA0IABAPd9vtB48qFn2EFbJy2z3Va9RAeiM6HUAFH
7pqxfnnt5prLlLPaOrX6B8YycilgneJE0IJhjjqykKnlSoFzoh+jgeQwgeEwDgYD
VR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFLyfpY56NRdg4Lxl
lLg1h8aYbgbIMB8GA1UdIwQYMBaAFMN9uxiHr49zhv90Ja7IVDKHiERYMCEGA1Ud
EQQaMBiCFm1hY3MtTWFjQm9vay1Qcm8ubG9jYWwwXgYIKgMEBQYHCAEEUnsiYXR0
cnMiOnsiaGYuQWZmaWxpYXRpb24iOiJib2VpbmciLCJoZi5FbnJvbGxtZW50SUQi
OiJ1c2VyMSIsImhmLlR5cGUiOiJjbGllbnQifX0wCgYIKoZIzj0EAwIDSAAwRQIh
APyXi4qBG9z7OC9Dm8BtesCJPzHEC6yDS8HfSl+9relbAiAd5A5AA0uUDUxevWBJ
FDsRMp74TaPaa4zoH0Pgfvao2g==
-----END CERTIFICATE-----
`

const airbusCert = `
-----BEGIN CERTIFICATE-----
MIICqjCCAlGgAwIBAgIUYsdhlD6gTlYiY9AHgGW81X+j29UwCgYIKoZIzj0EAwIw
WDELMAkGA1UEBhMCVUsxEjAQBgNVBAgTCUhhbXBzaGlyZTEQMA4GA1UEBxMHSHVy
c2xleTEPMA0GA1UEChMGYWlyYnVzMRIwEAYDVQQDEwljYS1haXJidXMwHhcNMjAw
NzE3MDUzNjAwWhcNMjEwNzE3MDU0MTAwWjBsMQswCQYDVQQGEwJVUzEXMBUGA1UE
CBMOTm9ydGggQ2Fyb2xpbmExFDASBgNVBAoTC0h5cGVybGVkZ2VyMR4wDQYDVQQL
EwZjbGllbnQwDQYDVQQLEwZhaXJidXMxDjAMBgNVBAMTBXVzZXIxMFkwEwYHKoZI
zj0CAQYIKoZIzj0DAQcDQgAEUJ3mc8Sg7Ev5efW3MG6q9i+gH6Hd02kGKULEB3PJ
c7UoyMcOQp9BVyU4qVzLyjdJ6/IYkmc1xqIvXBMqnkJp6KOB5DCB4TAOBgNVHQ8B
Af8EBAMCB4AwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUucTnnMjtOB1P6xO8CslW
MeNFBnEwHwYDVR0jBBgwFoAUq4YrEgccWcnzXPl86fa3eImRrpcwIQYDVR0RBBow
GIIWbWFjcy1NYWNCb29rLVByby5sb2NhbDBeBggqAwQFBgcIAQRSeyJhdHRycyI6
eyJoZi5BZmZpbGlhdGlvbiI6ImFpcmJ1cyIsImhmLkVucm9sbG1lbnRJRCI6InVz
ZXIxIiwiaGYuVHlwZSI6ImNsaWVudCJ9fTAKBggqhkjOPQQDAgNHADBEAiBzLal7
ordZP3bXySSRTJqubq7P0EOADYZ+DNdHeUu+VgIgJFEzGCDGM52wilL51cQnhAvl
dlZx77Ee+m7UzgeDWkY=
-----END CERTIFICATE-----
`

func setCreator(t *testing.T, stub *shimtest.MockStub, mspID string, idbytes []byte) {
	sid := &msp.SerializedIdentity{Mspid: mspID, IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b
}

func TestInit(t *testing.T) {
	response := stub.MockInit("Init", nil)
	if response.Status != shim.OK {
		t.Error("Error Init")
	}
	t.Log(string(response.Payload))
}

func TestCreateAirplaneSuccess(t *testing.T) {
	setCreator(t, stub, "BoeingMSP", []byte(boeingCert))

	args := SetupArgsArray("CreatePlane",
		"123",
		"Batch1",
		"2012-01-02",
		"BA-737MAx",
		"2",
		"123023,123123",
		"200",
		"1000.5",
		"900.5",
	)
	response := stub.MockInvoke("TestInvoke", args)

	if response.Status != shim.OK {
		t.Error(response.Message)
	}

	t.Log(string(response.Payload))
}

func TestQueryBySN(t *testing.T) {
	args := SetupArgsArray("QueryBySN", "123")
	response := stub.MockInvoke("TestQueryBySN", args)

	if response.Status != shim.OK {
		t.Error(response.Message)
	}

	t.Log(string(response.Payload))
}

func TestChangeOwnership(t *testing.T) {
	setCreator(t, stub, "AirbusMSP", []byte(airbusCert))
	args := SetupArgsArray("ChangeOwnership", "123", "airbus")
	response := stub.MockInvoke("TestChangeOwnership", args)
	if response.Status == shim.OK {
		t.Error("It should be failed")
	}

	setCreator(t, stub, "BoeingMSP", []byte(boeingCert))
	args = SetupArgsArray("ChangeOwnership", "123", "airbus")
	response = stub.MockInvoke("TestChangeOwnership", args)
	if response.Status != shim.OK {
		t.Error(response.Message)
	}

	args = SetupArgsArray("QueryBySN", "123")
	response = stub.MockInvoke("TestQueryBySN", args)
	if response.Status != shim.OK {
		t.Error(response.Message)
	}
	airplane := new(Airplane)
	_ = json.Unmarshal(response.Payload, airplane)
	if airplane.OwnerOrg != "airbus" || airplane.UserOrg != "" {
		t.Error("Error Changing Ownership")
	}

	t.Log(airplane)
}

func TestConfirmOwnership(t *testing.T) {
	setCreator(t, stub, "BoeingMSP", []byte(boeingCert))
	args := SetupArgsArray("ConfirmOwnership", "123")
	response := stub.MockInvoke("TestConfirmOwnership", args)
	if response.Status == shim.OK {
		t.Error("It Should Failed")
	}

	setCreator(t, stub, "AirbusMSP", []byte(airbusCert))
	args = SetupArgsArray("ConfirmOwnership", "123")
	response = stub.MockInvoke("TestConfirmOwnership", args)
	if response.Status != shim.OK {
		t.Error(response.Message)
	}

	t.Log(string(response.Payload))

}

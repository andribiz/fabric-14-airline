package main

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/msp"
)

var (
	stub = shimtest.NewMockStub("invoiceStub", new(InvoiceCC))
	key  string
)

const cert = `
-----BEGIN CERTIFICATE-----
MIIClTCCAjugAwIBAgIUUV4jk6W+mOuMLwqE5ezrfCcWxXYwCgYIKoZIzj0EAwIw
VDELMAkGA1UEBhMCVUsxEjAQBgNVBAgTCUhhbXBzaGlyZTEQMA4GA1UEBxMHSHVy
c2xleTENMAsGA1UEChMEb3JnMTEQMA4GA1UEAxMHY2Etb3JnMTAeFw0yMDA3Mjgx
MDIwMDBaFw0yMTA3MjgxMDI1MDBaMGsxCzAJBgNVBAYTAklEMRIwEAYDVQQIEwlK
YXdhVGltdXIxDDAKBgNVBAcTA1NCWTENMAsGA1UEChMEb3JnMTEcMA0GA1UECxMG
Y2xpZW50MAsGA1UECxMEb3JnMTENMAsGA1UEAxMEZGluaTBZMBMGByqGSM49AgEG
CCqGSM49AwEHA0IABPLY8AKUpGclhlPh7o7aZGDw4zyj8E9r5qcNxTYRJ0kstsJc
G4m3nzNI3AhGyss3LQMDcr2aZ7ml7qcNq3tTZrGjgdMwgdAwDgYDVR0PAQH/BAQD
AgeAMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFBPyH8TymCCKVA3YjUzzGCRQnJ/I
MB8GA1UdIwQYMBaAFP/XWLU/uSJa4rINYZU/14YyevfvMHAGCCoDBAUGBwgBBGR7
ImF0dHJzIjp7ImFwcHMuYWRtaW4iOiJmYWxzZSIsImhmLkFmZmlsaWF0aW9uIjoi
b3JnMSIsImhmLkVucm9sbG1lbnRJRCI6ImRpbmkiLCJoZi5UeXBlIjoiY2xpZW50
In19MAoGCCqGSM49BAMCA0gAMEUCIQCTTvf38BVmy7RQs1rJYF0iYh0nNLu4DbBh
FNwphcXuuwIgIT0oY0+rghWRVqG3+to0DgLLJd9YuhKADQvwE/D+P58=
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

func TestCreate(t *testing.T) {
	response := stub.MockInit("Init", nil)
	if response.Status != shim.OK {
		t.Error("Init Error")
	}
}

func TestCreateInvoice(t *testing.T) {
	setCreator(t, stub, "Org1MSP", []byte(cert))
	args := [][]byte{
		[]byte("CreateInvoice"),
		[]byte("123"),
		[]byte("2020-07-29"),
		[]byte("User1"),
		[]byte("Org2MSP"),
		[]byte("1000000"),
		[]byte("100000"),
	}
	response := stub.MockInvoke("testCreate", args)
	if response.Status != shim.OK {
		t.Error("Error Create Invoice " + response.Message)
	}
	key = string(response.Payload)
	t.Log(key)
}

func TestGetInvoiceByID(t *testing.T) {
	args := [][]byte{
		[]byte("GetInvoiceByID"),
		[]byte(key),
	}
	resp := stub.MockInvoke("TestQuery", args)
	if resp.Status != shim.OK {
		t.Error(resp.Message)
	}
	t.Log(string(resp.Payload))
}

func TestGetInvoiceByPartner(t *testing.T) {
	setCreator(t, stub, "Org1MSP", []byte(cert))
	args := [][]byte{
		[]byte("CreateInvoice"),
		[]byte("456"),
		[]byte("2020-07-30"),
		[]byte("User1"),
		[]byte("Org2MSP"),
		[]byte("1000000"),
		[]byte("100000"),
	}
	_ = stub.MockInvoke("testCreate456", args)

	args = [][]byte{
		[]byte("GetInvoiceByPartner"),
		[]byte("dini"),
	}

	resp := stub.MockInvoke("TestInvoiceByPartner", args)
	if resp.Status != shim.OK {
		t.Error(resp.Message)
	}
	t.Log(string(resp.Payload))
}

// func TestQueryByPagination(t *testing.T) {
// 	args := [][]byte{
// 		[]byte("QueryInvoiceByPartner"),
// 		[]byte("dini"),
// 		[]byte(""),
// 	}
// 	resp := stub.MockInvoke("TestQueryByPagination", args)
// 	if resp.Status != shim.OK {
// 		t.Error(resp.Message)
// 	}
// 	t.Log(string(resp.Payload))
// }

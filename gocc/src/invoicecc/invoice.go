package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type InvoiceCC struct {
}

type InvoiceState string

const (
	OPEN InvoiceState = "OPEN"
	PAID InvoiceState = "PAID"
)

type Invoice struct {
	DocType          string       `json:"docType"`
	InvID            string       `json:"InvID"`
	InvNumber        string       `json:"invNumber"`
	InvDate          time.Time    `json:"invDate"`
	PartnerID        string       `json:"partnerID"`
	PartnerOrg       string       `json:"partnerOrg"`
	CounterPartnerID string       `json:"counterPartnerID"`
	CounterPrtnerOrg string       `json:"counterPartnerOrg"`
	Subtotal         float64      `json:"subtotal"`
	VAT              float64      `json:"vat"`
	State            InvoiceState `json:"state"`
}

type InvoiceQueryResult struct {
	FetchedRecordsCount int32
	Payload             []*Invoice
	Bookmark            string
}

type InvoiceLine struct {
	DocType    string  `json:"docType"`
	InvID      string  `json:"invID"`
	Product    string  `json:"product"`
	ProductQty int32   `json:"productQty"`
	ProductUom string  `json:"productUom"`
	Price      float64 `json:"price"`
}

func (invCC *InvoiceCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("Initialized"))
}

func (invCC *InvoiceCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, args := stub.GetFunctionAndParameters()

	switch funcName {
	case "CreateInvoice":
		return invCC.createInvoice(stub, args)
	case "GetInvoiceByID":
		return invCC.getInvoiceByID(stub, args)
	case "GetInvoiceByPartner":
		return invCC.getInvoiceByPartner(stub, args)
	case "QueryInvoiceByPartner":
		return invCC.queryInvoiceByPartner(stub, args)
	case "GetInvoiceLines":
		return invCC.getInvoiceLines(stub, args)
	}
	return shim.Error("Invalid Function Call")
}

func (invCC *InvoiceCC) createInvoice(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	trans, err := stub.GetTransient()
	if err != nil {
		return shim.Error(err.Error())
	}

	invLines, ok := trans["InvoiceLines"]
	if !ok {
		return shim.Error("Transient Invoice Lines Data Not Found")
	}
	invoiceLines := []*InvoiceLine{}
	err = json.Unmarshal(invLines, &invoiceLines)
	if err != nil {
		return shim.Error(err.Error())
	}

	DocType := "invoice"
	InvID := stub.GetTxID()
	InvNumber := args[0]
	InvDate, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		return shim.Error(err.Error())
	}
	PartnerID, found, err := cid.GetAttributeValue(stub, "hf.EnrollmentID")
	if err != nil {
		return shim.Error(err.Error())
	}
	if !found {
		return shim.Error("Invalid Cert")
	}
	PartnerOrg, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	CounterPartnerID := args[2]
	CounterPartnerOrg := args[3]
	Subtotal, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	VAT, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	State := OPEN

	invoice := Invoice{DocType, InvID, InvNumber, InvDate, PartnerID, PartnerOrg, CounterPartnerID, CounterPartnerOrg, Subtotal, VAT, State}
	data, err := json.Marshal(invoice)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(InvID, data)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Storing PrivateData
	for i, line := range invoiceLines {
		line.InvID = InvID
		line.DocType = "InvoiceLines"
		byteLine, err := json.Marshal(line)
		if err != nil {
			return shim.Error(err.Error())
		}

		collection := fmt.Sprintf("collection_%s", PartnerOrg)
		key := fmt.Sprintf("%s_%d", InvID, i)
		stub.PutPrivateData(collection, key, byteLine)
	}

	// Storing Composite Keys
	key, err := shim.CreateCompositeKey("invoice~partner", []string{PartnerID, InvID})
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(key, []byte{0x00})
	if err != nil {
		return shim.Error(err.Error())
	}
	key, err = shim.CreateCompositeKey("invoice~counterpartner", []string{CounterPartnerID, InvID})
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(key, []byte{0x00})
	if err != nil {
		return shim.Error(err.Error())
	}

	stub.SetEvent("InvoiceCreated", data)

	return shim.Success([]byte(InvID))
}

func (invCC *InvoiceCC) getInvoiceLines(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	orgID, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	query := `
	{
		"selector": {
			"invID" : "%s"
		}
	}
	`
	query = fmt.Sprintf(query, args[0])

	collection := fmt.Sprintf("collection_%s", orgID)
	it, err := stub.GetPrivateDataQueryResult(collection, query)
	if err != nil {
		return shim.Error(err.Error())
	}
	var res []*InvoiceLine
	for it.HasNext() {
		queryRes, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		var invoiceLine = new(InvoiceLine)
		err = json.Unmarshal(queryRes.Value, invoiceLine)
		if err != nil {
			return shim.Error(err.Error())
		}
		res = append(res, invoiceLine)
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(resBytes)
}

func (invCC *InvoiceCC) getInvoiceByID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	data, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	if data == nil {
		return shim.Success([]byte("Data Not Found"))
	}

	// var invoice = new (InvoiceCC)
	return shim.Success(data)
}

func (invCC *InvoiceCC) getInvoiceByPartner(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	stateIt, err := stub.GetStateByPartialCompositeKey("invoice~partner", []string{args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer stateIt.Close()
	var res []*Invoice

	for stateIt.HasNext() {
		queryRes, err := stateIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		_, composite, err := stub.SplitCompositeKey(queryRes.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		invID := composite[1]
		data, err := stub.GetState(invID)
		if err != nil {
			return shim.Error(err.Error())
		}
		var invoice = new(Invoice)
		_ = json.Unmarshal(data, invoice)
		res = append(res, invoice)
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(resBytes)
}

func (invCC *InvoiceCC) queryInvoiceByPartner(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	query := `
	{
		"selector": {
			"partnerID" : "%s",
			"docType": "invoice"
		}
	}
	`
	query = fmt.Sprintf(query, args[0])
	it, metadata, err := stub.GetQueryResultWithPagination(query, 10, args[1])
	var Payload []*Invoice

	for it.HasNext() {
		kv, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		var invoice = new(Invoice)
		err = json.Unmarshal(kv.Value, invoice)
		if err != nil {
			return shim.Error(err.Error())
		}
		Payload = append(Payload, invoice)
	}
	result := InvoiceQueryResult{
		metadata.FetchedRecordsCount,
		Payload,
		metadata.Bookmark,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(data)
}

func main() {
	invCC := new(InvoiceCC)
	if err := shim.Start(invCC); err != nil {
		panic(err)
	}
}

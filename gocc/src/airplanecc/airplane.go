package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type AirplaneCC struct {
}

type Airplane struct {
	SerialNumber         string    `json:"serialNumber"`
	BatchProduction      string    `json:"batchProduction"`
	DateProduction       time.Time `json:"dateProduction"`
	AirplaneType         string    `json:"airplaneType"`
	DateAcquisition      time.Time `json:"dateAquisition"`
	FactoryOrg           string    `json:"factoryOrg"`
	OwnerOrg             string    `json:"ownerOrg"`
	UserOrg              string    `json:"userOrg"`
	NumberEngine         int       `json:"numberEngine"`
	SNEngine             string    `json:"snEngine"`
	MaxCapacityPassenger int       `json:"maxCapacityPassenger"`
	MaxWeightKg          float64   `json:"maxWeightKg"`
	NetWeightKg          float64   `json:"netWeightKg"`
}

type AirplanePagination struct {
	FetchRecord int32                `json:"fetchRecord"`
	Data        map[string]*Airplane `json:"data"`
	Bookmark    string               `json:"string"`
}

func (airCC *AirplaneCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	err := stub.PutState("perm_write", []byte("org1"))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Initiated"))
}

func (airCC *AirplaneCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function_name, args := stub.GetFunctionAndParameters()

	switch function_name {
	case "CreatePlane":
		return airCC.createPlane(stub, args)
	case "ChangeOwnership":
		return airCC.changeOwnership(stub, args)
	case "QueryBySN":
		return airCC.queryBySN(stub, args)
	case "ConfirmOwnership":
		return airCC.confirmOwnership(stub, args)
	case "QueryPlaneProduction":
		return airCC.queryPlaneProduction(stub, args)
	case "QueryPlaneProductionPagination":
		return airCC.queryPlaneProductionPagination(stub, args)
	}

	return shim.Error("Invalid Function Call")
}

func (airCC *AirplaneCC) checkWritePerm(stub shim.ChaincodeStubInterface) error {

	// Get Auth from db
	data, err := stub.GetState("perm_write")
	if err != nil {
		return err
	}
	orgs := strings.Split(string(data[:]), ",")

	val, err := GetOrg(stub)
	if err != nil {
		return err
	}
	for _, org := range orgs {
		if org == val {
			return nil
		}
	}
	return fmt.Errorf("Access Error")

}

func (airCC *AirplaneCC) createPlane(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if err := airCC.checkWritePerm(stub); err != nil {
		return shim.Error(err.Error())
	}

	if len(args) != 9 {
		return shim.Error("Invalid length parameter")
	}

	serialNumber := args[0]
	batchProduction := args[1]
	dateProduction, err := time.Parse("2006-01-02", args[2])
	if err != nil {
		return shim.Error(err.Error())
	}
	airplaneType := args[3]
	numberEngine, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("NumberEngine field Need integer")
	}
	snEngine := args[5]
	maxCapacityPassenger, err := strconv.Atoi(args[6])
	if err != nil {
		return shim.Error("Max Capacity Passenger Need Integer")
	}
	maxWeightKg, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error("Max Weight need Float")
	}
	netWeightKg, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		return shim.Error("Net Weight need Float")
	}

	// Get Sender ID
	id, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	org, _, _ := cid.GetAttributeValue(stub, "hf.Affiliation")

	airplane := Airplane{
		SerialNumber:         serialNumber,
		BatchProduction:      batchProduction,
		DateProduction:       dateProduction,
		AirplaneType:         airplaneType,
		UserOrg:              id,
		OwnerOrg:             org,
		FactoryOrg:           org,
		DateAcquisition:      dateProduction,
		NumberEngine:         numberEngine,
		SNEngine:             snEngine,
		MaxCapacityPassenger: maxCapacityPassenger,
		MaxWeightKg:          maxWeightKg,
		NetWeightKg:          netWeightKg,
	}

	data, _ := json.Marshal(airplane)
	key := serialNumber
	err = stub.PutState(key, data)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(key))
}

func (airCC *AirplaneCC) queryBySN(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Need parameter serial number")
	}

	data, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	if data == nil {
		return shim.Error(args[0] + " does not exists")
	}

	return shim.Success([]byte(data))
}

func (airCC *AirplaneCC) changeOwnership(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 2 {
		return shim.Error("Invalid paramter length ")
	}

	// Get Sender Org
	org, err := GetOrg(stub)
	if err != nil {
		shim.Error(err.Error())
	}

	response := airCC.queryBySN(stub, []string{args[0]})
	if response.Status != shim.OK {
		return shim.Error(response.Message)
	}

	airplane := new(Airplane)
	err = json.Unmarshal(response.Payload, airplane)
	if err != nil {
		return shim.Error(err.Error())
	}

	if airplane.OwnerOrg != org {
		return shim.Error("Invalid Ownership")
	}

	airplane.OwnerOrg = args[1]
	airplane.UserOrg = ""

	data, _ := json.Marshal(airplane)

	err = stub.PutState(args[0], data)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(args[0] + " Ownership has changed"))
}

func (airCC *AirplaneCC) queryPlaneProduction(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	query := `{ "selector":  {
					"factoryOrg": "%s",
					"dateProduction": {
						"$gte": "%s",
						"$lte": "%s"
					}
				}
		}
	`
	query = fmt.Sprintf(query, args[0], args[1], args[2])
	fmt.Print(query)

	queryIt, err := stub.GetQueryResult(query)
	if err != nil {
		return shim.Error(err.Error())
	}

	defer queryIt.Close()

	airplanes := make(map[string]*Airplane)

	for queryIt.HasNext() {
		result, err := queryIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		airplane := new(Airplane)
		err = json.Unmarshal(result.GetValue(), airplane)
		if err != nil {
			return shim.Error(err.Error())
		}

		airplanes[result.GetKey()] = airplane
	}

	data, err := json.Marshal(airplanes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(data)
}

func (airCC *AirplaneCC) queryPlaneProductionPagination(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	query := `{ "selector":  {
					"factoryOrg": "%s",
					"dateProduction": {
						"$gte": "%s",
						"$lte": "%s"
					}
				}
		}
	`
	query = fmt.Sprintf(query, args[0], args[1], args[2])
	fmt.Print(query)

	pageSize, err := strconv.ParseInt(args[3], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}

	queryIt, metaInfo, err := stub.GetQueryResultWithPagination(query, int32(pageSize), args[4])
	if err != nil {
		return shim.Error(err.Error())
	}

	defer queryIt.Close()

	airplanes := make(map[string]*Airplane)

	for queryIt.HasNext() {
		result, err := queryIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		airplane := new(Airplane)
		err = json.Unmarshal(result.GetValue(), airplane)
		if err != nil {
			return shim.Error(err.Error())
		}

		airplanes[result.GetKey()] = airplane
	}

	res := AirplanePagination{
		FetchRecord: metaInfo.FetchedRecordsCount,
		Data:        airplanes,
		Bookmark:    metaInfo.Bookmark,
	}

	data, err := json.Marshal(res)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(data)
}

func (airCC *AirplaneCC) confirmOwnership(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Invalid parameter length")
	}

	id, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	org, err := GetOrg(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	resp := airCC.queryBySN(stub, []string{args[0]})
	if resp.Status != shim.OK {
		return shim.Error(resp.Message)
	}

	airplane := new(Airplane)
	_ = json.Unmarshal(resp.Payload, airplane)

	if airplane.OwnerOrg != org {
		return shim.Error("Invalid Ownership")
	}
	airplane.UserOrg = id

	data, err := json.Marshal(airplane)
	if err != nil {
		return shim.Error(err.Error())
	}

	if err = stub.PutState(args[0], data); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(args[0] + " Ownership Confirmed"))
}

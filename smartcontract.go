/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
)

type serverConfig struct {
	CCID    string
	Address string
}

type SmartContract struct {
	contractapi.Contract
}

type Tender struct {
	Id          string `json:"Id"`
	TenderID    string `json:"TenderID"`
	Accountcode string `json:"Accountcode"`
	Account     string `json:"Account"`
	Name        string `json:"Name"`
	Currency    string `json:"Currency"`
	Branch      string `json:"Branch"`
	Amount      string `json:"Amount"`
	Status      string `json:"Status"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Tender
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	var inits []Tender

	for _, tender := range inits {
		dataJSON, err := json.Marshal(tender)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(tender.Id, dataJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}
	fmt.Println("Initiation Tender-Chain v1.0 chaincode Success")
	return nil
}

func (s *SmartContract) CreateData(ctx contractapi.TransactionContextInterface, id, tenderid, accountCode, account, name, currency, branch, amount, status string) error {
	log.Println("===========CreateData==============")

	log.Println("===========DataExists Start==============")
	exists, err := s.DataExists(ctx, id)
	log.Println(exists)
	log.Println("===========If err Run err==============")
	if err != nil {
		return err
	}
	log.Println("===========If exists Run err==============")
	if exists {
		return fmt.Errorf("file %s already exists", id)
	}
	tender := Tender{
		Id:          id,
		TenderID:    tenderid,
		Accountcode: accountCode,
		Account:     account,
		Name:        name,
		Currency:    currency,
		Branch:      branch,
		Amount:      amount,
		Status:      status,
	}
	log.Println("===========json.Marshal Start==============")
	dataJSON, err := json.Marshal(tender)
	log.Println("===========If err Run err==============")
	if err != nil {
		return err
	}
	log.Println("===========Finish==============")
	return ctx.GetStub().PutState(id, dataJSON)
}

func (s *SmartContract) ReadData(ctx contractapi.TransactionContextInterface, id string) (string, error) {

	//此方法是讀出 Read Txid，所以每次查都不一樣
	// transaction := ctx.GetStub().GetTxID()
	// log.Println("txid: " + transaction)

	var transaction string
	dataJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return "nil", fmt.Errorf("failed to read from tender. %s", err.Error())
	}
	if dataJSON == nil {
		return "nil", fmt.Errorf("file %s does not exist", id)
	}
	myString := string(dataJSON[:])

	// 為了取出 Txid
	//---------------------------
	historyQueryIterator, err := ctx.GetStub().GetHistoryForKey(id)

	// In case of error - return error
	if err != nil {
		return "MUST provide time !!!", err
	}
	defer historyQueryIterator.Close()

	var resultModification *queryresult.KeyModification

	for historyQueryIterator.HasNext() {

		// Get the next record
		resultModification, err = historyQueryIterator.Next()

		if err != nil {
			return "Error in reading history record!!!", err
		}
		transaction = resultModification.GetTxId()
	}
	//---------------------------

	resultJSON := "{ \"txid\": " + "\"" + transaction + "\"" + ", \"data\":" + myString + "}"
	// var tender Tender
	// err = json.Unmarshal(resultJSON, &tender)
	// if err != nil {
	// 	return nil, err
	// }

	return resultJSON, nil

}

func (s *SmartContract) UpdateData(ctx contractapi.TransactionContextInterface, id, tenderid, accountCode, account, name, currency, branch, amount, status string) error {
	log.Println("===========UpdateData==============")
	exists, err := s.DataExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("file %s does not exist", id)
	}

	tender := Tender{
		Id:          id,
		TenderID:    tenderid,
		Accountcode: accountCode,
		Account:     account,
		Name:        name,
		Currency:    currency,
		Branch:      branch,
		Amount:      amount,
		Status:      status,
	}

	dataJSON, err := json.Marshal(tender)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, dataJSON)
}

func (s *SmartContract) DeleteData(ctx contractapi.TransactionContextInterface, id string) error {
	log.Println("===========DeleteData==============")
	exists, err := s.DataExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("file %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) DataExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	log.Println("===========DataExists==============")
	dataJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from chain. %s", err.Error())
	}

	return dataJSON != nil, nil
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (string, error) {
	log.Println("===========constructQueryResponseFromIterator Start==============")
	var tenders []*Tender
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "nil", err
		}
		var tender Tender
		err = json.Unmarshal(queryResult.Value, &tender)
		if err != nil {
			return "nil", err
		}
		tenders = append(tenders, &tender)
	}
	jsondata, _ := json.Marshal(tenders)
	log.Println(string(jsondata))
	log.Println(reflect.TypeOf(string(jsondata)))

	return string(jsondata), nil
}

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) (string, error) {
	log.Println("===========getQueryResultForQueryString Start==============")
	log.Println(queryString)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return "nil", err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func (s *SmartContract) GetAllData(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	log.Println("===========GetAllData==============")
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []QueryResult

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var tender Tender
		err = json.Unmarshal(queryResponse.Value, &tender)
		if err != nil {
			return nil, err
		}

		queryResult := QueryResult{Key: queryResponse.Key, Record: &tender}
		results = append(results, queryResult)
	}

	return results, nil
}

func (s *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, hash string) (string, error) {
	log.Println("===========GetHistory==============")

	// Get the history for the key i.e., VIN#
	historyQueryIterator, err := ctx.GetStub().GetHistoryForKey(hash)

	// In case of error - return error
	if err != nil {
		return "MUST provide time !!!", err
	}
	defer historyQueryIterator.Close()

	// Local variable to hold the history record
	var resultModification *queryresult.KeyModification
	counter := 0
	resultJSON := "["

	// Start a loop with check for more rows
	for historyQueryIterator.HasNext() {

		// Get the next record
		resultModification, err = historyQueryIterator.Next()

		if err != nil {
			return "Error in reading history record!!!", err
		}

		// Append the data to local variable
		data := "{\"txn\":" + resultModification.GetTxId()
		data += " , \"value\": " + string(resultModification.GetValue()) + "}  "
		if counter > 0 {
			data = ", " + data
		}
		resultJSON += data

		counter++

	}

	// Close the iterator
	historyQueryIterator.Close()

	// finalize the return string
	resultJSON += "]"
	resultJSON = "{ \"counter\": " + strconv.Itoa(counter) + ", \"txns\":" + resultJSON + "}"

	// return success
	return resultJSON, nil
}

func main() {
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	chaincode, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("error create Tender chaincode: %s", err)
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		log.Panicf("error starting Tender chaincode: %s", err)
	}
}

/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var entityIndexStr = "_entityindex" //name for the key/value that will store a list of all known marbles

// Entity
type Entity struct {
	Name   string  `json:"name"` //the fieldtags are needed to keep case from bouncing around
	Role   string  `json:"role"`
	TxnBal float64 `json:"txnbal"`
	PtBal  float64 `json:"ptbal"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval))) //making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty) //marshal an emtpy array of strings to clear the index
	err = stub.PutState(entityIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke a transaction
func (t *SimpleChaincode) transfer(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var from, to string
	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	from = args[0]
	to = args[1]

	fromAsbytes, err := stub.GetState(from)
	if err != nil {
		return nil, err
	}
	toAsbytes, err := stub.GetState(to)
	if err != nil {
		return nil, err
	}

	fromEntity := Entity{}
	json.Unmarshal(fromAsbytes, &fromEntity)

	toEntity := Entity{}
	json.Unmarshal(toAsbytes, &toEntity)

	txnAmt, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, err
	}
	rdAmt, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return nil, err
	}

	fromEntity.TxnBal = fromEntity.TxnBal - txnAmt
	toEntity.TxnBal = toEntity.TxnBal + txnAmt

	toEntity.PtBal = toEntity.PtBal + rdAmt
	fromEntity.PtBal = fromEntity.PtBal - rdAmt

	jsonAsBytes, _ := json.Marshal(fromEntity) //save new index
	err = stub.PutState(fromEntity.Name, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	jsonAsBytes, _ = json.Marshal(toEntity) //save new index
	err = stub.PutState(toEntity.Name, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	return nil, nil

}

// Invoke
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	if function == "transfer" { //read a variable
		return t.transfer(stub, args)
	} else if function == "create_entity" {
		return t.initEntity(stub, args)
	}
	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")

}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil //send it onward
}

// ============================================================================================================================
// Init Entity - create a new entity, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) initEntity(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//   0       1       2        3
	// "Name", "Role", "TxnBal", "PtBal"
	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}

	fmt.Println("- start init entity")
	if len(args[0]) <= 0 {
		fmt.Println("1st argument must be a non-empty string")
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		fmt.Println("2nd argument must be a non-empty string")
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		fmt.Println("3rd argument must be a non-empty string")
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		fmt.Println("4th argument must be a non-empty string")
		return nil, errors.New("4th argument must be a non-empty string")
	}

	txnbal, err := strconv.ParseFloat(args[2], 64)
	fmt.Println(txnbal)
	if (err != nil) || (txnbal < 0) {
		fmt.Println("3rd argument must be a numeric string")
		return nil, errors.New("3rd argument must be a numeric string")
	}

	ptbal, err := strconv.ParseFloat(args[3], 64)
	if (err != nil) || (ptbal < 0) {
		fmt.Println("4th argument must be a numeric string")
		return nil, errors.New("4th argument must be a numeric string")
	}

	str := `{"name": "` + args[0] + `", "role": "` + args[1] + `", "txnbal": ` + args[2] + `, "ptbal": "` + args[3] + `"}`
	err = stub.PutState(args[0], []byte(str)) //store marble with id as key
	if err != nil {
		fmt.Println("Writing failed")
		return nil, err
	}

	//get the entity index
	entityAsBytes, err := stub.GetState(entityIndexStr)
	if err != nil {
		fmt.Println("Failed to get entity index")
		return nil, errors.New("Failed to get entity index")
	}
	var entityIndex []string
	json.Unmarshal(entityAsBytes, &entityIndex) //un stringify it aka JSON.parse()

	//append
	entityIndex = append(entityIndex, args[0]) //add entity name to index list
	fmt.Println("! entity index: ", entityIndex)
	jsonAsBytes, _ := json.Marshal(entityIndex)
	err = stub.PutState(entityIndexStr, jsonAsBytes) //store name of entity
	if err != nil {
		fmt.Println("Failed to write")
		return nil, errors.New("Failed to write")
	}
	fmt.Println("- end init entity")
	return nil, nil
}

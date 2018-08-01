package main

import (
    "fmt"
    "strings"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "github.com/hyperledger/fabric/core/chaincode/shim/ext/entities"
    "github.com/hyperledger/fabric/protos/peer"
    "github.com/hyperledger/fabric/bccsp"
    "github.com/hyperledger/fabric/bccsp/factory"


    
)
const DECKEY = "DECKEY"
const ENCKEY = "ENCKEY"
const IV = "IV"

// cvChain
type cvChain struct {
    bccspInst bccsp.BCCSP
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *cvChain) Init(stub shim.ChaincodeStubInterface) peer.Response {
    // Get the args from the transaction proposal
    args := stub.GetStringArgs()
    if len(args) != 4 {
        return shim.Success(nil)
    }

    // Set up any variables or assets here by calling stub.PutState()

    // We store the key and the value on the ledger
    err := stub.PutState(args[0]+";"+args[1], []byte(strings.Join(args[2:],";")))
    if err != nil {
        return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
    }
    return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *cvChain) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
    // Extract the function and args from the transaction proposal
    fn, args := stub.GetFunctionAndParameters()
    tMap, _ := stub.GetTransient()
    var result string
    var err error
    if tMap==nil{
        if fn == "addRecord" {
            result, err = addRecord(stub, args)
            if err != nil {
                return shim.Error(err.Error())
            }else{
                return shim.Success([]byte(result))
            }
        } else if fn == "getRecord" {
            result, err = getRecord(stub, args)
            if err != nil {
                return shim.Error(err.Error())
            }else{
                return shim.Success([]byte(result))
            }
        }else{
            return shim.Error("Wrong instructions1")
        }

    }else{
        if fn =="encRecord" {
            // make sure there's a key in transient - the assumption is that
            // it's associated to the string "ENCKEY"
            if _, in := tMap[ENCKEY]; !in {
                return shim.Error(fmt.Sprintf("Expected transient encryption key %s", ENCKEY))
            }

            return t.Encrypter(stub, args[0:], tMap[ENCKEY], tMap[IV])
        }else if fn== "decRecord"{
            // make sure there's a key in transient - the assumption is that
            // it's associated to the string "DECKEY"
            if _, in := tMap[DECKEY]; !in {
                return shim.Error(fmt.Sprintf("Expected transient decryption key %s", DECKEY))
            }

            return t.Decrypter(stub, args[0:], tMap[DECKEY], tMap[IV])
        }else{
            return shim.Error("Wrong instructions2")
        }


    }

    // Return the result as success payload
    return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func addRecord(stub shim.ChaincodeStubInterface, args []string) (string, error) {
    if len(args) != 4 {
        return "", fmt.Errorf("Incorrect arguments. Expecting a key and 3 values")
    }
    //Firstly, check if ID has existed
    err := stub.PutState(args[0]+";"+args[1], []byte(strings.Join(args[2:],";")))

    if err != nil {
        return "", fmt.Errorf("Failed to add record: %s", args[0])
    }
    return args[0], nil
}

// Get returns the value of the specified asset key
func getRecord(stub shim.ChaincodeStubInterface, args []string) (string, error) {
    if len(args) != 2 {
        return "", fmt.Errorf("Incorrect arguments. Expecting a key ID and YEAR")
    }

    value, err := stub.GetState(args[0]+";"+args[1])
    if err != nil {
        return "", fmt.Errorf("Failed to get record: %s with error: %s", args[0], err)
    }
    if value == nil {
        return "", fmt.Errorf("No record: %s", args[0])
    }

    return strings.Split(string(value),";")[0],nil


}

func (t *cvChain) Encrypter(stub shim.ChaincodeStubInterface, args []string, encKey, IV []byte) peer.Response {
    // create the encrypter entity - we give it an ID, the bccsp instance, the key and (optionally) the IV
    ent, err := entities.NewAES256EncrypterEntity("ID", t.bccspInst, encKey, IV)
    if err != nil {
        return shim.Error(fmt.Sprintf("entities.NewAES256EncrypterEntity failed, err %s", err))
    }

    if len(args) != 4 {
        return shim.Error("Incorrect arguments. Expecting a key and 3 values")
    }

    key := args[0]+";"+args[1]
    cleartextValue := []byte(strings.Join(args[2:],";"))

    // here, we encrypt cleartextValue and assign it to key
    err = encryptAndPutState(stub, ent, key, cleartextValue)
    if err != nil {
        return shim.Error(fmt.Sprintf("encryptAndPutState failed, err %+v", err))
    }
    return shim.Success(nil)
}

func (t *cvChain) Decrypter(stub shim.ChaincodeStubInterface, args []string, decKey, IV []byte) peer.Response {
    // create the encrypter entity - we give it an ID, the bccsp instance, the key and (optionally) the IV
    ent, err := entities.NewAES256EncrypterEntity("ID", t.bccspInst, decKey, IV)
    if err != nil {
        return shim.Error(fmt.Sprintf("entities.NewAES256EncrypterEntity failed, err %s", err))
    }

    if len(args) != 2 {
        return shim.Error("Incorrect arguments. Expecting a key ID and YEAR")
    }

    key := args[0]+";"+args[1]

    // here we decrypt the state associated to key
    cleartextValue, err := getStateAndDecrypt(stub, ent, key)
    if err != nil {
        return shim.Error(fmt.Sprintf("getStateAndDecrypt failed, err %+v", err))
    }
    if cleartextValue==nil{
        return shim.Error("No Record")
    }

    return shim.Success([]byte(strings.Split(string(cleartextValue),";")[0]))



}
func encryptAndPutState(stub shim.ChaincodeStubInterface, ent entities.Encrypter, key string, value []byte) error {
    // at first we use the supplied entity to encrypt the value
    ciphertext, err := ent.Encrypt(value)
    if err != nil {
        return err
    }

    return stub.PutState(key, ciphertext)
}
func getStateAndDecrypt(stub shim.ChaincodeStubInterface, ent entities.Encrypter, key string) ([]byte, error) {
    // at first we retrieve the ciphertext from the ledger
    ciphertext, err := stub.GetState(key)
    if err != nil {
        return nil, err
    }

    // GetState will return a nil slice if the key does not exist.
    // Note that the chaincode logic may want to distinguish between
    // nil slice (key doesn't exist in state db) and empty slice
    // (key found in state db but value is empty). We do not
    // distinguish the case here
    if len(ciphertext) == 0 {
        return nil, nil
    }

    return ent.Decrypt(ciphertext)
}


// main function starts up the chaincode in the container during instantiate
func main() {
    factory.InitFactories(nil)
    err := shim.Start(&cvChain{factory.GetDefault()})
    if err != nil {
        fmt.Printf("Error starting cvChain chaincode: %s", err)
    }
}
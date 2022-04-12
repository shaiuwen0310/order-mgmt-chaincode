package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

type serverConfig struct {
	CCID    string
	Address string
}

type storeOrderInfo struct {
	Txid      string `json:"txid"`      // 交易代號
	Hash      string `json:"hash"`      // 單據/文件/字段+所有附件檔案進行壓縮並提取其hash code
	Identity  string `json:"identity"`  // 身份/系統名稱(倍智模房網,倍智模管家,迪維歐六臂雲)
	GroupId   string `json:"groupid"`   // 集團ID
	CompanyId string `json:"companyid"` // 公司ID
	UserId    string `json:"userid"`    // 前端的使用者ID/用戶名
	FormSeqNo string `json:"formseqno"` // 單據編號
	FromHash  string `json:"fromhash"`  // 將所有的來源資料進行壓縮並提取其hash code
	Time      string `json:"time"`      // 秒數，Unix timestamp
	Type      int64  `json:"type"`      // 上鏈數字的型態 0:file, 1:string
	Deptid    string `json:"deptid"`    // 部門/組織ID
}

const (
	MESSAGE_0   string = "success"                                              // 應用: 成功
	MESSAGE_101 string = "Incorrect number of parameters"                       // 應用: 參數數量有誤
	MESSAGE_102 string = "An exception occurred while accessing data"           // 系統: 存取資料時發生異常(使用fabric shim sdk時)
	MESSAGE_103 string = "This key value already exists"                        // 應用: uniqueKey重複，表示檔案已經儲存過
	MESSAGE_104 string = "This key value has not been stored in the blockchain" // 應用: 無此uniqueKey，表示尚未將此檔案儲存到區塊鏈中
	MESSAGE_105 string = "The type field is greater than 1"                     // 應用: type欄位不正確
	MESSAGE_106 string = "The identity field is different"                      // 應用: identity跟原本記錄的不同
	MESSAGE_107 string = "The groupId field is different"                       // 應用: groupId跟原本記錄的不同
	MESSAGE_108 string = "The companyId field is different"                     // 應用: companyId跟原本記錄的不同
	MESSAGE_109 string = "The formSeqNo field is different"                     // 應用: formSeqNo跟原本記錄的不同
)

const (
	RTNC_0   int64 = 0
	RTNC_101 int64 = 101
	RTNC_102 int64 = 102
	RTNC_103 int64 = 103
	RTNC_104 int64 = 104
	RTNC_105 int64 = 105
	RTNC_106 int64 = 106
	RTNC_107 int64 = 107
	RTNC_108 int64 = 108
	RTNC_109 int64 = 109
)

type SmartContract struct{}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()

	if function == "set" {
		return s.Set(APIstub, args)
	} else if function == "get" {
		return s.Get(APIstub, args)
	} else if function == "gethist" {
		return s.getHistory(APIstub, args)
	} else if function == "modify" {
		return s.modify(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name: " + function)
}

func (s *SmartContract) Set(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 回應的json結構
	var setResp = make(map[string]interface{})

	if len(args) != 10 {
		setResp["rtnc"] = RTNC_101
		setResp["message"] = MESSAGE_101
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	}

	//輸入參數
	//uniqueKey為key值，其他為value
	uniqueKey := args[0]
	userId := args[1]
	fromHash := args[2]
	hash := args[3]
	identity := args[4]
	groupId := args[5]
	companyId := args[6]
	formSeqNo := args[7]
	typeStr := args[8]
	deptId := args[9]

	txnTime := strconv.Itoa(int(time.Now().Unix()))
	txid := APIstub.GetTxID()

	setResp["uniqueKey"] = uniqueKey

	// 將type轉換為int64
	flag, _ := strconv.ParseInt(typeStr, 10, 64)
	// 目前沒有大於1的type
	if flag > 1 {
		setResp["rtnc"] = RTNC_105
		setResp["message"] = MESSAGE_105
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	}

	//檢查是否有這個hash值
	hashKeyAsBytes, err := APIstub.GetState(uniqueKey)
	if err != nil {
		setResp["rtnc"] = RTNC_102
		setResp["message"] = MESSAGE_102
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	} else if hashKeyAsBytes != nil {
		// key值重複，表示此key值已存在
		setResp["rtnc"] = RTNC_103
		setResp["message"] = MESSAGE_103
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	}

	// 將檔案資訊全部儲存在storeOrderInfo結構中，轉成json字串，最後將hash當成key值儲存在區塊鏈中
	var storeInfo = storeOrderInfo{Txid: txid, Hash: hash, Identity: identity, GroupId: groupId, CompanyId: companyId, UserId: userId, FormSeqNo: formSeqNo, FromHash: fromHash, Time: txnTime, Type: flag, Deptid: deptId}
	storeInfoAsBytes, _ := json.Marshal(storeInfo)
	APIstub.PutState(uniqueKey, storeInfoAsBytes)

	// 回應的json結構
	setResp["rtnc"] = RTNC_0
	setResp["message"] = MESSAGE_0
	setResp["txid"] = txid
	setResp["time"] = txnTime
	mapResp, _ := json.Marshal(setResp)
	return shim.Success(mapResp)

}

func (s *SmartContract) Get(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	var vResp = make(map[string]interface{})

	if len(args) != 1 {
		vResp["rtnc"] = RTNC_101
		vResp["message"] = MESSAGE_101
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	//輸入參數
	uniqueKey := args[0]

	vResp["uniqueKey"] = uniqueKey

	//檢核key值是否存在
	keyAsBytes, err := APIstub.GetState(uniqueKey)
	var dbstoreOrderInfo storeOrderInfo
	json.Unmarshal(keyAsBytes, &dbstoreOrderInfo) //反序列化

	if err != nil {
		vResp["rtnc"] = RTNC_102
		vResp["message"] = MESSAGE_102
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	} else if keyAsBytes == nil {
		// 查無此key，表示尚未將此檔案儲存到區塊鏈中
		vResp["rtnc"] = RTNC_104
		vResp["message"] = MESSAGE_104
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	// 表示查詢到檔案，做json格式回應
	vResp["rtnc"] = RTNC_0
	vResp["message"] = MESSAGE_0
	vResp["txid"] = dbstoreOrderInfo.Txid
	vResp["hash"] = dbstoreOrderInfo.Hash
	vResp["groupId"] = dbstoreOrderInfo.GroupId
	vResp["companyId"] = dbstoreOrderInfo.CompanyId
	vResp["userId"] = dbstoreOrderInfo.UserId
	vResp["formSeqNo"] = dbstoreOrderInfo.FormSeqNo
	vResp["fromHash"] = dbstoreOrderInfo.FromHash
	vResp["time"] = dbstoreOrderInfo.Time
	vResp["type"] = dbstoreOrderInfo.Type
	vResp["deptId"] = dbstoreOrderInfo.Deptid
	mapResp, _ := json.Marshal(vResp)
	return shim.Success(mapResp)

}

func (s *SmartContract) getHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// getHistory中，錯誤用map回，正確用buffer組成

	// 回應的json結構
	var vResp = make(map[string]interface{})

	if len(args) != 1 {
		vResp["rtnc"] = RTNC_101
		vResp["message"] = MESSAGE_101
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	//輸入參數
	uniqueKey := args[0]

	vResp["uniqueKey"] = uniqueKey

	resultsIterator, err := APIstub.GetHistoryForKey(uniqueKey)

	if err != nil {
		vResp["rtnc"] = RTNC_102
		vResp["message"] = MESSAGE_102
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}
	defer resultsIterator.Close()

	var jsonArr []interface{}
	for resultsIterator.HasNext() {
		// response包含所有資訊，response.Value為儲存的數值
		response, err := resultsIterator.Next()
		if err != nil {
			vResp["rtnc"] = RTNC_102
			vResp["message"] = MESSAGE_102
			mapResp, _ := json.Marshal(vResp)
			return shim.Success(mapResp)
		}

		valueStr := []byte(response.Value)
		valueStruct := storeOrderInfo{}
		json.Unmarshal(valueStr, &valueStruct)
		jsonArr = append(jsonArr, valueStruct)

	}

	if jsonArr == nil {
		vResp["rtnc"] = RTNC_104
		vResp["message"] = MESSAGE_104
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	vResp["rtnc"] = RTNC_0
	vResp["message"] = MESSAGE_0
	vResp["info"] = jsonArr
	mapResp, _ := json.Marshal(vResp)
	return shim.Success(mapResp)

}

func (s *SmartContract) modify(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	var vResp = make(map[string]interface{})

	if len(args) != 10 {
		vResp["rtnc"] = RTNC_101
		vResp["message"] = MESSAGE_101
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	//輸入參數
	uniqueKey := args[0]
	vResp["uniqueKey"] = uniqueKey

	storeorderinfoAsBytes, err := APIstub.GetState(uniqueKey)

	if err != nil {
		vResp["rtnc"] = RTNC_102
		vResp["message"] = MESSAGE_102
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	} else if storeorderinfoAsBytes == nil {
		// 無此uniqueKey
		vResp["rtnc"] = RTNC_104
		vResp["message"] = MESSAGE_104
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	var dbstoreOrderInfo storeOrderInfo
	json.Unmarshal(storeorderinfoAsBytes, &dbstoreOrderInfo)

	// `已儲存的值`與`前端帶的值`進行比較
	// 目前規定Identity GroupId CompanyId FormSeqNo要一致
	if dbstoreOrderInfo.Identity != args[4] {
		vResp["rtnc"] = RTNC_106
		vResp["message"] = MESSAGE_106
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	} else if dbstoreOrderInfo.GroupId != args[5] {
		vResp["rtnc"] = RTNC_107
		vResp["message"] = MESSAGE_107
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	} else if dbstoreOrderInfo.CompanyId != args[6] {
		vResp["rtnc"] = RTNC_108
		vResp["message"] = MESSAGE_108
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	} else if dbstoreOrderInfo.FormSeqNo != args[7] {
		vResp["rtnc"] = RTNC_109
		vResp["message"] = MESSAGE_109
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	storeorderinfo := storeOrderInfo{}
	json.Unmarshal(storeorderinfoAsBytes, &storeorderinfo)

	storeorderinfo.UserId = args[1]
	storeorderinfo.FromHash = args[2]
	storeorderinfo.Hash = args[3]

	// 將type欄位轉換為int64
	flag, _ := strconv.ParseInt(args[8], 10, 64)
	// 目前沒有大於1的type
	if flag > 1 {
		vResp["rtnc"] = RTNC_105
		vResp["message"] = MESSAGE_105
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}
	storeorderinfo.Type = flag
	storeorderinfo.Deptid = args[9]

	time1 := strconv.Itoa(int(time.Now().Unix()))
	storeorderinfo.Time = time1
	storeorderinfo.Txid = APIstub.GetTxID()

	storeorderinfoAsBytes, _ = json.Marshal(storeorderinfo)
	APIstub.PutState(uniqueKey, storeorderinfoAsBytes)
	json.Unmarshal(storeorderinfoAsBytes, &storeorderinfo)
	vResp["rtnc"] = RTNC_0
	vResp["message"] = MESSAGE_0
	vResp["uniqueKey"] = uniqueKey
	vResp["txid"] = storeorderinfo.Txid
	vResp["time"] = storeorderinfo.Time
	mapResp, _ := json.Marshal(vResp)
	return shim.Success(mapResp)
}

// func main() {
// 	err := shim.Start(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error creating new Smart Contract: %s", err)
// 	}
// }

func main() {
	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      new(SmartContract),
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		fmt.Printf("error starting orderform chaincode: %s", err)
	}
}

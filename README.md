# order-mgmt 外部鏈碼服務
### 概況
調整合約回應代碼及訊息，刪除delete功能，調整getHistory寫法，修改main()為外部鏈碼之啟動方式。
### 產生go.mod go.sum的方式
使用orderform.go為範例
* 程式碼內添加所需之import
* go mod init chaincode-orderform
* go run orderform.go，自動產生go.mod內容。因為沒添加環境變數，執行失敗為正確狀況。
### connection-config/
* 根據參數不同要進行調整
* 但安裝config在peer上時，是透過這些config產生hash值，所有peer安裝的config要先一致，安裝上去後再進去peer內修改成正確參數


package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type KeyDistributeReq struct {
	KeyValue      string
	Version       string
	EffectiveDate string
}

type KeyMsgHeader struct {
	transactionID string `xml:"name,attr"`
	timeStamp     string `xml:"age,attr"`
	accessToken   string `xml:"career"`
	commandType   string `xml:"interests"`
}

type KeyMsgBody struct {
	KeyReqs []KeyDistributeReq `xml:"KeyDistributeReq"`
}

type KeyMsg struct {
	XMLName xml.Name     `xml:"xml"`
	Header  KeyMsgHeader `xml:"header"`
	Body    KeyMsgBody   `xml:"body"`
}

/*
<?xml version="1.0" encoding="UTF-8"?>
<message>
    <header transactionID="100000001" timeStamp="2010-11-16T00:00:00.0Z" accessToken ="" commandType="KeyDistributeReq"/>
        <body>
           < KeyDistributeReq KeyValue="xxx" Version="xxxxxx" EffectiveDate="xxxxxxx"/>
           < KeyDistributeReq KeyValue="xxx" Version="xxxxxx" EffectiveDate="xxxxxxx"/>
        </body>
</message>
*/

func MakeKeyRequestBody(key KeyMsg) ([]byte, error) {
	return nil, nil
}

type KeyDistributeRes struct {
	KeyDistributeRes string
}

/*
<?xml version="1.0" encoding="UTF-8"?>
<message>
<header transactionID="100000001" timeStamp="2010-11-16T00:00:00.0Z" accessToken ="" commandType="KeyDistributeRes"/>
        <body>
            <KeyDistributeRes errorCode=”0” errorDescription=””/>
<KeyDistributeRes errorCode=”0” errorDescription=””/>
        </body>
</message>
*/

func MakeKeyResponseBody(keyres KeyDistributeRes) ([]byte, error) {
	return nil, nil
}

func KeyDistributeHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if req.Method == "POST" {
		fmt.Println("Post req")
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Println("read error")
		}
		fmt.Println(string(data))
	}
	w.Write([]byte("{\"status\": 0, \"msg\": \"Init  SQL Faild\"}"))
}

func PostWebService(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if req.Method != "GET" {
		return
	}
	str := `"<?xml version="1.0" encoding="UTF-8"?>
<message>
    <header transactionID="100000001" timeStamp="2010-11-16T00:00:00.0Z" accessToken ="" commandType="KeyDistributeReq"/>
        <body>
           < KeyDistributeReq KeyValue="xxx" Version="xxxxxx" EffectiveDate="xxxxxxx"/>
           < KeyDistributeReq KeyValue="xxx" Version="xxxxxx" EffectiveDate="xxxxxxx"/>
        </body>
</message>"`
	fmt.Println(req.Method)
	KeyPostWebService("http://127.0.0.1:8001/key", "POST", str)
	w.Write([]byte("OK"))
}

//POST到webService
func KeyPostWebService(url string, method string, value string) string {

	res, err := http.Post(url, "text/xml; charset=utf-8", bytes.NewBuffer([]byte(value)))
	//这里随便传递了点东西
	if err != nil {
		fmt.Println("post error", err)
	}
	data, err := ioutil.ReadAll(res.Body)
	//取出主体的内容
	if err != nil {
		fmt.Println("read error", err)
	}
	res.Body.Close()
	fmt.Printf("result----%s", data)
	return string(data)
}

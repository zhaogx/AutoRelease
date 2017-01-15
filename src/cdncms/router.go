package main

import (
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"
)

func Router() {
	http.HandleFunc("/", mainhandler)
	http.HandleFunc("/notify", notifyhandler)
	http.HandleFunc("/commit", commitHandler)
	http.HandleFunc("/time", timeHandler)
	http.HandleFunc("/push", pushHandler)
	http.HandleFunc("/file", fileHandler)
	http.HandleFunc("/upload", xmlpushHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/scaninfo", ScanMsgHandler)
	http.HandleFunc("/key", KeyDistributeHandler)
	http.HandleFunc("/keytest", PostWebService)
	http.HandleFunc("/seek", seekhandler)
}

func commitHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("\"OK\""))
}

func mainhandler(w http.ResponseWriter, req *http.Request) {
	MachineInfo := runtime.GOOS + " " + runtime.GOARCH
	if runtime.GOOS == "windows" {

	}
	if runtime.GOOS == "linux" {

	}
	MachineDate := time.Now().String()
	MachineInfo += MachineDate
	for i := 0; i < 1; i++ {
		time.Sleep(1000000)
		w.Write([]byte(MachineInfo))
	}

}

func timeHandler(w http.ResponseWriter, req *http.Request) {
	MachineDate := time.Now().String()
	w.Write([]byte(MachineDate))
}

func pushHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("push"))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get option method
	fmt.Println(r.Form)
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		if len(r.Form["username"][0]) == 0 {
		}
		fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])
	}
}

type ContentReq struct {
	XMLName xml.Name `xml:"message"`
	//Version     string    `xml:"version,attr"`
	ReqHeader []header `xml:"header"`
	ReqBody   []body   `xml:"body"`
	//Description string    `xml:",innerxml"`
}

type header struct {
	XMLName       xml.Name `xml:"header"`
	transactionID string   `xml:"transactionID"`
	timeStamp     string   `xml:"timeStamp"`
	accessToken   string   `xml:"accessToken"`
	commandType   string   `xml:"commandType"`
}

type body struct {
	XMLName     xml.Name `xml:"body"`
	Reqs        []Req    `xml:"ContentDistributeReq"`
	Description string   `xml:",innerxml"`
}

type Req struct {
	XMLName      xml.Name `xml:"ContentDistributeReq"`
	StreamingNo  string   `xml:"StreamingNo"`
	CMSID        string   `xml:"CMSID"`
	ContentID    string   `xml:"ContentID"`
	ContentName  string   `xml:"ContentName"`
	ContentUrl   string   `xml:"ContentUrl"`
	UrlType      string   `xml:"UrlType"`
	ResponseType string   `xml:"ResponseType"`
	SystemId     string   `xml:"SystemId"`
	ContentType  string   `xml:"ContentType"`
	TSsupport    string   `xml:"TSsupport"`
}

func Xml2Json(xmlString string, value interface{}) (string, error) {
	if err := xml.Unmarshal([]byte(xmlString), value); err != nil {
		return "", err
	}
	js, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(js), nil
}

func xmlpushHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Printf("Read error: %v", err)
			return
		}

		defer f.Close()
		io.Copy(f, file)

		//fmt.Println(string(data))
		v := ContentReq{}
		fmt.Println(v)
		err = xml.Unmarshal(data, &v)
		if err != nil {
			fmt.Printf("xml error: %v", err)
			return
		}
		fmt.Println(v)
		fmt.Println(v.ReqBody)
		fmt.Println("\nbody-------------------------------")
		fmt.Println(v.ReqHeader)
		fmt.Println("\nhead-------------------------------")
		fmt.Println(v.XMLName)
		//fmt.Println("\ndata-------------------------------")
		//w.Write(data)

		//v1 := ContentReq{}
		//js, err := Xml2Json(string(data), &v1)
		////fmt.Println(js)
		//fmt.Fprintf(w, "%s", js)
		//fmt.Fprintf(w, "%v", v)
		//fmt.Fprint(w, "\nname-------------------------------")
		//fmt.Println(v.XMLName)
		////fmt.Println("\nmessage-------------------------------")
		////fmt.Println(v.message)
		//fmt.Println("\nheader-------------------------------")
		//fmt.Println(v.header)
		//fmt.Println("\nbody-------------------------------")
		//fmt.Println((v.body))

		//for i, val := range v.body.Reqs {
		//	fmt.Fprintln(w, v.body.Reqs[i].XMLName)
		//	fmt.Fprintln(w, v.body.Reqs[i].StreamingNo)
		//	fmt.Fprintln(w, v.body.Reqs[i].CMSID)
		//	fmt.Fprintln(w, v.body.Reqs[i].ContentID)
		//	fmt.Fprintln(w, v.body.Reqs[i].ContentName)
		//	fmt.Fprintln(w, v.body.Reqs[i].ContentUrl)
		//	fmt.Fprintln(w, v.body.Reqs[i].UrlType)
		//	fmt.Fprintln(w, v.body.Reqs[i].ResponseType)
		//	fmt.Fprintln(w, v.body.Reqs[i].SystemId)
		//	fmt.Fprintln(w, v.body.Reqs[i].ContentType)
		//	fmt.Fprintln(w, v.body.Reqs[i].TSsupport)
		//	fmt.Println(val)
		//}

		//w.Write([]byte(js))
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile("./test/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		io.Copy(f, file)
	}
}

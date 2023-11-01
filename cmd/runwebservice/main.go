package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/wuguojun0316/GoWebPratise/internal"
	_ "github.com/wuguojun0316/GoWebPratise/providers"
	"github.com/wuguojun0316/GoWebPratise/sessions"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var globalSessionManager *sessions.SessionManager

func init() {
	fmt.Println("init")
	var err error
	globalSessionManager, err = sessions.NewManager("memory", "userinfo", 0)
	if err != nil {
		fmt.Println(err.Error())
	}
	//go globalSessionManager.GC()
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(64)
	fmt.Println(r.MultipartForm)
	//fmt.Println("path", r.URL.Path)
	//fmt.Println("scheme", r.URL.Scheme)
	//fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("value:", strings.Join(v, ""))
	}
	len := r.ContentLength
	body := make([]byte, len)
	r.Body.Read(body)
	fmt.Println(string(body))
	fmt.Fprintln(w, string(body))

}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		t, _ := template.ParseFiles("../../web/login.html")
		var curtime int64 = time.Now().UnixNano()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(curtime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		w.Header().Set("Content-Type", "text/html")
		cookie := http.Cookie{Name: "token", Value: token, HttpOnly: true}
		w.Header().Set("Set-Cookie", cookie.String())
		log.Println(t.Execute(w, token))
	} else {
		err := r.ParseForm()
		if err != nil {
			log.Fatal("ParseForm: ", err)
		}
		cookie := r.Header["Cookie"]
		fmt.Println("Cookie:", cookie)
		fmt.Println("Form", r.Form)
		fmt.Println("Form", r.PostForm)
		//expiration := time.Now()
		//expiration = expiration.AddDate(0, 0, 1)
		//cookie := http.Cookie{Name: "userinfo", Value: "zmz", Expires: expiration}
		//http.SetCookie(w, &cookie)
		t, _ := template.ParseFiles("../../web/index.html")
		session := globalSessionManager.SessionStart(w, r)
		session.Set("username", r.Form["username"][0])
		log.Println(t.Execute(w, r.Form["username"][0]))
		fmt.Println("username: ", r.Form["username"])
		fmt.Println("password: ", r.Form["password"])
		fmt.Println("gender: ", r.Form["gender"])
		fmt.Println("language: ", r.Form["language"])
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("../../web/uploadfile.html")
		curtime := time.Now().UnixNano()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(curtime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		w.Header().Set("Content-Type", "text/html")
		t.Execute(w, token)
	} else if r.Method == "POST" {
		err := r.ParseMultipartForm(1024)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(r.MultipartForm.Value)
		fmt.Println("File:", len(r.MultipartForm.File))
		fileHeader := r.MultipartForm.File["filepath"][0]
		file, err := fileHeader.Open()
		if err == nil {
			data, err := ioutil.ReadAll(file)
			if err == nil {
				fmt.Fprintf(w, string(data))
			}
		}
	} else {

	}
}

func userinfo(w http.ResponseWriter, r *http.Request) {
	session, err := globalSessionManager.SessionRead(w, r)
	if err != nil {
		fmt.Println(err.Error())
		// 去登录
		http.Redirect(w, r, "login.html", 200)
	} else {
		t, _ := template.ParseFiles("../../web/userinfo.html")
		t.Execute(w, session.Get("username"))
	}
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	user := &internal.UserInfo{Response: internal.Response{Code: 100, Msg: "获取数据成功"}, Name: "zmz", Age: 18}
	data, _ := json.Marshal(user)
	w.Header().Set("Content-Type", "text/json")
	fmt.Fprintf(w, string(data))
}

func main() {
	fmt.Println("Go Web")
	http.HandleFunc("/", sayHello)
	http.HandleFunc("/login", login)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/userinfo.html", userinfo)
	http.HandleFunc("/userinfo", getUserInfo)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatalf("ListenAndServe:", err)
	}
}

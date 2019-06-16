package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dchest/captcha"
	"gopkg.in/mgo.v2"
)

// handleUser 함수는 유저리스트 페이지이다.
func handleUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "진행예정2순위 : '17.5.15~'17.6.30")
}

// handleSignup 함수는 회원가입 페이지이다.
func handleSignup(w http.ResponseWriter, r *http.Request) {
	t, err := LoadTemplates()
	if err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	captcha := struct{ CaptchaID string }{captcha.New()}
	err = t.ExecuteTemplate(w, "signup", captcha)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleSignin 함수는 로그인 페이지이다.
func handleSignin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "로그인")
}

// handleUsersInfo 함수는 유저 자료구조 페이지이다.
func handleUserinfo(w http.ResponseWriter, r *http.Request) {
	t, err := LoadTemplates()
	if err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()
	type recipe struct {
		Users []User
	}
	rcp := recipe{}
	rcp.Users, err = getUsers(session)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, "userinfo", rcp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

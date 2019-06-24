package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/digital-idea/ditime"
	"gopkg.in/mgo.v2"
)

// 이 파일은 restAPI가 설정되는 페이지이다.

// handleAPIAddproject 함수는 프로젝트를 추가한다.
func handleAPIAddproject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}

	id := r.FormValue("id")
	p := *NewProject(id)
	err = addProject(session, p)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	project, err := getProject(session, id)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	err = json.NewEncoder(w).Encode(project)
	if err != nil {
		log.Println(err)
	}
}

// handleAPIProject 함수는 프로젝트 정보를 불러온다.
func handleAPIProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	id := q.Get("id")
	project, err := getProject(session, id)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	err = json.NewEncoder(w).Encode(project)
	if err != nil {
		log.Println(err)
	}
}

// handleAPIProjectTags 함수는 프로젝트에 사용되는 태그리스트를 불러온다.
func handleAPIProjectTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data  []string `json:"data"`
		Error string   `json:"error"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	project := q.Get("project")
	_, err = getProject(session, project)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	rcp.Data, err = Distinct(session, project, "tag")
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
	}
	json.NewEncoder(w).Encode(rcp)
}

// handleAPIProjects 함수는 프로젝트 리스트를 반환한다.
func handleAPIProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data  []string `json:"data"`
		Error string   `json:"error"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		rcp.Error = "DB에 접속할 수 없습니다."
		json.NewEncoder(w).Encode(rcp)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	qStatus := q.Get("status")
	projectList, err := getProjects(session)
	if err != nil {
		rcp.Error = err.Error()
		json.NewEncoder(w).Encode(rcp)
		return
	}
	for _, p := range projectList {
		switch qStatus {
		case "unknown":
			if p.Status == UnknownProjectStatus {
				rcp.Data = append(rcp.Data, p.ID)
			}
		case "pre":
			if p.Status == PreProjectStatus {
				rcp.Data = append(rcp.Data, p.ID)
			}
		case "post":
			if p.Status == PostProjectStatus {
				rcp.Data = append(rcp.Data, p.ID)
			}
		case "layover":
			if p.Status == LayoverProjectStatus {
				rcp.Data = append(rcp.Data, p.ID)
			}
		case "backup":
			if p.Status == BackupProjectStatus {
				rcp.Data = append(rcp.Data, p.ID)
			}
		case "archive":
			if p.Status == ArchiveProjectStatus {
				rcp.Data = append(rcp.Data, p.ID)
			}
		case "lawsuit":
			if p.Status == LawsuitProjectStatus {
				rcp.Data = append(rcp.Data, p.ID)
			}
		default:
			// qStatus값이 빈 문자열이면 작업중인 프로젝트를 results 리스트에 추가한다.
			// 작업중인 상태는 pre(프리프로덕션), post(포스트프로덕션), backup(백업중)인 상태를 뜻한다.
			if p.Status == PreProjectStatus || p.Status == PostProjectStatus || p.Status == BackupProjectStatus {
				rcp.Data = append(rcp.Data, p.ID)
			}
		}
	}
	json.NewEncoder(w).Encode(rcp)
}

// handleAPIItem 함수는 아이템 자료구조를 불러온다.
func handleAPIItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	project := q.Get("project")
	slug := q.Get("slug")
	item, err := getItem(session, project, slug)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Println(err)
	}
}

// handleAPISearchname 함수는 입력 문자열을 포함하는 샷,에셋 정보를 검색한다.
func handleAPISearchname(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data []Item `json:"data"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	project := q.Get("project")
	name := q.Get("name")
	items, err := SearchName(session, project, name)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	rcp.Data = items
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		log.Println(err)
	}
}

// 검색 RestAPI
//
// Deprecated: handleAPIItems는 더 이상 사용하지 않는다. handleApi2Items를 사용할 것.
func handleAPIItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	op := SearchOption{
		Project:    q.Get("project"),
		Searchword: q.Get("searchword"),
		Sortkey:    q.Get("sortkey"),
		Assign:     str2bool(q.Get("assign")),
		Ready:      str2bool(q.Get("ready")),
		Wip:        str2bool(q.Get("wip")),
		Confirm:    str2bool(q.Get("confirm")),
		Done:       str2bool(q.Get("done")),
		Omit:       str2bool(q.Get("omit")),
		Hold:       str2bool(q.Get("hold")),
		Out:        str2bool(q.Get("out")),
		None:       str2bool(q.Get("none")),
		Shot:       str2bool(q.Get("shot")),
		Assets:     str2bool(q.Get("shot")),
		Type3d:     str2bool(q.Get("type3d")),
		Type2d:     str2bool(q.Get("type2d")),
	}
	result, err := Search(session, op)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	if len(result) == 0 {
		fmt.Fprintln(w, "{\"error\":\"검색결과 0건\"}")
		return
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Println(err)
	}
}

// handleAPI2Items 함수는 아이템을 검색한다.
func handleAPI2Items(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data []Item `json:"data"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}

	defer session.Close()
	q, err := URLUnescape(r.URL)
	if err != nil {
		log.Println(err)
		return
	}
	op := SearchOption{
		Project:    q.Get("project"),
		Searchword: q.Get("searchword"),
		Sortkey:    q.Get("sortkey"),
		Assign:     str2bool(q.Get("assign")),
		Ready:      str2bool(q.Get("ready")),
		Wip:        str2bool(q.Get("wip")),
		Confirm:    str2bool(q.Get("confirm")),
		Done:       str2bool(q.Get("done")),
		Omit:       str2bool(q.Get("omit")),
		Hold:       str2bool(q.Get("hold")),
		Out:        str2bool(q.Get("out")),
		None:       str2bool(q.Get("none")),
		Shot:       str2bool(q.Get("shot")),
		Assets:     str2bool(q.Get("shot")),
		Type3d:     str2bool(q.Get("type3d")),
		Type2d:     str2bool(q.Get("type2d")),
	}
	result, err := Search(session, op)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	rcp.Data = result
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		log.Println(err)
	}
}

// handleAPIShot 함수는 project, name을 받아서 shot을 반환한다.
func handleAPIShot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data Item `json:"data"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	project := q.Get("project")
	name := q.Get("name")
	result, err := Shot(session, project, name)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	rcp.Data = result
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		log.Println(err)
	}
}

// handleAPIUser 함수는 사용자의 id를 받아서 사용자 정보를 반환한다.
func handleAPIUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	token, err := GetToken(r)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	if *flagDebug {
		log.Println("Token:", token)
	}
	log.Println("Token:", token)
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	id := q.Get("id")
	user, err := getUser(session, id)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	type recipe struct {
		Data User `json:"data"`
	}
	rcp := recipe{}
	// 불필요한 정보는 초기화 시킨다.
	user.Password = ""
	user.Token = ""
	rcp.Data = user
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		log.Println(err)
	}
}

// handleAPISearchUser 함수는 단어를 받아서 조건에 맞는 사용자 정보를 반환한다.
func handleAPISearchUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	searchword := q.Get("searchword")
	users, err := searchUsers(session, strings.Split(searchword, ","))
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	type recipe struct {
		Data []User `json:"data"`
	}
	rcp := recipe{}
	// 불필요한 정보는 초기화 시킨다.
	for _, user := range users {
		user.Password = ""
		user.Token = ""
		rcp.Data = append(rcp.Data, user)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		log.Println(err)
	}
}

// handleAPISeqs 함수는 프로젝트의 시퀀스를 가져온다.
func handleAPISeqs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data []string `json:"data"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	project := q.Get("project")
	if project == "" {
		fmt.Fprintln(w, "{\"error\":\"project 정보가 없습니다.\"}")
		return
	}
	seqs, err := Seqs(session, project)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	rcp.Data = seqs
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		log.Println(err)
	}
}

// handleAPIShots 함수는 project, seq를 입력받아서 샷정보를 출력한다.
func handleAPIShots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data []string `json:"data"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	project := q.Get("project")
	if project == "" {
		fmt.Fprintln(w, "{\"error\":\"project 정보가 없습니다.\"}")
		return
	}
	seq := q.Get("seq")
	if seq == "" {
		fmt.Fprintln(w, "{\"error\":\"seq 정보가 없습니다.\"}")
		return
	}
	shots, err := Shots(session, project, seq)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	rcp.Data = shots
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		log.Println(err)
	}
}

// handleAPISetmov 함수는 Task에 mov를 설정한다.
func handleAPISetmov(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var task string
	var mov string
	info := r.PostForm
	for key, value := range info {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
			}
			project = v
		case "name", "shot", "asset":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
			}
			name = v
		case "task":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
			}
			// fursim은 회사에서 사용하고 있는 특수한 Task이다.
			// 샷 작업은 fursim이고 에셋작업은 fur로 불린다.(작업회의중)
			// 하지만, CSI 에서는 fur로 통일한다.
			if strings.ToLower(v) == "fursim" {
				v = "fur"
			}
			if strings.ToLower(v) == "lookdev" {
				v = "light"
			}
			if strings.ToLower(v) == "look" {
				v = "light"
			}
			if strings.ToLower(v) == "rig" {
				v = "sim"
			}
			chkTask := false
			for _, t := range TASKS {
				if strings.ToLower(t) == strings.ToLower(v) {
					chkTask = true
					break
				}
			}
			if !chkTask {
				fmt.Fprintln(w, v+"값을 Task로 사용할 수 없습니다.")
				return
			}
			task = v
		case "mov": // 앞뒤샷 포함 여러개의 mov를 등록할 수 있다.
			mov = strings.Join(value, ";")
		default:
			fmt.Fprintln(w, key+"키는 사용할 수 없습니다.(project, shot, asset, task, mov 키값만 사용가능합니다.)")
			return
		}

	}
	if mov == "" {
		fmt.Fprintln(w, "mov 키값을 설정해주세요.")
		return
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	if *flagDebug {
		fmt.Println(project, name, task, mov)
	}
	err = setMov(session, project, name, task, mov)
	if err != nil {
		fmt.Fprintln(w, err)
	}
}

// handleAPIRenderSize 함수는 아이템에 RenderSize를 설정한다.
func handleAPISetRenderSize(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var size string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "size":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			if !regexpImageSize.MatchString(v) {
				fmt.Fprintln(w, "size 는 2048x1152 형태여야 합니다.")
				return
			}
			size = v
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetImageSize(session, project, name, "rendersize", size)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPIDistortionSize 함수는 아이템의 DistortionSize를 설정한다.
func handleAPISetDistortionSize(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var size string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "size":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			if !regexpImageSize.MatchString(v) {
				fmt.Fprintln(w, "size 는 2048x1152 형태여야 합니다.")
				return
			}
			size = v
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetImageSize(session, project, name, "dsize", size)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetJustIn 함수는 아이템에 JustIn 값을 설정한다.
func handleAPISetJustIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var frame int
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "frame":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			frame = n
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetFrame(session, project, name, "justin", frame)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetJustOut 함수는 아이템에 JustOut 값을 설정한다.
func handleAPISetJustOut(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var frame int
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "frame":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			frame = n
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetFrame(session, project, name, "justout", frame)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPIPlateSize 함수는 아이템의 PlateSize를 설정한다.
func handleAPISetPlateSize(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var size string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "size":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			if !regexpImageSize.MatchString(v) {
				fmt.Fprintln(w, "size 는 2048x1152 형태여야 합니다.")
				return
			}
			size = v
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetImageSize(session, project, name, "platesize", size)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// PostFormValueInList 는 PostForm 쿼리시 Value값이 1개라면 값을 리턴한다.
func PostFormValueInList(key string, values []string) (string, error) {
	if len(values) != 1 {
		return "", errors.New(key + "값이 여러개 입니다.")
	}
	if key == "startdate" && values[0] == "" { // Task 시작일은 빈 문자를 허용한다.
		return "", nil
	}
	if key == "predate" && values[0] == "" { // 1차마감일은 빈 문자를 허용한다.
		return "", nil
	}
	if key == "date" && values[0] == "" { // 2차마감일은 빈 문자를 허용한다.
		return "", nil
	}
	if values[0] == "" {
		return "", errors.New(key + "값이 빈 문자입니다.")
	}
	return values[0], nil
}

// handleAPISetCameraPubPath 함수는 아이템의 Camera PubPath를 설정한다.
func handleAPISetCameraPubPath(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var path string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "path":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			path = v
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetCameraPubPath(session, project, name, path)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetCameraPubTask 함수는 아이템의 Camera PubTask를 설정한다.
func handleAPISetCameraPubTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var task string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "task":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			task = v
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetCameraPubTask(session, project, name, task)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetCameraProjection 함수는 아이템의 Camera Projection 여부를 설정한다.
func handleAPISetCameraProjection(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var projection bool
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "projection":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			if strings.ToLower(v) == "true" {
				projection = true
			}
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetCameraProjection(session, project, name, projection)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetThummov 함수는 아이템의 Thummov 값을 설정한다.
func handleAPISetThummov(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var path string
	var typ string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "path":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			path = v
		case "type":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			typ = v
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetThummov(session, project, name, typ, path)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetStatus 함수는 아이템의 task에 대한 상태를 설정한다.
func handleAPISetStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var task string
	var status string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "task":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			task = v
		case "status":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			status = v
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetStatus(session, project, name, task, status)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetStartdate 함수는 아이템의 task에 대한 시작일을 설정한다.
func handleAPISetStartdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var task string
	var startdate string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "task":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			task = v
		case "startdate":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			startdate, err = ditime.ToFullTime("current", v) // 작업시작시간은 현재시간으로 등록한다.
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetStartdate(session, project, name, task, startdate)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetPredate 함수는 아이템의 task에 대한 1차마감일을 설정한다.
func handleAPISetPredate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm() // 받은 문자를 파싱합니다. 파싱되면 map이 됩니다.
	var project string
	var name string
	var task string
	var predate string
	args := r.PostForm
	for key, value := range args {
		switch key {
		case "project":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			project = v
		case "name":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			name = v
		case "task":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			task = v
		case "predate":
			v, err := PostFormValueInList(key, value)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			predate, err = ditime.ToFullTime("end", v) // 1차 마감일은 퇴근시간으로 등록한다.
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
		}
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintln(w, *flagDBIP+" DB에 접속할 수 없습니다.")
		return
	}
	defer session.Close()
	err = SetPredate(session, project, name, task, predate)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
}

// handleAPISetelliteItems 함수는 project, rollmedia을 받아서 setellite 정보를 반환한다.
func handleAPISetelliteItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data []Setellite `json:"data"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	project := q.Get("project")
	rollmedia := q.Get("rollmedia")
	rcp.Data, err = SetelliteItems(session, project, rollmedia)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		fmt.Fprintf(w, "{\"err\":\"%v\"}\n", err)
	}
}

// handleAPISetelliteSearch 함수는 project, searchword의 검색어를 통해 setellite 정보를 반환한다.
func handleAPISetelliteSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Get Only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	type recipe struct {
		Data []Setellite `json:"data"`
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	defer session.Close()
	q := r.URL.Query()
	project := q.Get("project")
	searchword := q.Get("searchword")
	rcp.Data, err = SetelliteSearch(session, project, searchword)
	if err != nil {
		fmt.Fprintf(w, "{\"error\":\"%v\"}\n", err)
		return
	}
	err = json.NewEncoder(w).Encode(rcp)
	if err != nil {
		fmt.Fprintf(w, "{\"err\":\"%v\"}\n", err)
	}
}

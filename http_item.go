package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/digital-idea/dipath"
	"github.com/disintegration/imaging"
	"gopkg.in/mgo.v2"
)

// handleSearchSubmitv2 함수는 검색창의 옵션을 파싱하고 검색 URI로 리다이렉션 한다.
func handleSearchSubmitv2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Post Only", http.StatusMethodNotAllowed)
		return
	}
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
		return
	}
	Project := r.FormValue("Project")
	Searchword := r.FormValue("Searchword")
	Sortkey := r.FormValue("Sortkey")
	Assign := str2bool(r.FormValue("Assign"))
	Ready := str2bool(r.FormValue("Ready"))
	Wip := str2bool(r.FormValue("Wip"))
	Confirm := str2bool(r.FormValue("Confirm"))
	Done := str2bool(r.FormValue("Done"))
	Omit := str2bool(r.FormValue("Omit"))
	Hold := str2bool(r.FormValue("Hold"))
	Out := str2bool(r.FormValue("Out"))
	None := str2bool(r.FormValue("None"))
	Template := r.FormValue("Template")
	Task := r.FormValue("Task")
	redirectURL := fmt.Sprintf(`/inputmode?project=%s&searchword=%s&sortkey=%s&assign=%t&ready=%t&wip=%t&confirm=%t&done=%t&omit=%t&hold=%t&out=%t&none=%t&template=%s&task=%s`,
		Project,
		Searchword,
		Sortkey,
		Assign,
		Ready,
		Wip,
		Confirm,
		Done,
		Omit,
		Hold,
		Out,
		None,
		Template,
		Task,
	)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// handleEditItem 함수는 Item 편집페이지이다.
func handleEditItem(w http.ResponseWriter, r *http.Request) {
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	q := r.URL.Query()
	project := q.Get("project")
	slug := q.Get("slug")
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()
	type recipe struct {
		ID      string
		Project Project
		Item    Item
		Devmode bool
		User
		SearchOption
	}
	rcp := recipe{}
	rcp.Devmode = *flagDevmode
	rcp.User, err = getUser(session, ssid.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.SearchOption = handleRequestToSearchOption(r)
	rcp.Project, err = getProject(session, project)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.Item, err = getItem(session, project, slug)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = TEMPLATES.ExecuteTemplate(w, "edititem", rcp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleEditedItem 함수는 Item 이 수정된 이후 이동하는 페이지이다.
func handleEditedItem(w http.ResponseWriter, r *http.Request) {
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err = TEMPLATES.ExecuteTemplate(w, "editeditem", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleIndex 함수는 index 페이지이다.
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// 등록되지 않은 URL을 처리한다.
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
		return
	}
	type recipe struct {
		SearchOption
	}
	rcp := recipe{}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()
	err = rcp.SearchOption.LoadCookie(session, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 아무 상태도 선택되어있지 않다면 기본 상태설정으로 변경한다.
	if rcp.SearchOption.isStatusOff() {
		rcp.SearchOption.setStatusDefault()
	}

	url := fmt.Sprintf("/inputmode?project=%s&sortkey=%s&template=index&endpoint=searchv2&assign=%t&ready=%t&wip=%t&confirm=%t&done=%t&omit=%t&hold=%t&out=%t&none=%t&task=%s&searchword=%s",
		rcp.SearchOption.Project,
		rcp.SearchOption.Sortkey,
		rcp.SearchOption.Assign,
		rcp.SearchOption.Ready,
		rcp.SearchOption.Wip,
		rcp.SearchOption.Confirm,
		rcp.SearchOption.Done,
		rcp.SearchOption.Omit,
		rcp.SearchOption.Hold,
		rcp.SearchOption.Out,
		rcp.SearchOption.None,
		rcp.SearchOption.Task,
		rcp.SearchOption.Searchword,
	)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// handleAddShot 함수는 shot을 추가하는 페이지이다.
func handleAddShot(w http.ResponseWriter, r *http.Request) {
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
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
		User        User
		Projectlist []string
		Devmode     bool
		SearchOption
	}
	rcp := recipe{}
	err = rcp.SearchOption.LoadCookie(session, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.Devmode = *flagDevmode
	u, err := getUser(session, ssid.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.User = u
	rcp.Projectlist, err = Projectlist(session)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = TEMPLATES.ExecuteTemplate(w, "addShot", rcp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleAddShotSubmit 함수는 shot을 생성한다.
func handleAddShotSubmit(w http.ResponseWriter, r *http.Request) {
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
		return
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()
	project := r.FormValue("Project")
	name := r.FormValue("Name")
	stereo := r.FormValue("Stereo")
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_'
	}
	names := strings.FieldsFunc(name, f)
	type Shot struct {
		Project string
		Name    string
		Error   string
	}
	var success []Shot
	var fails []Shot
	for _, n := range names {
		if n == " " || n == "" { // 사용자가 실수로 여러개의 스페이스를 추가할 수 있다.
			continue
		}
		s := Shot{}
		s.Project = project
		s.Name = n
		if !regexpShotname.MatchString(n) {
			s.Error = "SS_0010 형식의 이름이 아닙니다"
			fails = append(fails, s)
			continue
		}
		i := Item{}
		i.Name = n
		i.Type = "org"
		if stereo == "true" {
			i.Type = "left"
		}
		i.Project = project
		i.ID = i.Name + "_" + i.Type
		i.Seq = strings.Split(i.Name, "_")[0]
		i.Cut = strings.Split(i.Name, "_")[1]
		i.Shottype = "2d"
		i.Thumpath = fmt.Sprintf("/%s/%s_%s.jpg", i.Project, i.Name, i.Type)
		i.Platepath = fmt.Sprintf("/show/%s/seq/%s/%s/plate/", i.Project, i.Seq, i.Name)
		i.Thummov = fmt.Sprintf("/show/%s/seq/%s/%s/plate/%s_%s.mov", i.Project, i.Seq, i.Name, i.Name, i.Type)
		i.Scantime = time.Now().Format(time.RFC3339)
		i.Updatetime = time.Now().Format(time.RFC3339)
		i.Status = ASSIGN
		// 기본적으로 comp Task를 추가한다.
		i.Tasks = make(map[string]Task)
		i.Tasks["comp"] = Task{Status: ASSIGN, Title: "comp"}
		err = addItem(session, project, i)
		if err != nil {
			s.Error = err.Error()
			fails = append(fails, s)
			continue
		}
		success = append(success, s)
		// slack log
		err = slacklog(session, project, fmt.Sprintf("Add Shot: %s\nProject: %s, Author: %s", n, project, ssid.ID))
		if err != nil {
			log.Println(err)
			continue
		}
	}

	type recipe struct {
		Success []Shot
		Fails   []Shot
		User    User
		Devmode bool
		SearchOption
	}
	rcp := recipe{}
	err = rcp.SearchOption.LoadCookie(session, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.Devmode = *flagDevmode
	rcp.Success = success
	rcp.Fails = fails
	u, err := getUser(session, ssid.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.User = u

	w.Header().Set("Content-Type", "text/html")
	err = TEMPLATES.ExecuteTemplate(w, "addShot_success", rcp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleAddAsset 함수는 asset을 추가하는 페이지이다.
func handleAddAsset(w http.ResponseWriter, r *http.Request) {
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
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
		User        User
		Projectlist []string
		Devmode     bool
		SearchOption
	}
	rcp := recipe{}
	err = rcp.SearchOption.LoadCookie(session, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.Devmode = *flagDevmode
	u, err := getUser(session, ssid.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.User = u
	rcp.Projectlist, err = Projectlist(session)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = TEMPLATES.ExecuteTemplate(w, "addAsset", rcp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleAddAssetSubmit 함수는 asset을 생성한다.
func handleAddAssetSubmit(w http.ResponseWriter, r *http.Request) {
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
		return
	}
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()
	project := r.FormValue("Project")
	name := r.FormValue("Name")
	assettype := r.FormValue("Assettype")
	construction := r.FormValue("Construction")
	crowdAsset := str2bool(r.FormValue("CrowdAsset"))
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '_'
	}
	names := strings.FieldsFunc(name, f)
	type Asset struct {
		Project string
		Name    string
		Error   string
	}
	var success []Asset
	var fails []Asset
	for _, n := range names {
		if n == " " || n == "" { // 사용자가 실수로 여러개의 스페이스를 추가할 수 있다.
			continue
		}
		a := Asset{}
		a.Project = project
		a.Name = n
		if !regexpAssetname.MatchString(n) {
			a.Error = "stone 또는 stone01 형식의 이름이 아닙니다"
			fails = append(fails, a)
			continue
		}
		i := Item{}
		i.Name = n
		i.Type = "asset"
		i.Project = project
		i.ID = i.Name + "_" + i.Type
		i.Status = ASSIGN
		i.Updatetime = time.Now().Format(time.RFC3339)
		i.Assettype = assettype
		i.Assettags = []string{assettype, construction}
		i.CrowdAsset = crowdAsset
		// 기본적인 Task를 추가한다.
		i.Tasks = make(map[string]Task)
		if assettype == "comp" {
			t := Task{
				Status: ASSIGN,
				Title:  "comp",
			}
			i.Tasks["comp"] = t
		} else if assettype == "env" {
			t := Task{
				Status: ASSIGN,
				Title:  "env",
			}
			i.Tasks["env"] = t
			t = Task{
				Status: ASSIGN,
				Title:  "matte",
			}
			i.Tasks["matte"] = t
		} else {
			t := Task{
				Status: ASSIGN,
				Title:  "model",
			}
			i.Tasks["model"] = t
		}
		err = addItem(session, project, i)
		if err != nil {
			a.Error = err.Error()
			fails = append(fails, a)
			continue
		}
		success = append(success, a)
	}

	type recipe struct {
		Success []Asset
		Fails   []Asset
		User    User
		Devmode bool
		SearchOption
	}
	rcp := recipe{}
	err = rcp.SearchOption.LoadCookie(session, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.Devmode = *flagDevmode
	rcp.Success = success
	rcp.Fails = fails
	u, err := getUser(session, ssid.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.User = u

	w.Header().Set("Content-Type", "text/html")
	err = TEMPLATES.ExecuteTemplate(w, "addAsset_success", rcp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleEditItemSubmitv2 함수는 Item의 수정사항을 처리하는 페이지이다. legacy
func handleEditItemSubmitv2(w http.ResponseWriter, r *http.Request) {
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	if ssid.AccessLevel == 0 {
		http.Redirect(w, r, "/invalidaccess", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	//var logstring string
	//기존 Item의 값을 가지고 온다.
	project := r.FormValue("project")
	id := r.FormValue("id")
	file, fileHandler, fileErr := r.FormFile("Thumbnail")
	if fileErr == nil {
		if !(fileHandler.Header.Get("Content-Type") == "image/jpeg" || fileHandler.Header.Get("Content-Type") == "image/png") {
			http.Error(w, "업로드 파일이 jpeg 또는 png 파일이 아닙니다", http.StatusInternalServerError)
			return
		}
		//파일이 없다면 fileErr 값은 "http: no such file" 값이 된다.
		// 썸네일 파일이 존재한다면 아래 프로세스를 거친다.
		mediatype, fileParams, err := mime.ParseMediaType(fileHandler.Header.Get("Content-Disposition"))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if *flagDebug {
			fmt.Println(mediatype)
			fmt.Println(fileParams)
			fmt.Println(fileHandler.Header.Get("Content-Type"))
			fmt.Println()
		}
		tempPath := os.TempDir() + fileHandler.Filename
		tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 사용자가 업로드한 파일을 tempFile에 복사한다.
		io.Copy(tempFile, io.LimitReader(file, MaxFileSize))
		tempFile.Close()
		defer os.Remove(tempPath)
		//fmt.Println(tempPath)
		thumbnailPath := fmt.Sprintf("%s/%s/%s.jpg", *flagThumbPath, project, id)
		thumbnailDir := filepath.Dir(thumbnailPath)
		// 썸네일을 생성할 경로가 존재하지 않는다면 생성한다.
		_, err = os.Stat(thumbnailDir)
		if os.IsNotExist(err) {
			err := os.MkdirAll(thumbnailDir, 0775)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// 디지털아이디어의 경우 스캔시스템에서 수동으로 이미지를 폴더에 생성하는 경우가 있다.
			if *flagCompany == "digitalidea" {
				err = dipath.Ideapath(thumbnailDir)
				if err != nil {
					log.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		// 이미지변환
		src, err := imaging.Open(tempPath)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Resize the cropped image to width = 200px preserving the aspect ratio.
		dst := imaging.Fill(src, 410, 222, imaging.Center, imaging.Lanczos)
		err = imaging.Save(dst, thumbnailPath)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 디지털아이디어의 경우 스캔시스템에서 수동으로 이미지를 수정하는 경우도 있다.
		if *flagCompany == "digitalidea" {
			err = dipath.Ideapath(thumbnailPath)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	http.Redirect(w, r, "/editeditem", http.StatusSeeOther)
}

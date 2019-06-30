package main

import (
	"log"
	"net/http"
	"strconv"

	"gopkg.in/mgo.v2"
)

// handleAddProject 함수는 프로젝트를 추가하는 페이지이다.
func handleAddProject(w http.ResponseWriter, r *http.Request) {
	_, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	t, err := LoadTemplates()
	if err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err = t.ExecuteTemplate(w, "addProject", nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleProjectinfo 함수는 프로젝트 자료구조 페이지이다.
func handleProjectinfo(w http.ResponseWriter, r *http.Request) {
	ssid, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	t, err := LoadTemplates()
	if err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	q := r.URL.Query()
	status := q.Get("status")
	w.Header().Set("Content-Type", "text/html")
	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer session.Close()
	type recipe struct {
		Projects []Project
		MailDNS  string
		User     User
	}
	rcp := recipe{}
	u, err := getUser(session, ssid.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rcp.User = u
	rcp.MailDNS = *flagMailDNS
	if status != "" {
		rcp.Projects, err = getStatusProjects(session, ToProjectStatus(status))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		rcp.Projects, err = getProjects(session)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = t.ExecuteTemplate(w, "projectinfo", rcp)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ToProjectStatus 함수는 문자를 받아서 ProjectStatus 형으로 변환합니다.
func ToProjectStatus(s string) ProjectStatus {
	switch s {
	case "pre", "ready", "준비":
		return PreProjectStatus
	case "post", "wip":
		return PostProjectStatus
	case "layover", "중단":
		return LayoverProjectStatus
	case "backup", "백업":
		return BackupProjectStatus
	case "archive", "done", "종료":
		return ArchiveProjectStatus
	case "lawsuit", "소송":
		return LawsuitProjectStatus
	default:
		return UnknownProjectStatus
	}
}

// handleEditProjectSubmit 함수는 Projectinfo의  수정정보를 처리하는 페이지이다.
func handleEditProjectSubmit(w http.ResponseWriter, r *http.Request) {
	_, err := GetSessionID(r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
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
	current, err := getProject(session, r.FormValue("Id"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renewal := current //과거 프로젝트 값을 셋팅한다.
	if current.Name != r.FormValue("Name") {
		renewal.Name = r.FormValue("Name")
	}

	if current.MailHead != r.FormValue("MailHead") {
		renewal.MailHead = r.FormValue("MailHead")
	}
	if current.Style != r.FormValue("Style") {
		renewal.Style = r.FormValue("Style")
	}
	if current.Stereo != str2bool(r.FormValue("Stereo")) {
		renewal.Stereo = str2bool(r.FormValue("Stereo"))
	}
	if current.Screenx != str2bool(r.FormValue("Screenx")) {
		renewal.Screenx = str2bool(r.FormValue("Screenx"))
	}
	if current.Netapp != str2bool(r.FormValue("Netapp")) {
		renewal.Netapp = str2bool(r.FormValue("Netapp"))
	}
	if current.Director != r.FormValue("Director") {
		renewal.Director = r.FormValue("Director")
	}
	if current.Super != r.FormValue("Super") {
		renewal.Super = r.FormValue("Super")
	}
	renewal.CgSuper = r.FormValue("CgSuper")
	renewal.Pd = r.FormValue("Pd")
	renewal.Pm = r.FormValue("Pm")
	renewal.Pa = r.FormValue("Pa")
	renewal.Message = r.FormValue("Message")
	renewal.Wiki = r.FormValue("Wiki")
	renewal.Edit = r.FormValue("Edit")
	renewal.NoteHighlight = r.FormValue("NoteHighlight")
	aspectratio, err := strconv.ParseFloat(r.FormValue("AspectRatio"), 64)
	if err == nil {
		renewal.AspectRatio = aspectratio
	}
	startframe, err := strconv.Atoi(r.FormValue("StartFrame"))
	if err == nil {
		renewal.StartFrame = startframe
	}
	versionnum, err := strconv.Atoi(r.FormValue("VersionNum"))
	if err == nil {
		renewal.VersionNum = versionnum
	}
	seqnum, err := strconv.Atoi(r.FormValue("SeqNum"))
	if err == nil {
		renewal.SeqNum = seqnum
	}
	renewal.Issue = r.FormValue("Issue")
	platewidth, err := strconv.Atoi(r.FormValue("PlateWidth"))
	if err == nil {
		renewal.PlateWidth = platewidth
	}
	plateheight, err := strconv.Atoi(r.FormValue("PlateHeight"))
	if err == nil {
		renewal.PlateHeight = plateheight
	}
	renewal.ResizeType = r.FormValue("ResizeType")
	renewal.PlateExt = r.FormValue("PlateExt")
	renewal.Camera = r.FormValue("Camera")
	renewal.PlateInColorspace = r.FormValue("PlateInColorspace")
	renewal.PlateOutColorspace = r.FormValue("PlateOutColorspace")
	renewal.ProxyOutColorspace = r.FormValue("ProxyOutColorspace")
	renewal.PostProductionProxyCodec = r.FormValue("PostProductionProxyCodec")
	outputmovWidth, err := strconv.Atoi(r.FormValue("OutputMov.Width"))
	if err == nil {
		renewal.OutputMov.Width = outputmovWidth
	}
	outputmovHeight, err := strconv.Atoi(r.FormValue("OutputMov.Height"))
	if err == nil {
		renewal.OutputMov.Height = outputmovHeight
	}
	renewal.OutputMov.Codec = r.FormValue("OutputMov.Codec")
	outputmovFps, err := strconv.ParseFloat(r.FormValue("OutputMov.Fps"), 64)
	if err == nil {
		renewal.OutputMov.Fps = outputmovFps
	}
	renewal.OutputMov.InColorspace = r.FormValue("OutputMov.InColorspace")
	renewal.OutputMov.OutColorspace = r.FormValue("OutputMov.OutColorspace")
	editmovWidth, err := strconv.Atoi(r.FormValue("EditMov.Width"))
	if err == nil {
		renewal.EditMov.Width = editmovWidth
	}
	editmovHeight, err := strconv.Atoi(r.FormValue("EditMov.Height"))
	if err == nil {
		renewal.EditMov.Height = editmovHeight
	}
	renewal.EditMov.Codec = r.FormValue("EditMov.Codec")
	editmovFps, err := strconv.ParseFloat(r.FormValue("EditMov.Fps"), 64)
	if err == nil {
		renewal.EditMov.Fps = editmovFps
	}
	renewal.EditMov.InColorspace = r.FormValue("EditMov.InColorspace")
	renewal.EditMov.OutColorspace = r.FormValue("EditMov.OutColorspace")
	// 마일스톤 추가하기.
	status, err := strconv.Atoi(r.FormValue("Status"))
	if err == nil {
		renewal.Status = ProjectStatus(status)
	}
	renewal.Lut = r.FormValue("Lut")
	renewal.LutInColorspace = r.FormValue("LutInColorspace")
	renewal.LutOutColorspace = r.FormValue("LutOutColorspace")
	renewal.Description = r.FormValue("Description")
	renewal.NukeGizmo = r.FormValue("NukeGizmo")
	renewal.FxElement = r.FormValue("FxElement")
	renewal.MayaCropMaskSize = r.FormValue("MayaCropMaskSize")
	cropaspectratio, err := strconv.ParseFloat(r.FormValue("CropAspectRatio"), 64)
	if err == nil {
		renewal.CropAspectRatio = cropaspectratio
	}
	houdiniImportScale, err := strconv.ParseFloat(r.FormValue("HoudiniImportScale"), 64)
	if err == nil {
		renewal.HoudiniImportScale = houdiniImportScale
	}
	screenxOverlay, err := strconv.ParseFloat(r.FormValue("ScreenxOverlay"), 64)
	if err == nil {
		renewal.ScreenxOverlay = screenxOverlay
	}
	// 새로 변경된 정보를 DB에 저장한다.
	err = setProject(session, renewal)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, "edited", nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

package main

import (
	"fmt"
	"log"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Searchv2 함수는 다음 검색함수이다.
// 검색알고리즘은 복잡하기 때문에 한 파일에서 다룬다.
func Searchv2(session *mgo.Session, op SearchOption) ([]Item, error) {
	results := []Item{}
	// 검색어가 없다면 바로 빈 값을 리턴한다.
	if op.Searchword == "" {
		return results, nil
	}
	// 체크박스가 아무것도 켜있지 않다면 바로 빈 값을 리턴한다.
	if !op.Assign && !op.Ready && !op.Wip && !op.Confirm && !op.Done && !op.Omit && !op.Hold && !op.Out && !op.None {
		return results, nil
	}
	// 검색어중 연산에 필요한 검색어는 제거한다.
	var words []string
	var selectTasks []string
	for _, word := range strings.Split(op.Searchword, " ") {
		// task를 searchbox UX가 아닌 타이핑으로도 선언할 수 있어야 한다.
		if strings.HasPrefix(word, "task:") {
			selectTasks = append(selectTasks, strings.TrimPrefix(word, "task:"))
			continue
		}
		switch word {
		case "":
		case "or", "||":
		case "and", "&&":
		default:
			words = append(words, word)
		}
	}

	// 검색어가 존재하지 않으면 빈 결과를 반환한다.
	if len(words) == 0 {
		return results, nil
	}
	// 프로젝트 문자열이 빈 값이라면 전체 리스트중에서 첫번째 프로젝트를 선언한다.
	if op.Project == "" {
		plist, err := Projectlist(session)
		if err != nil {
			return results, err
		}
		op.Project = plist[0]
	}

	session.SetMode(mgo.Monotonic, true)
	c := session.DB("project").C(op.Project)

	// Task 처리
	allTasks, err := TasksettingNames(session)
	if err != nil {
		log.Println(err)
	}
	if op.Task != "" {
		selectTasks = append(selectTasks, op.Task)
	}

	wordQueries := []bson.M{}

	for _, word := range words {
		query := []bson.M{}
		if MatchShortTime.MatchString(word) { // 1121 형식의 날짜
			regFullTime := fmt.Sprintf(`^\d{4}-%s-%sT\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}$`, word[0:2], word[2:4])
			if len(selectTasks) == 0 {
				for _, task := range allTasks {
					query = append(query, bson.M{"tasks." + strings.ToLower(task) + ".date": &bson.RegEx{Pattern: regFullTime}})
					query = append(query, bson.M{"tasks." + strings.ToLower(task) + ".predate": &bson.RegEx{Pattern: regFullTime}})
				}
				query = append(query, bson.M{"ddline2d": &bson.RegEx{Pattern: regFullTime}})
				query = append(query, bson.M{"ddline3d": &bson.RegEx{Pattern: regFullTime}})
			} else {
				for _, task := range selectTasks {
					query = append(query, bson.M{"tasks." + task + ".date": &bson.RegEx{Pattern: regFullTime}})
					query = append(query, bson.M{"tasks." + task + ".predate": &bson.RegEx{Pattern: regFullTime}})
				}
			}
			query = append(query, bson.M{"name": &bson.RegEx{Pattern: word}}) // 샷 이름에 숫자가 포함되는 경우도 검색한다.
		} else if MatchNormalTime.MatchString(word) {
			// 데일리 날짜를 검색한다.
			// 2016-11-21 형태는 데일리로 간주합니다.
			// jquery 달력의 기본형식이기도 합니다.
			regFullTime := fmt.Sprintf(`^%sT\d{2}:\d{2}:\d{2}[-+]\d{2}:\d{2}$`, word)
			if len(selectTasks) == 0 {
				for _, task := range allTasks {
					query = append(query, bson.M{"tasks." + strings.ToLower(task) + ".mdate": &bson.RegEx{Pattern: regFullTime}})
				}
			} else {
				for _, task := range selectTasks {
					query = append(query, bson.M{"tasks." + strings.ToLower(task) + ".mdate": &bson.RegEx{Pattern: regFullTime}})
				}
			}
		} else if regexpTimecode.MatchString(word) {
			query = append(query, bson.M{"justtimecodein": word})
			query = append(query, bson.M{"justtimecodeout": word})
			query = append(query, bson.M{"scantimecodein": word})
			query = append(query, bson.M{"scantimecodeout": word})
		} else if strings.HasPrefix(word, "tag:") {
			query = append(query, bson.M{"tag": strings.TrimPrefix(word, "tag:")})
		} else if strings.HasPrefix(word, "assettags:") {
			query = append(query, bson.M{"assettags": strings.TrimPrefix(word, "assettags:")})
		} else if strings.HasPrefix(word, "deadline2d:") {
			query = append(query, bson.M{"ddline2d": &bson.RegEx{Pattern: strings.TrimPrefix(word, "deadline2d:"), Options: "i"}})
		} else if strings.HasPrefix(word, "deadline3d:") {
			query = append(query, bson.M{"ddline3d": &bson.RegEx{Pattern: strings.TrimPrefix(word, "deadline3d:"), Options: "i"}})
		} else if strings.HasPrefix(word, "shottype:") {
			query = append(query, bson.M{"shottype": &bson.RegEx{Pattern: strings.TrimPrefix(word, "shottype:"), Options: "i"}})
		} else if strings.HasPrefix(word, "type:shot") {
			query = append(query, bson.M{"$or": []bson.M{bson.M{"type": "org"}, bson.M{"type": "left"}}})
		} else if strings.HasPrefix(word, "type:asset") {
			query = append(query, bson.M{"type": "asset"})
		} else if strings.HasPrefix(word, "status:") {
			status := strings.ToLower(strings.TrimPrefix(word, "status:"))
			if len(selectTasks) != 0 {
				for _, task := range selectTasks {
					switch status {
					case "assign":
						query = append(query, bson.M{"tasks." + task + ".status": ASSIGN})
					case "ready":
						query = append(query, bson.M{"tasks." + task + ".status": READY})
					case "wip":
						query = append(query, bson.M{"tasks." + task + ".status": WIP})
					case "confirm":
						query = append(query, bson.M{"tasks." + task + ".status": CONFIRM})
					case "done":
						query = append(query, bson.M{"tasks." + task + ".status": DONE})
					case "omit":
						query = append(query, bson.M{"tasks." + task + ".status": OMIT})
					case "hold":
						query = append(query, bson.M{"tasks." + task + ".status": HOLD})
					case "out":
						query = append(query, bson.M{"tasks." + task + ".status": OUT})
					case "none":
						query = append(query, bson.M{"tasks." + task + ".status": NONE})
					}
				}
			} else {
				switch status {
				case "assign":
					query = append(query, bson.M{"status": ASSIGN})
				case "ready":
					query = append(query, bson.M{"status": READY})
				case "wip":
					query = append(query, bson.M{"status": WIP})
				case "confirm":
					query = append(query, bson.M{"status": CONFIRM})
				case "done":
					query = append(query, bson.M{"status": DONE})
				case "omit":
					query = append(query, bson.M{"status": OMIT})
				case "hold":
					query = append(query, bson.M{"status": HOLD})
				case "out":
					query = append(query, bson.M{"status": OUT})
				case "none":
					query = append(query, bson.M{"status": NONE})
				}
			}
		} else if strings.HasPrefix(word, "user:") {
			if len(selectTasks) == 0 {
				if strings.TrimPrefix(word, "user:") == "notassign" {
					for _, task := range allTasks {
						query = append(query, bson.M{"tasks." + strings.ToLower(task) + ".user": ""})
					}
				} else {
					for _, task := range allTasks {
						query = append(query, bson.M{"tasks." + strings.ToLower(task) + ".user": &bson.RegEx{Pattern: strings.TrimPrefix(word, "user:")}})
					}
				}
			} else {
				for _, task := range selectTasks {
					if strings.TrimPrefix(word, "user:") == "notassign" {
						query = append(query, bson.M{"tasks." + task + ".user": ""})
					} else {
						query = append(query, bson.M{"tasks." + task + ".user": &bson.RegEx{Pattern: strings.TrimPrefix(word, "user:")}})
					}
				}
			}
		} else if strings.HasPrefix(word, "rnum:") { // 롤넘버 형태일 때
			query = append(query, bson.M{"rnum": &bson.RegEx{Pattern: strings.TrimPrefix(word, "rnum:"), Options: "i"}})
		} else if regexTaskStatusQuery.MatchString(word) {
			// 위 패턴이면 : 문자로 스플릿하고 상태를 숫자로 바꾼다.
			queryString := strings.Split(word, ":")[0]
			status := StatusString2string(strings.Split(word, ":")[1])
			query = append(query, bson.M{queryString: status})
		} else {
			switch word {
			case "all", "All", "ALL", "올", "미ㅣ", "dhf", "전체":
				query = append(query, bson.M{})
			case "shot", "샷", "전샷", "전체샷":
				query = append(query, bson.M{"type": "org"})
				query = append(query, bson.M{"type": "left"})
			case "asset", "assets", "에셋":
				query = append(query, bson.M{"type": "asset"})
			case "전권":
				query = append(query, bson.M{"tag": "1권"})
				query = append(query, bson.M{"tag": "2권"})
				query = append(query, bson.M{"tag": "3권"})
				query = append(query, bson.M{"tag": "4권"})
				query = append(query, bson.M{"tag": "5권"})
				query = append(query, bson.M{"tag": "6권"})
				query = append(query, bson.M{"tag": "7권"})
				query = append(query, bson.M{"tag": "8권"})
			default:
				query = append(query, bson.M{"id": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"comments.text": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"sources.title": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"sources.path": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"references.title": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"references.path": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"note.text": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"tag": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"assettags": &bson.RegEx{Pattern: word, Options: "i"}})
				query = append(query, bson.M{"scanname": &bson.RegEx{Pattern: word, Options: ""}})
				query = append(query, bson.M{"rnum": &bson.RegEx{Pattern: word, Options: ""}})
				// Task가 선언 되어있을 때
				if len(selectTasks) == 0 {
					for _, task := range allTasks {
						query = append(query, bson.M{"tasks." + strings.ToLower(task) + ".user": &bson.RegEx{Pattern: word}})
					}
				} else {
					for _, task := range selectTasks {
						query = append(query, bson.M{"tasks." + strings.ToLower(task) + ".user": &bson.RegEx{Pattern: word}})
					}
				}
			}
		}
		wordQueries = append(wordQueries, bson.M{"$or": query})
	}

	statusQueries := []bson.M{}
	if len(selectTasks) == 0 {
		if op.Assign {
			statusQueries = append(statusQueries, bson.M{"status": ASSIGN})
		}
		if op.Ready {
			statusQueries = append(statusQueries, bson.M{"status": READY})
		}
		if op.Wip {
			statusQueries = append(statusQueries, bson.M{"status": WIP})
		}
		if op.Confirm {
			statusQueries = append(statusQueries, bson.M{"status": CONFIRM})
		}
		if op.Done {
			statusQueries = append(statusQueries, bson.M{"status": DONE})
		}
		if op.Omit {
			statusQueries = append(statusQueries, bson.M{"status": OMIT})
		}
		if op.Hold {
			statusQueries = append(statusQueries, bson.M{"status": HOLD})
		}
		if op.Out {
			statusQueries = append(statusQueries, bson.M{"status": OUT})
		}
		if op.None {
			statusQueries = append(statusQueries, bson.M{"status": NONE})
		}
	} else {
		for _, task := range selectTasks {
			if op.Assign {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": ASSIGN})
			}
			if op.Ready {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": READY})
			}
			if op.Wip {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": WIP})
			}
			if op.Confirm {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": CONFIRM})
			}
			if op.Done {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": DONE})
			}
			if op.Omit {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": OMIT})
			}
			if op.Hold {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": HOLD})
			}
			if op.Out {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": OUT})
			}
			if op.None {
				statusQueries = append(statusQueries, bson.M{"tasks." + task + ".status": NONE})
			}
		}
	}
	// 각 단어에 대한 쿼리를 and 로 검색할지 or 로 검색할지 결정한다.
	expression := "$and"
	for _, word := range strings.Split(op.Searchword, " ") {
		if word == "or" || word == "||" {
			expression = "$or"
		}
	}
	queries := []bson.M{
		bson.M{expression: wordQueries},
	}
	// 상태 쿼리가 존재하면 상태에 대해서 or 처리한다.
	if len(statusQueries) != 0 {
		queries = append(queries, bson.M{"$or": statusQueries})
	}
	q := bson.M{"$and": queries}
	if *flagDebug {
		fmt.Println("검색에 사용한 쿼리리스트")
		fmt.Println(q)
		fmt.Println()
	}
	// 정렬설정
	switch op.Sortkey {
	// 스캔길이, 스캔날짜는 역순으로 정렬한다.
	// 스캔길이는 보통 난이도를 결정하기 때문에 역순(긴 길이순)을 매니저인 팀장,실장은 우선적으로 봐야한다.
	// 스캔날짜는 IO팀에서 최근 등록한 데이터를 많이 검토하기 때문에 역순(최근등록순)으로 봐야한다.
	case "scanframe", "scantime":
		op.Sortkey = "-" + op.Sortkey
	case "taskdate":
		if len(selectTasks) != 0 {
			op.Sortkey = "tasks." + op.Task + ".date"
		}
	case "taskpredate":
		if len(selectTasks) != 0 {
			op.Sortkey = "tasks." + op.Task + ".predate"
		}
	case "": // 기본적으로 id로 정렬한다.
		op.Sortkey = "id"
	}
	err = c.Find(q).Sort(op.Sortkey).All(&results)
	if err != nil {
		log.Println("DB Find Err : ", err)
		return nil, err
	}
	return results, nil
}

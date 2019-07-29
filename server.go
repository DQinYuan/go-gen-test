package randgen

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)


/*
container env

ConfPath = "/root/conf"
RmPath = "/root/randgen"
ResultPath = "/root/result"
*/
const CONFPATH = "CONFPATH"
const RMPATH = "RMPATH"
const RESULTPATH = "RESULTPATH"

// yy zz storage dir
var ConfPath = os.Getenv(CONFPATH)
// randgen main path
var RmPath = os.Getenv(RMPATH)
// result file path
var ResultPath = os.Getenv(RESULTPATH)

type Server struct {
	Db        *sql.DB
	DbiPrefix string
	DefaultZz string
}

func (this *Server) Listen(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/loaddata", this.loadData)
	err := http.ListenAndServe(":"+strconv.Itoa(port), mux)
	if err != nil {
		log.Fatalln(err)
	}
}

func (this *Server) loadData(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	payload := LoadDataRequest{}
	json.NewDecoder(r.Body).Decode(&payload)

	fmt.Println(payload.Zz)
	fmt.Println(payload.Yy)

	tmpZz := filepath.Join(ConfPath, fmt.Sprintf("%s.zz", payload.DB))
	tmpDb := payload.DB + "_tmp"

	g := &generator{
		db:      this.Db,
		perlDbi: this.DbiPrefix + tmpDb,
		tmpDb:   tmpDb,
		tmpZz:   tmpZz,
		tmpYy:   filepath.Join(ConfPath, fmt.Sprintf("%s.yy", payload.DB)),
		tmpRes:  filepath.Join(ResultPath, fmt.Sprintf("%s.sql", payload.DB)),
	}

	var zzContent string
	if payload.Zz != "" { zzContent = payload.Zz } else { zzContent = this.DefaultZz }

	sqls, err := g.LoadSqls(payload.Yy, zzContent, payload.Queries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write(MustJosnMarshal(&LoadDataResponse{SQLs: sqls}))
}

func MustJosnMarshal(v interface{}) []byte {
	bytes, _ := json.Marshal(v)
	return bytes
}

type LoadDataRequest struct {
	Yy      string `json:"yy"`
	Zz      string `json:"zz"`
	DB      string `json:"Db"`
	Queries int    `json:"queries"`
}

type LoadDataResponse struct {
	SQLs []string `json:"sql"`
}

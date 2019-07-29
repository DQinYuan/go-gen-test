package randgen

import (
	"database/sql"
	"fmt"
	"github.com/pingcap/errors"
	"github.com/dqinyuan/go-mysqldump"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)


type generator struct {
	db        *sql.DB
	perlDbi   string
	tmpDb     string
	tmpZz     string
	tmpYy     string
	// temp file to storage sqls, without '.sql' suffix
	tmpRes    string
}

func (this *generator) createDatabase() error {
	tmpls := []string{
		"drop database if exists %s",
		"create database %s",
		"use %s",
	}

	for _, tmpl := range tmpls {
		_, err := this.db.Exec(fmt.Sprintf(tmpl, this.tmpDb))
		if err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (this *generator) genTableData() error {
	log.Printf("# Table Grammar File %s\n", this.tmpZz)
	cmd := exec.Command("perl", "gentest.pl",
		"--dsn", this.perlDbi,
		"--gendata", this.tmpZz, "--grammar", " ")

	cmd.Dir = RmPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	// 因为在perl命令行中没有指定yy 文件,所以返回2是正常情况
	if err != nil && err.Error() != "exit status 2" {
		return errors.Trace(err)
	}

	return nil
}

func (this *generator) dumpDataBase() error  {
	err := DeleteFileIfExist(this.tmpRes)
	if err != nil {
		return errors.Trace(err)
	}

	dumpDir := "/"
	dumper, err := mysqldump.Register(this.db, dumpDir, this.tmpRes)
	if err != nil {
		return errors.Trace(err)
	}

	dumpFilename, err := dumper.Dump(true)
	if err != nil {
		return errors.Trace(err)
	}

	log.Printf("dump successfully to %s\n", dumpFilename)
	return nil
}

func (this *generator) genQueries(queries int) (*os.File,  error)  {
	log.Printf("# Query Grammar File %s\n", this.tmpYy)

	appendFile, err := os.OpenFile(this.tmpRes, os.O_APPEND | os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("perl", "gensql.pl",
		"--dsn", this.perlDbi,
		"--grammar", this.tmpYy,
		"--queries", strconv.Itoa(queries))
	cmd.Dir = RmPath
	cmd.Stdout = appendFile
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return appendFile, nil
}

func (this *generator) LoadSqls(yyContent string, zzContent string, queries int) ([]string, error) {
	if strings.TrimSpace(yyContent) == "" {
		return nil, errors.New("yy can not be empty")
	}
	// dump yy content to tmpYy file
	err := ioutil.WriteFile(this.tmpYy, []byte(yyContent), 0666)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if strings.TrimSpace(zzContent) != "" {
		// dump zz content to tmpZz file
		err := ioutil.WriteFile(this.tmpZz, []byte(zzContent), 0666)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	err = this.createDatabase()
	if err != nil {
		return nil, err
	}

	err = this.genTableData()
	if err != nil {
		return nil, err
	}

	err = this.dumpDataBase()
	if err != nil {
		return nil, err
	}

	f, err := this.genQueries(queries)
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, errors.Trace(err)
	}

	sqlContent, err := ioutil.ReadAll(f)

	sqls := strings.Split(string(sqlContent), ";\n")

	res := make([]string, 0)
	for _, sql := range sqls {
		trimed := strings.TrimSpace(sql)
		// delete gensql prefix
		if strings.HasPrefix(trimed, "#") {
			i := strings.LastIndex(trimed, "\n")
			trimed = trimed[i + 1:]
		}
		if trimed != "" {
			res = append(res, trimed)
		}
	}

	return res, nil
}

func FileIsExist(path string) bool {
	var exist = true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// 删除文件
func DeleteFileIfExist(path string) error {
	if FileIsExist(path) {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

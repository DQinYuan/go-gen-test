package randgen

import (
	"fmt"
	"github.com/monitoring-system/dbtest/util"
	"path/filepath"
	"testing"
)

func TestGenerator(t *testing.T) {
	t.SkipNow();
	dsn := "root:123456@tcp(127.0.0.1:3306)/?charset=utf8&parseTime=True&loc=Local"
	db, err := util.OpenDBWithRetry("mysql", dsn)
	if err != nil {
		t.Fatalf("Db open error %v\n", err)
	}
	defer db.Close()

	tmpZz, _ := filepath.Abs("tmp.zz")
	tmpYy, _ := filepath.Abs("tmp.yy")
	tmpRes, _ := filepath.Abs("res.sql")

	g := &generator{
		db:      db,
		perlDbi: "dbi:mysql:host=127.0.0.1:port=3306:user=root:database=testdb:password=123456",
		tmpDb:   "testdb",
		tmpZz:   tmpZz,
		tmpYy:   tmpYy,
		tmpRes:  tmpRes,
	}

	ConfPath = "."
	ResultPath = "."
	RmPath = "/home/dqyuan/language/Mysql/randgen-2.2.0"

	yyContent := `
query:
 	update | insert | delete ;

update:
 	UPDATE _table SET _field = digit WHERE condition LIMIT _digit ;

delete:
	DELETE FROM _table WHERE condition LIMIT _digit ;

insert:
	INSERT INTO _table ( _field ) VALUES ( _digit ) ;

condition:
 	_field < digit | _field = _digit ;
`
	zzContent := `
$tables = {
        rows => [0, 1, 10, 100],
        partitions => [ undef , 'KEY (pk) PARTITIONS 2' ]
};

$fields = {
        types => [ 'int', 'char', 'enum', 'set' ],
        indexes => [undef, 'key' ],
        null => [undef, 'not null'],
        default => [undef, 'default null'],
        sign => [undef, 'unsigned'],
        charsets => ['utf8', 'latin1']
};

$data = {
        numbers => [ 'digit', 'null', undef ],
        strings => [ 'letter', 'english' ],
        blobs => [ 'data' ],
	temporals => ['date', 'year', 'null', undef ]
}
`
	sqls, err := g.LoadSqls(yyContent, zzContent, 11)
	if err != nil {
		t.Fatalf("test error %+v\n", err)
	}


	for _, sql := range sqls {
		fmt.Println("=========================================")
		fmt.Println(sql)
	}
}
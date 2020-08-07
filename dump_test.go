package dbunit

import (
	"fmt"
	"os"
	"testing"

	"github.com/ilibs/gosql/v2"
)

func TestDumpSQL(t *testing.T) {
	if os.Getenv("DEV_DATABASE_HOST") != "" {
		db, err := gosql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", os.Getenv("DEV_DATABASE_USERNAME"), os.Getenv("DEV_DATABASE_PASSWORD"), os.Getenv("DEV_DATABASE_HOST"), "example")+"?charset=utf8&parseTime=True&loc=Asia%2FShanghai")
		if err != nil {
			t.Fatal("db open error:", err)
		}
		data, err := Dump(db, "testdata/fixtures/documents.yml", "select * from documents limit 10")
		if err != nil {
			t.Fatal("dump documents error:", err)
		}

		userIds := Pluck(data, "user_id")

		_, err = Dump(db, "testdata/fixtures/users.yml", "select * from users where id in(?)", userIds)
		if err != nil {
			t.Fatal("dump users error:", err)
		}

		_, err = Dump(db, "testdata/fixtures/members.yml", "select * from members where id = 0")
		if err != nil {
			t.Fatal("dump users error:", err)
		}
	}
}

func Test_getPrimaryKey(t *testing.T) {
	if os.Getenv("DEV_DATABASE_HOST") != "" {
		db, err := gosql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8&parseTime=True&loc=%s",
			os.Getenv("DEV_DATABASE_USERNAME"),
			os.Getenv("DEV_DATABASE_PASSWORD"),
			os.Getenv("DEV_DATABASE_HOST"), "example", "Asia%2FShanghai"),
		)
		pk, err := getPrimaryKey(db, "select * from users limit 1")

		if err != nil {
			t.Fatal("getPrimaryKey error", err)
		}

		if pk != "id" {
			t.Fatal("getPrimaryKey error must get id")
		}
	}

}

func Test_parseTableName(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "not where",
			args: args{
				"select * from users",
			},
			want: "users",
		},
		{
			name: "where",
			args: args{
				"select * from users where user_id in(1) limit 10",
			},
			want: "users",
		},
		{
			name: "table as",
			args: args{
				"select * from users as u where user_id in(1) limit 10",
			},
			want: "users",
		},
		{
			name: "table as",
			args: args{
				"select * from `users` as u where user_id in(1) limit 10",
			},
			want: "users",
		},
		{
			name: "table as",
			args: args{
				"select * from users u where user_id in(1) limit 10",
			},
			want: "users",
		},
		{
			name: "table as",
			args: args{
				"select * from users u",
			},
			want: "users",
		},
		{
			name: "table as",
			args: args{
				"select * from users as u",
			},
			want: "users",
		},
		{
			name: "not match",
			args: args{
				"select * users",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTableName(tt.args.query); got != tt.want {
				t.Errorf("parseTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

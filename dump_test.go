package dbunit

import (
	"testing"

	"github.com/ilibs/gosql/v2"
)

func TestDumpSQL(t *testing.T) {
	Run(t, "testdata/schema.sql", func(t *testing.T, db *gosql.DB) {
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
	})
}

func Test_getPrimaryKey(t *testing.T) {
	Run(t, "testdata/schema.sql", func(t *testing.T, db *gosql.DB) {
		pk, err := getPrimaryKey(db, "select * from users limit 1")

		if err != nil {
			t.Fatal("getPrimaryKey error", err)
		}

		if pk != "id" {
			t.Fatal("getPrimaryKey error must get id")
		}
	})

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
		{
			name: "table is task",
			args: args{
				"select * from task",
			},
			want: "task",
		},
		{
			name: "table is task_push",
			args: args{
				"select * from task_push where user_id = 1",
			},
			want: "task_push",
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

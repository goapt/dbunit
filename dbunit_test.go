package dbunit

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ilibs/gosql/v2"
)

type Users struct {
	Id           int       `json:"id" db:"id"`                                                    // 用户ID
	UserName     string    `json:"user_name" db:"user_name"`                                      // 用户名，用于展示
	Email        string    `json:"email" db:"email"`                                              // 邮箱
	RealName     string    `json:"real_name" db:"real_name"`                                      // 真实姓名
	Password     string    `json:"password" db:"password"`                                        // 密码
	Avatar       string    `json:"avatar" db:"avatar"`                                            // 用户头像
	Status       int       `json:"status" db:"status"`                                            // 1 启用 2停用
	About        string    `json:"about" db:"about"`                                              // 个人简介
	Role         string    `json:"role" db:"role"`                                                // 用户角色admin,leader,user
	Organization string    `json:"organization" db:"organization"`                                // 部门组织
	CreatedAt    time.Time `json:"created_at" db:"created_at"  time_format:"2006-01-02 15:04:05"` //
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"  time_format:"2006-01-02 15:04:05"` //

}

func (*Users) TableName() string {
	return "users"
}

func (*Users) PK() string {
	return "id"
}

func TestRun(t *testing.T) {
	t.Run("default fixtures", func(t *testing.T) {
		Run(t, "testdata/schema.sql", func(t *testing.T, db *gosql.DB) {
			user := &Users{
				Id: 1,
			}

			err := db.Model(user).Get()

			if err != nil {
				t.Fatal(err)
			}

			if user.Email != "test@test.cn" {
				t.Fatalf("user mismatch want %s,but get %s", "test@test.cn", user.Email)
			}
		})
	})

	t.Run("select fixtures", func(t *testing.T) {
		Run(t, "testdata/schema.sql", func(t *testing.T, db *gosql.DB) {
			user := &Users{
				Id: 1,
			}

			err := db.Model(user).Get()
			if err != sql.ErrNoRows {
				t.Fatal(err)
			}
		}, "testdata/fixtures/actions.yml", "testdata/fixtures/shares.yml")
	})

	t.Run("custom fixtures", func(t *testing.T) {
		Run(t, "testdata/schema.sql", func(t *testing.T, db *gosql.DB) {
			var ct int
			err := db.QueryRowx("select count(1) from wx__user").Scan(&ct)

			if err != nil {
				t.Fatal(err)
			}

			if ct == 0 {
				t.Fatalf("wx_user mismatch want %s,but get %d", " > 0", ct)
			}
		}, "testdata/inter")
	})
}

func TestNew(t *testing.T) {
	New(t, func(d *DBUnit) {
		db := d.NewDatabase("testdata/schema.sql", "testdata/fixtures/users.yml")
		// more database
		_ = d.NewDatabase("testdata/schema.sql")
		user := &Users{
			Id: 1,
		}

		err := db.Model(user).Get()

		if err != nil {
			t.Fatal(err)
		}

		if user.Email != "test@test.cn" {
			t.Fatalf("user mismatch want %s,but get %s", "test@test.cn", user.Email)
		}
	})
}

func TestLoad(t *testing.T) {
	// SetDatabase("root:123456@tcp(127.0.0.1:33306)/")
	test := NewTest("testdata/schema.sql")
	t.Cleanup(func() {
		test.Drop()
	})

	test.Load("testdata/custom")
}

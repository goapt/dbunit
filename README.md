# 数据库单元测试辅助
<a href="https://github.com/goapt/dbunit/actions"><img src="https://github.com/goapt/dbunit/workflows/build/badge.svg" alt="Build Status"></a>
<a href="https://codecov.io/gh/goapt/dbunit"><img src="https://codecov.io/gh/goapt/dbunit/branch/master/graph/badge.svg" alt="codecov"></a>
<a href="https://goreportcard.com/report/github.com/goapt/dbunit"><img src="https://goreportcard.com/badge/github.com/goapt/dbunit" alt="Go Report Card
"></a>
<a href="https://pkg.go.dev/github.com/goapt/dbunit"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square" alt="GoDoc"></a>
<a href="https://opensource.org/licenses/mit-license.php" rel="nofollow"><img src="https://badges.frapsoft.com/os/mit/mit.svg?v=103"></a>

## 使用

```shell script
go get github.com/goapt/dbunit
```

1、单数据库自动辅助测试函数

```go
dbunit.Run(t, "testdata/schema.sql", func(t *testing.T, db *gosql.DB) {
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
```

> dbunit.Run 会自动创建数据库，并且自动导入`testdata/fixtures`目录下面的测试数据

2、多数据库调用

```go
dbunit.New(t, func(d *DBUnit) {
    db := d.NewDatabase("testdata/schema.sql","testdata/fixtures/users.yml")
    // more database
    db2 = d.NewDatabase("testdata/schema2.sql")
    user := &Users{
        Id: 1,
    }

    err := db.Model(user).Get()
    .....
})
```

## 从测试库导出测试数据文件

### 使用脚本导出数据
```go
db, err := gosql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", os.Getenv("DEV_DATABASE_USERNAME"), os.Getenv("DEV_DATABASE_PASSWORD"), os.Getenv("DEV_DATABASE_HOST"), "example")+"?charset=utf8&parseTime=True&loc=Asia%2FShanghai")
if err != nil {
    t.Fatal("db open error:", err)
}
// 导出前10个文档信息
data, err := dbunit.Dump(db, "testdata/fixtures/documents.yml", "select * from documents limit 10")
if err != nil {
    t.Fatal("dump documents error:", err)
}

// 从结果集中获取用户ID
userIds := dbunit.Pluck(data, "user_id")

//  导出所有相关的用户
_, err = dbunit.Dump(db, "testdata/fixtures/users.yml", "select * from users where id in(?)", userIds)
if err != nil {
    t.Fatal("dump users error:", err)
}
```

> 如果导出的数据已经在数据集文件中存在，则会忽略，判断依据为主键一致


### 导出测试集
你可以在参看 `testdata/fixture.go` 脚本编写测试集导出脚本

```go
// 文章数据集
data, err := dbunit.Dump(db, "testdata/fixtures/articles.yml",
    "select * from articles where doc_id in(1,2)",
)
if err != nil {
    panic(err)
}
```

## 关于fixture

当数据中需要动态数据，比如时间，可以参考如下做法，或者贡献你要的templateFunc
```yaml
- id: 1
  created_at: {{now}}
  updated_at: {{now}}
```



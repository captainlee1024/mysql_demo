// Package main provides ...
package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// 定义一个全局对象db
var db *sql.DB

// 定义一个初始化数据库的函数
func initMySQL() (err error) {
	// DNS: Data Source Name
	dsn := "root:644315@tcp(127.0.0.1:3306)/go_test"

	// Open 函数只会验证参数的格式是否正确，实际上并不创建用户数据库连接，不会校验账号密码是否正确。
	// 如果要检查数据源的名称是否真实有效，应该调用Ping方法
	// 注意！全局变量的初始化，使用=而不是声明一个db变量(使用:=)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	// 做完错误检查之后，确保db不为nil
	// 尝试与数据库建立连接(校验dsn是否正确)
	err = db.Ping()
	if err != nil {
		fmt.Printf("connect to db failed, err: %v\n", err)
		return err
	}

	// 数值根据业务情况来定
	//db.SetConnMaxLifetime(time.Second * 10) // 连接存活时间　10秒钟
	//db.SetMaxOpenConns(1) // 最大连接数
	//db.SetMaxIdleConns(10)                  // 最大空闲连接数
	return
}

// model
type user struct {
	id   int
	age  int
	name string
}

// 查询

// 查询单条数据
func queryRowDemo() {
	sqlStr := `select id, name, age from mysql_demo_user where id = ?`
	var u user
	// db.QueryRow()执行一次查询，并期望返回最多一行结果(即Row).
	// QueryRow总是返回非nil的值，直到返回值的Scan方法被调用时才返回被延迟的错误。
	// 注意！！！确保QueryRow之后调用Scan方法，否则持有的数据库连接不会被释放
	// 在Scan方法里面会调用rows.Close()进行关闭连接
	// 否则连接会一直随着row变量一直占用着
	//row := db.QueryRow(sqlStr, 20) // 返回row对象,查询出来一行

	// 如果我们这里把最大连接数设置为1，然后在第一条查新scan之前再次查询程序就会卡住
	// 连接池只有一个连接，一直在等待连接池里的连接，上面的连接一直被占用没有手动释放
	//row = db.QueryRow(sqlStr, 21)
	//err := row.Scan(&u.id, &u.name, &u.age) // 通过Scan一个一个字段扫描出来赋值给结构体的字段

	// 为了防止忘记在查询之后立即调用Scan()，一般会采用下面这种写法
	err := db.QueryRow(sqlStr, 21).Scan(&u.id, &u.name, &u.age)
	if err != nil {
		fmt.Printf("scan fialed, err:%v\n", err)
		return
	}
	fmt.Printf("id:%d name:%s age:%d\n", u.id, u.name, u.age)
}

// 多行查询
func queryMultiRowDemo() {
	sqlStr := `
	select id, name, age
	from mysql_demo_user
	where id > ?`
	rows, err := db.Query(sqlStr, 0)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return
	}
	// 注意！！！关闭rows释放持有的数据库连接
	defer rows.Close()

	// 循环读取结果集中的数据
	for rows.Next() {
		var u user
		err := rows.Scan(&u.id, &u.name, &u.age)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return
		}
		fmt.Printf("id:%d name:%s age:%d\n", u.id, u.name, u.age)
	}
}

// 插入数据
// 插入、更新和操作都使用Exec方法
func insertRowDemo() {
	sqlStr := "insert into mysql_demo_user(name, age) values(?,?)"
	ret, err := db.Exec(sqlStr, "小李", 22)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	// 上面已经有一个err，这里虽好先声明theId，不要使用声明并赋值符号
	var theId int64
	theId, err = ret.LastInsertId() // 新插入数据的ID
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("insert success, the id is %d\n", theId)
}

// 更新数据
func updateRowDemo() {
	sqlStr := `
	update mysql_demo_user
	set age = ?
	where id = ?`

	ret, err := db.Exec(sqlStr, 25, 20)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return
	}
	var n int64
	n, err = ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("udpate success, affected rows:%d\n", n)
}

// 删除数据
func deleteRowDemo() {
	sqlStr := `
	delete from mysql_demo_user
	where id = ?`
	ret, err := db.Exec(sqlStr, 20)
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return
	}

	var n int64
	n, err = ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("delete success, affected rows:%d\n", n)
}

// 预处理查询
func prepareQueryDemo() {
	sqlStr := `select id, name, age from mysql_demo_user where id > ?`
	stmt, err := db.Prepare(sqlStr) // 返回一个状态
	if err != nil {
		fmt.Printf("prepare failed, err:%v\n", err)
		return
	}
	// 在err非空判断之后执行一个延迟关闭，释放连接
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query(0)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return
	}

	// 在判断之后延迟调用Close()
	defer rows.Close()

	// 循环读取结果集中的数据
	for rows.Next() {
		var u user
		err := rows.Scan(&u.id, &u.name, &u.age)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return
		}
		fmt.Printf("id:%d name:%s age:%d\n", u.id, u.name, u.age)
	}
}

// 预处理插入
func prepareInsertDemo() {
	sqlStr := `insert into mysql_demo_user(name, age) values(?,?)`
	stmt, err := db.Prepare(sqlStr) // 返回一个状态供查询
	if err != nil {
		fmt.Printf("prepare failed, err:%v\n", err)
		return
	}

	// 判断之后延迟调用一个Close()
	defer stmt.Close()

	_, err = stmt.Exec("小虎", 20)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}

	var ret sql.Result
	ret, err = stmt.Exec("胖虎", 10)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}

	var theId int64
	theId, err = ret.LastInsertId()
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("insert success, the id is %d\n", theId)
}

// sql注入事例
func sqlInjectDemo(name string) {
	sqlStr := fmt.Sprintf(
		`select id, name, age
	from mysql_demo_user
	where name='%s'`, name)
	fmt.Printf("SQL:%v\n", sqlStr)
	var u user
	err := db.QueryRow(sqlStr).Scan(&u.id, &u.name, &u.age)
	if err != nil {
		fmt.Printf("exec fialed, err:%v\n", err)
		return
	}
	fmt.Printf("user:%#v\n", u)
}

// 事务示例
func transactionDemo() {
	// 开启事务
	tx, err := db.Begin()

	// 判断小步骤是否成功决定是否回滚
	if err != nil {
		if tx != nil { // 判空以下，确保tx有值之后在调用回滚函数
			tx.Rollback() // 回滚
		}
		fmt.Printf("begin trans fialed, err:%v\n", err)
		return
	}

	sqlStr1 := `update mysql_demo_user set age = 30 where id = ?`
	ret1, err := tx.Exec(sqlStr1, 22)
	if err != nil {
		tx.Rollback() // 回滚
		fmt.Printf("exec sql1 failed, err:%v\n", err)
		return
	}
	affRow1, err := ret1.RowsAffected()
	if err != nil {
		tx.Rollback() // 回滚
		fmt.Printf("get affRow1 failed, err:%v\n", err)
		return
	}
	fmt.Printf("affRow1:%d\n", affRow1)

	sqlStr2 := `update mysql_demo_user set age = 30 where id = ?`
	// 数据库中没有ID为1的记录，下面返回受影响行数为0，最后事务会回滚
	ret2, err := tx.Exec(sqlStr2, 1)
	// 数据库有该条记录，两次修改影响行数都是1，最后事务会正常提交
	//ret2, err := tx.Exec(sqlStr2, 25)
	if err != nil {
		tx.Rollback() // 回滚
		fmt.Printf("exec sql2 failed, err:%v\n", err)
		return
	}
	affRow2, err := ret2.RowsAffected()
	if err != nil {
		tx.Rollback() // 回滚
		fmt.Printf("get affRow2 failed, err:%v\n", err)
		return
	}
	fmt.Printf("affRow2:%d\n", affRow2)

	// 提交事务
	// 当上面两个受影响行affRow都是1的时候说明两次更新都成功了，这里再提交事务
	if affRow1 == 1 && affRow2 == 1 {
		err := tx.Commit()
		if err != nil {
			tx.Rollback() // 回滚
			fmt.Printf("commit failed, err:%v\n", err)
			return
		}
	} else {
		tx.Rollback()
		fmt.Println("事务回滚啦...")
		return
	}
	fmt.Println("exec trans success!")
}

func main() {
	err := initMySQL() // 调用初始化数据库的函数
	if err != nil {
		fmt.Printf("connect to dv failed, err:%v\n", err)
		return
	}

	// Close() 用来释放掉数据库连接相关的资源
	defer db.Close() //注意这行代码要写在上面的err判断的下面

	fmt.Println("connect to db success")
	// TODO: CRUD <09-09-20, CaptainLee1024> //
	//queryRowDemo()
	//queryMultiRowDemo()
	//insertRowDemo()
	//updateRowDemo()
	//deleteRowDemo()

	// TODO: 预处理 <09-09-20, CaptainLee1024> //
	//prepareQueryDemo()
	//prepareInsertDemo()

	// TODO: SQL注入示例 <09-09-20, CaptainLee1024> //
	// 输入以下字符串会引发SQL注入问题"xxx' or 1=1#"
	// "xxx' union select * from mysql_demo_user #"
	// "xxx and (select count(*) from mysql_demo_user) < 10 #"
	//sqlInjectDemo("xxx' or 1=1#")

	// TODO: Go实现MySQL事务示例 <09-09-20, CaptainLee1024> //
	//transactionDemo()
}

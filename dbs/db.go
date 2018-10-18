package dbs

import (
	"database/sql"
	"fmt"
	"go-copyright/configs"
	_ "strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DBConn *sql.DB

type (
	Account struct {
		ID         int    `json:"id"`
		Email      string `json:"email"`
		Username   string `json:"username"`
		IdentityID string `json:"identity_id"`
		Address    string `json:"address"`
	}
	Content struct {
		ContentID    int       `json:"content_id"`
		Title        string    `json:"title"`
		Content      string    `json:"content"`
		Content_hash string    `json:"content_hash"`
		Ts           time.Time `json:"ts"`
	}
	AccountContent struct {
		Content_hash string    `json:"content_hash"`
		TokenID      int       `json:"token_id"`
		Address      string    `json:"address"`
		Ts           time.Time `json:"ts"`
	}
	Auction struct {
		Content_hash string `json:"content_hash"`
		TokenID      int    `json:"token_id"`
		Address      string `json:"address"`
		Percent      int    `json:"percent"`
		Price        int    `json:"price"`
		Status       int    `json:"status"`
	}
	Vote struct {
		Address      string `json:"address"`
		Content_hash string `json:"content_hash"`
		Comment      string `json:"comment"`
	}
)

func init() {
	fmt.Println("call dbs.Init", configs.Config)
	DBConn = InitDB(configs.Config.Db.ConnStr, configs.Config.Db.Driver)
}

func InitDB(connstr, Driver string) *sql.DB {
	db, err := sql.Open(Driver, connstr)
	if err != nil {
		panic(err.Error())
	}

	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return db
}

//通用查询，返回map嵌套map
func DBQuery(sql string) ([]map[string]string, int, error) {
	fmt.Println("query is called:", sql)
	rows, err := DBConn.Query(sql)
	if err != nil {
		fmt.Println("query data err", err)
		return nil, 0, err
	}
	//得到列名数组
	cols, err := rows.Columns()
	//获取列的个数
	colCount := len(cols)
	values := make([]string, colCount)
	oneRows := make([]interface{}, colCount)
	for k, _ := range values {
		oneRows[k] = &values[k] //将查询结果的返回地址绑定，这样才能变参获取数据
	}
	//存储最终结果
	results := make([]map[string]string, 1)
	idx := 0
	for rows.Next() {
		rows.Scan(oneRows...)
		rowmap := make(map[string]string)
		for k, v := range values {
			rowmap[cols[k]] = v

		}
		if idx > 0 {
			results = append(results, rowmap)
		} else {
			results[0] = rowmap
		}
		idx++
		//fmt.Println(values)
	}
	//fmt.Println("---------------------------------------")
	fmt.Println("query..idx===", idx)
	return results, idx, nil

}
func Create(sql string) (int64, error) {
	res, err := DBConn.Exec(sql)
	if err != nil {
		fmt.Println("exec sql err,", err, "sql is ", sql)
		return -1, err
	}
	return res.LastInsertId()
}

func (ctx *Content) AddContent() (int64, error) {
	fmt.Println("hash==", ctx.Content_hash, "title=", ctx.Title, len(ctx.Content))
	res, err := DBConn.Exec("insert into content(title,content,content_hash) values(?,?,?)",
		ctx.Title,
		ctx.Content,
		ctx.Content_hash,
	)
	if err != nil {
		fmt.Println("insert into content err", err)
		return -1, err
	}
	id, err := res.LastInsertId()
	ctx.ContentID = int(id)
	fmt.Println(" ctx.ContentID===", ctx.ContentID)
	return res.LastInsertId()
}

func (ctx *Content) GetContent() error {
	row := DBConn.QueryRow("select title,content from content where title=?", ctx.Title)
	if row == nil {
		fmt.Println("select content err")
		return nil
	}
	//ctx.Content = make([]byte, 19677)
	err := row.Scan(&ctx.Title, &ctx.Content)
	fmt.Println(err, ctx.Title, ctx.Content)
	return err
}

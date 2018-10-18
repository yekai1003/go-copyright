package routes

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"go-copyright/dbs"
	"go-copyright/eths"
	"go-copyright/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
)

const PAGE_MAX_PICTURES int = 5

func PingHandler(c echo.Context) error {

	return c.String(http.StatusOK, "pong")
}

//创建账户
func CreateAccount(c echo.Context) error {
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//解析账户信息
	account := &dbs.Account{}
	if err := c.Bind(account); err != nil {
		resp.Errno = utils.RECODE_DATAERR
		//resp.ErrMsg = RecodeText(resp.Errno)
		return err
	}
	//创建以太坊账号
	accid, err := eths.NewAcc(account.IdentityID)
	if err != nil {
		resp.Errno = utils.RECODE_IPCERR
		return err
	}
	//插入到数据库中
	sql := fmt.Sprintf("insert into account(email,username,identity_id,address) values('%s','%s','%s','%s')", account.Email, account.Username, account.IdentityID, accid)
	//dbs.Create("insert into account(email,username,identity_id) values('yekai@','yekai','123')")
	n, err := dbs.Create(sql)
	if err != nil {
		resp.Errno = utils.RECODE_DBERR
		return err
	}

	//session处理
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values["account_id"] = n
	sess.Values["address"] = accid
	sess.Save(c.Request(), c.Response())

	return nil
}

//判断用户是否登录
func GetSession(c echo.Context) error {
	//通过session来判断是否可以登陆
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	sess, err := session.Get("session", c)
	fmt.Println("getsession===", err)

	accid := sess.Values["account_id"]
	fmt.Println()
	//fmt.Println("getsession.....", accid, ok)
	if accid == nil {
		resp.Errno = utils.RECODE_LOGINERR
		return nil
	}
	fmt.Println("GetSession======", accid)
	mapResp := make(map[string]interface{})
	mapResp["account_id"] = accid
	mapResp["address"] = sess.Values["address"]
	resp.Data = mapResp
	return nil
}

//登陆
func Login(c echo.Context) error {
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//未登录需要检查数据库内容
	//解析请求json
	account := &dbs.Account{}
	if err := c.Bind(account); err != nil {
		resp.Errno = utils.RECODE_DATAERR
		return err
	}
	//查询数据库
	//select * from account where username='yekai23' and identity_id = '123';
	sql := fmt.Sprintf("select * from account where username='%s' and identity_id = '%s'", account.Username, account.IdentityID)
	m, n, err := dbs.DBQuery(sql)
	if err != nil || n <= 0 {
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	fmt.Println(m, n)
	data := m[0]
	//写入session
	sess, err := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	fmt.Println(data)
	sess.Values["account_id"] = data["account_id"]
	sess.Values["address"] = data["address"]
	fmt.Println(sess.Name())

	sess.Save(c.Request(), c.Response())

	return nil
}

//上传图片
func UploadPic(c echo.Context) error {
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	fmt.Println("begin UploadPic")
	//解析前端信息
	content := &dbs.Content{}
	h, err := c.FormFile("fileName")
	if err != nil {
		fmt.Println("Formfile err", err)
		resp.Errno = utils.RECODE_PARAMERR
		return err
	}

	//打开该文件
	src, err := h.Open()
	if err != nil {
		fmt.Println("Open file err", err)
		resp.Errno = utils.RECODE_IOERR
		return err
	}
	defer src.Close()

	//打开要保存的位置
	content.Content = "static/photo/" + h.Filename
	dst, err := os.Create(content.Content)
	if err != nil {
		fmt.Println("Create file err", err)
		resp.Errno = utils.RECODE_IOERR
		return err
	}
	defer dst.Close()

	cHash := make([]byte, h.Size)
	//src.Seek(0, os.SEEK_SET) //移动到文件头的位置
	n, err := src.Read(cHash)
	if err != nil || int64(n) != h.Size {
		resp.Errno = utils.RECODE_IOERR
		return err
	}
	content.Content_hash = fmt.Sprintf("%x", sha256.Sum256(cHash))
	fmt.Println("sha256.Sum256(cHash)===", sha256.Sum256(cHash), common.HexToHash(content.Content_hash))
	//content.Content_hash = common.ToHex(common.HexToHash(content.Content_hash))
	content.Title = h.Filename
	dst.Write(cHash) //写入文件

	fmt.Println("content:", content)
	//9178af53170dcdc7b280b7afcef54dd22295e335dfcb00330a58c8274e58721e
	//添加到数据库中
	_, err = content.AddContent()
	if err != nil {
		fmt.Println("failed to add content ", err)
		return err
	}
	//调用智能合约，发放奖励
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	fromAddr, ok := sess.Values["address"].(string)
	if fromAddr == "" || !ok {
		return errors.New("session not exists")
	}
	fmt.Println("end UploadPic")
	go eths.UploadPic(fromAddr, "yekai", content.Content_hash, content.Title)
	return err
}

//查询图片信息 - 根据图片hash
func GetContent(c echo.Context) error {
	content := &dbs.Content{}
	content.Title = c.Param("title")
	fmt.Println("queryContent is called,dna=", content.Title)
	//查询数据库
	err := content.GetContent()
	if err != nil {
		fmt.Println("queryContent query err", err)
		return err
	}

	//发送对应文件块
	http.ServeFile(c.Response(), c.Request(), content.Content)
	//	if err := SendFile(c, content.Content, content.Title); err != nil {
	//		return err
	//	}
	return nil
}

//查看当前用户所有图片 -- 接口
func GetAccountContent(c echo.Context) error {
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)

	sess, _ := session.Get("session", c)
	acctAddr, ok := sess.Values["address"].(string)
	if !ok {
		fmt.Println("session not exists", ok)
		resp.Errno = utils.RECODE_SESSIONERR
		return nil
	}

	//content.AccountID = accid
	fmt.Println("acctAddr===", acctAddr)

	sql := fmt.Sprintf("select title,a.content_hash,a.token_id from account_content a,content b where a.content_hash = b.content_hash and address='%s'", acctAddr)
	titles, n, err := dbs.DBQuery(sql)
	if err != nil {
		fmt.Println("query user content err", err)
		resp.Errno = utils.RECODE_DBERR
		return err
	}

	total_page := int(n)/PAGE_MAX_PICTURES + 1
	current_page := 1
	mapResp := make(map[string]interface{})
	mapResp["total_page"] = total_page
	mapResp["current_page"] = current_page
	contents := make([]interface{}, 1)

	for k, v := range titles {
		if k == 0 {
			contents[0] = v
		} else {
			contents = append(contents, v)
		}

	}
	mapResp["contents"] = contents
	resp.Data = mapResp

	return nil
}

//图片发起拍卖
func ContentAuction(c echo.Context) error {
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//得到拍卖资源的数据
	auction := &dbs.Auction{}
	sess, _ := session.Get("session", c)
	acctAddr, ok := sess.Values["address"].(string)
	if !ok {
		fmt.Println("session not exists", ok)
		resp.Errno = utils.RECODE_SESSIONERR
		return nil
	}
	auction.Address = acctAddr
	//获取请求json数据
	if err := c.Bind(auction); err != nil {
		resp.Errno = utils.RECODE_DATAERR
		fmt.Println("failed to bind data(ContentAution)", err)
		return err
	}
	fmt.Println("auction====", auction)
	if auction.Content_hash == "" {
		fmt.Println("request data err")
		resp.Errno = utils.RECODE_DATAERR
		return nil
	}
	//插入到数据库中
	sql := fmt.Sprintf("insert into auction(content_hash,address,token_id,percent,price,status) values('%s','%s',%d,%d,%d,0)",
		auction.Content_hash,
		auction.Address,
		auction.TokenID,
		auction.Percent,
		auction.Price,
	)
	if _, err := dbs.Create(sql); err != nil {
		fmt.Println("failed to insert auction,", err, sql)
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	return nil
}

//select title,token_id,price from content a,auction b where a.content_hash = b.content_hash and b.address = '0x5539054d29aa6986ae821c0b040b8e91dc0f0ce3' and status = 0

func GetAuctions(c echo.Context) error {
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//获得账户地址
	//auction := &dbs.Auction{}
	sess, _ := session.Get("session", c)
	acctAddr, ok := sess.Values["address"].(string)
	if !ok {
		fmt.Println("session not exists", ok)
		resp.Errno = utils.RECODE_SESSIONERR
		return nil
	}
	fmt.Println(acctAddr)
	//auction.Address = acctAddr
	//查询数据库
	sql := fmt.Sprintf("select title,token_id,price from content a,auction b where a.content_hash = b.content_hash and b.address <> '%s' and status = 0", acctAddr)
	//sql := fmt.Sprintf("select title,token_id,price from content a,auction b where a.content_hash = b.content_hash and status = 0")
	aucsData, _, err := dbs.DBQuery(sql)
	if err != nil {
		fmt.Println("query auction data err", err)
		resp.Errno = utils.RECODE_DBERR
		return nil
	}
	resp.Data = aucsData
	return err
}

//拍卖终结者
func BidAuction(c echo.Context) error {
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//获取参数
	price := c.QueryParam("price")
	tokenID := c.QueryParam("tokenid")
	if price == "" || tokenID == "" {
		fmt.Println("BidAuction get params err", price, tokenID)
		resp.Errno = utils.RECODE_PARAMERR
		return errors.New("params err")
	}
	//查询数据库
	sql := fmt.Sprintf("select * from auction where token_id = %s", tokenID)
	auctData, _, err := dbs.DBQuery(sql)
	if err != nil {
		fmt.Println("BidAuction:failed to query data from auction", sql, err)
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	data := auctData[0]
	fmt.Println(data)
	//修改数据库
	sql = fmt.Sprintf("update auction set status = 1, price=%s where token_id=%s", price, tokenID)
	if _, err = dbs.Create(sql); err != nil {
		fmt.Println("failed to update auction", sql, err)
		resp.Errno = utils.RECODE_DBERR
		return err
	}
	//调用以太坊 -- 这里出一个问题，对于边缘情况的处理
	//1.调用split分割资产
	//2.调用转账
	go func() {
		_tokenid, _ := strconv.ParseInt(tokenID, 10, 32)
		_price, _ := strconv.ParseInt(price, 10, 32)
		_weight, _ := strconv.ParseInt(data["weight"], 10, 32)
		_seller := data["address"]
		fmt.Println("call eth---begin---", _tokenid, _price, _weight, _seller)
		sess, _ := session.Get("session", c)
		_acctAddr, ok := sess.Values["address"].(string)
		if !ok {
			fmt.Println("session not exists", ok)
			resp.Errno = utils.RECODE_SESSIONERR
			return
		}
		err = eths.EthSplitAsset(_tokenid, _weight, _acctAddr)
		if err != nil {
			fmt.Println("failed to call EthSplitAsset", err)
			return
		}

		if err = eths.EthTransfer20(_acctAddr, "yekai", _seller, _price); err != nil {
			fmt.Println("failed to call EthTransfer20", err)
			return
		}
	}()
	return err
}

//投票
func ContentVote(c echo.Context) error {
	var resp utils.Resp
	resp.Errno = utils.RECODE_OK
	defer utils.ResponseData(c, &resp)
	//获取参数
	vote := &dbs.Vote{}
	if err := c.Bind(vote); err != nil {
		resp.Errno = utils.RECODE_DATAERR
		return err
	}
	fmt.Println(vote)
	//mysql数据库操作
	//1. 先查询是否超过投票时间
	//select * from content where content_hash ='db5cda1a1fb5a313743ff5e25c98e9187da1665d27e80954ffee2817ff1c41b3' and now()<=date_add(ts,interval 1 day);
	//2. 如果允许投票插入到mysql中

	//eth操作，投票时扣除erc20，转账到基金会账户
	//1. 从session获取账户address
	//2. 调用以太坊erc20转账函数
	return nil
}

package main

import (
	"fmt"

	"go-copyright/configs"
	"go-copyright/routes"

	"go-copyright/eths"
	_ "net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

var EchoObj *echo.Echo //echo框架对象全局定义

//静态文件处理
func staicFile() {
	EchoObj.Static("/", "static/pc/home")
	//EchoObj.Static("/register.html", "static/pc/home/index.html")
	EchoObj.Static("/static", "static")
	EchoObj.Static("/user", "static/pc/user")
	EchoObj.Static("/css", "static/pc/css")
	EchoObj.Static("/upload", "static/pc/upload")
	EchoObj.Static("/assets", "static/pc/assets")
}

func main() {
	fmt.Printf("get config %v \n", configs.Config.Common)
	EchoObj = echo.New()             //创建echo对象
	EchoObj.Use(middleware.Logger()) //安装日志中间件
	EchoObj.Use(middleware.Recover())
	//EchoObj.Use(session.MiddlewareWithConfig(session.Config{}))
	EchoObj.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	//	EchoObj.Use(middleware.GzipWithConfig(middleware.GzipConfig{
	//		Level: 5,
	//	}))
	staicFile() //静态文件处理
	//路由设置
	EchoObj.GET("/ping", routes.PingHandler)          //路由测试
	EchoObj.POST("/account", routes.CreateAccount)    //注册
	EchoObj.GET("/session", routes.GetSession)        //session
	EchoObj.POST("/login", routes.Login)              //登陆
	EchoObj.POST("/content", routes.UploadPic)        //上传图片
	EchoObj.GET("/content", routes.GetAccountContent) //查看用户图片
	EchoObj.GET("/content/:title", routes.GetContent) //查看具体图片
	EchoObj.POST("/auction", routes.ContentAuction)   //发起拍卖
	EchoObj.GET("/auctions", routes.GetAuctions)      //查看当前拍卖图片
	EchoObj.GET("/auction/bid", routes.BidAuction)    //拍卖处理
	//启动订阅
	go eths.EventSubscribe("ws://localhost:8546", configs.Config.Eth.PxaAddr)
	EchoObj.Logger.Fatal(EchoObj.Start(configs.Config.Common.Port)) //启动服务
}

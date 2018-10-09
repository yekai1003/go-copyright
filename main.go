package main

import (
	"fmt"

	"go-copyright-p1/configs"
	"go-copyright-p1/routes"

	_ "net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var config *configs.ServerConfig //引用配置文件结构
var EchoObj *echo.Echo           //echo框架对象全局定义

func main() {
	config = configs.GetConfig() //配置文件读取
	if config == nil {
		return
	}
	fmt.Printf("get config %v \n", config.Common)
	EchoObj = echo.New()             //创建echo对象
	EchoObj.Use(middleware.Logger()) //安装日志中间件
	EchoObj.Use(middleware.Recover())
	EchoObj.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	EchoObj.GET("/ping", routes.PingHandler)                //路由测试函数
	EchoObj.Logger.Fatal(EchoObj.Start(config.Common.Port)) //启动服务
}

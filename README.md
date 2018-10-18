# go-copyright-p1

### 目录介绍

-  static  html页面目录
-  configs 配置文件读取处理
-  routes  路由处理
-  etc   配置文件处理
-  utils 通用处理，错误信息
-  dbs   数据库处理文件
-  eths  以太坊相关处理

## 环境安装需要

### 配置文件读取插件需要安装

```
go get -u github.com/BurntSushi/toml
```

### echo框架安装

先安装crypto,labstack使用了该库，需要借助github，go get不能用
```
cd $GOPATH/src
mkdir -p golang.org/x/
cd golang.org/x/
git clone https://github.com/golang/crypto.git
```
安装echo
```
go get -u github.com/labstack/echo
go get -u github.com/labstack/echo-contrib/session
```



### 开发过程可能需要用到的库安装方法如下


mysql的go语言驱动安装
```
go get -u github.com/go-sql-driver/mysql
```

其他可能涉及的库
```
go get -u github.com/labstack/gommon/
go get -u github.com/dgrijalva/jwt-go
go get -u github.com/go-sql-driver/mysql
```

### echo框架学习资料

[echo框架学习](https://echo.labstack.com/guide)

### 数据库建库脚本

[建库语句](etc/copyright.sql)
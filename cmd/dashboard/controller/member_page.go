package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/naiba/nezha/pkg/mygin"
	"github.com/naiba/nezha/service/dao"
)

type memberPage struct {
	r *gin.Engine
}

func (mp *memberPage) serve() {
	mr := mp.r.Group("")
	mr.Use(mygin.Authorize(mygin.AuthorizeOption{
		Member:   true,
		IsPage:   true,
		Msg:      "此页面需要登录",
		Btn:      "点此登录",
		Redirect: "/login",
	}))
	mr.GET("/server", mp.server)
}

func (mp *memberPage) server(c *gin.Context) {
	dao.ServerLock.RLock()
	defer dao.ServerLock.RUnlock()
	c.HTML(http.StatusOK, "page/server", mygin.CommonEnvironment(c, gin.H{
		"Title":   "服务器管理",
		"Servers": dao.ServerList,
	}))
}

package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/naiba/com"

	"github.com/naiba/nezha/model"
	"github.com/naiba/nezha/pkg/mygin"
	"github.com/naiba/nezha/service/dao"
)

type memberAPI struct {
	r gin.IRouter
}

func (ma *memberAPI) serve() {
	mr := ma.r.Group("")
	mr.Use(mygin.Authorize(mygin.AuthorizeOption{
		Member:   true,
		IsPage:   false,
		Msg:      "访问此接口需要登录",
		Btn:      "点此登录",
		Redirect: "/login",
	}))

	mr.POST("/logout", ma.logout)
	mr.POST("/server", ma.addOrEditServer)
	mr.DELETE("/server/:id", ma.delete)
}

func (ma *memberAPI) delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id < 1 {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: "错误的 Server ID",
		})
		return
	}
	dao.ServerLock.Lock()
	defer dao.ServerLock.Unlock()
	if err := dao.DB.Delete(&model.Server{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("数据库错误：%s", err),
		})
		return
	}
	delete(dao.ServerList, strconv.FormatUint(id, 10))
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}

type serverForm struct {
	ID     uint64
	Name   string `binding:"required"`
	Secret string
}

func (ma *memberAPI) addOrEditServer(c *gin.Context) {
	admin := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var sf serverForm
	var s model.Server
	err := c.ShouldBindJSON(&sf)
	if err == nil {
		dao.ServerLock.Lock()
		defer dao.ServerLock.Unlock()
		s.Name = sf.Name
		s.Secret = sf.Secret
		s.ID = sf.ID
	}
	if sf.ID == 0 {
		s.Secret = com.MD5(fmt.Sprintf("%s%s%d", time.Now(), sf.Name, admin.ID))
		s.Secret = s.Secret[:10]
		err = dao.DB.Create(&s).Error
	} else {
		err = dao.DB.Save(&s).Error
	}
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求错误：%s", err),
		})
		return
	}
	dao.ServerList[fmt.Sprintf("%d", s.ID)] = &s
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}

type logoutForm struct {
	ID uint64
}

func (ma *memberAPI) logout(c *gin.Context) {
	admin := c.MustGet(model.CtxKeyAuthorizedUser).(*model.User)
	var lf logoutForm
	if err := c.ShouldBindJSON(&lf); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求错误：%s", err),
		})
		return
	}
	if lf.ID != admin.ID {
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("请求错误：%s", "用户ID不匹配"),
		})
		return
	}
	dao.DB.Model(admin).UpdateColumns(model.User{
		Token:        "",
		TokenExpired: time.Now(),
	})
	c.JSON(http.StatusOK, model.Response{
		Code: http.StatusOK,
	})
}

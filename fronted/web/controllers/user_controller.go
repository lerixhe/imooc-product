package controllers

import (
	"imooc-product/common"
	"imooc-product/datamodels"
	"imooc-product/encrypt"
	"imooc-product/services"
	"imooc-product/tool"
	"strconv"

	"github.com/kataras/golog"

	"github.com/kataras/iris/mvc"

	"github.com/kataras/iris"
)

type UserController struct {
	Ctx         iris.Context
	UserService services.IUserService
}

func (u *UserController) GetRegister() mvc.View {
	return mvc.View{
		Name: "user/register.html",
	}
}
func (u *UserController) PostRegister() {
	// 创建空user实例
	user := new(datamodels.User)
	// 解析前，需要转换请求里的表单，先给Form字段注入数据。
	u.Ctx.Request().ParseForm()
	// 创建解码器，设置tag
	decoder := common.NewDecoder(&common.DecoderOptions{TagName: "form"})
	// 使用解码器，将表单数据，解析到空实例中
	err := decoder.Decode(u.Ctx.Request().Form, user)
	if err != nil {
		golog.Error(err)
		u.Ctx.Redirect("/user/error")
		//本身不存在，但我们定义了访问一个不存在的结果即错误页面
		return
	}
	golog.Debug("根据用户注册表单，得到user实例：", user)
	_, err = u.UserService.AddUser(user)
	if err != nil {
		golog.Error(err)
		u.Ctx.Redirect("/user/error")
		return
	}
	u.Ctx.Redirect("/user/login")
}
func (u *UserController) GetLogin() mvc.View {
	return mvc.View{
		Name: "user/login.html",
	}
}
func (u *UserController) PostLogin() mvc.Response {
	// 由于数据量少，直接通过上下文获取表单中的数据
	userName := u.Ctx.FormValue("userName")
	password := u.Ctx.FormValue("password")
	_, isOK := u.UserService.IsPwdSussess(userName, password)
	if !isOK {
		golog.Error("登陆失败，用户名或密码不正确！")
		return mvc.Response{Path: "/user/login"}
	}
	user, err := u.UserService.GetUserByName(userName)
	if err != nil {
		golog.Error("获取用户信息失败！")
		return mvc.Response{Path: "/user/login"}
	}
	// 将用户id写入Cookie
	tool.GlobalCookie(u.Ctx, "uid", strconv.FormatInt(user.ID, 10), 60*60*24)
	golog.Debug("cookie设置成功！", u.Ctx.GetCookie("uid"))
	// 加uid加密
	uidString, err := encrypt.EnPwdCode([]byte(strconv.FormatInt(user.ID, 10)))
	if err != nil {
		golog.Error("加密失败！", err)
		return mvc.Response{Path: "/user/login"}
	}
	// 将用户加密信息写入Cookie
	tool.GlobalCookie(u.Ctx, "sign", uidString, 60*60*24)
	golog.Debug("登陆成功！")
	return mvc.Response{Path: "/product/detail"}
}

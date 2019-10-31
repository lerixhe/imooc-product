package controllers

import (
	"imooc-product/common"
	"imooc-product/datamodels"
	"imooc-product/services"
	"imooc-product/tool"

	"github.com/kataras/golog"

	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"

	"github.com/kataras/iris"
)

type UserController struct {
	Ctx         iris.Context
	UserService services.IUserService
	Session     *sessions.Session
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
		golog.Debug("登陆失败，用户名或密码不正确！")
		return mvc.Response{Path: "/user/login"}
	}
	user, err := u.UserService.GetUserByName(userName)
	if err != nil {
		golog.Debug("获取用户信息失败！")
		return mvc.Response{Path: "/user/login"}
	}
	u.Session.Set("sessionID"+userName+password+"md5", user.ID)
	golog.Debug("session设置成功！", u.Session.Get("sessionID"+userName+password+"md5"))
	tool.GlobalCookie(u.Ctx, "userlogin", "sessionID"+userName+password+"md5", 60*60*24)
	golog.Debug("cookie设置成功！", u.Ctx.GetCookie("userlogin"))
	golog.Debug("登陆成功！")
	return mvc.Response{Path: "/product/detail"}
}

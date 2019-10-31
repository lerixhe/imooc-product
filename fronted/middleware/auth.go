package middleware

import (
	"github.com/kataras/golog"
	"github.com/kataras/iris"
)

// 校验cookie是否存在
func AuthConProduct(ctx iris.Context) {
	sessionKey := ctx.GetCookie("userlogin")
	if sessionKey == "" {
		golog.Debug("用户未登录")
		ctx.Redirect("/user/login")
		return
	}
	golog.Debug("用户持有cookie")
	ctx.Next()
}

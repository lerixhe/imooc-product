package middleware

import (
	"imooc-product/encrypt"

	"github.com/kataras/golog"
	"github.com/kataras/iris"
)

// 校验cookie是否存在
func AuthConProduct(ctx iris.Context) {
	userID := ctx.GetCookie("uid")
	signKey := ctx.GetCookie("sign")
	if userID == "" || signKey == "" {
		golog.Debug("用户未登录")
		ctx.Redirect("/user/login")
		return
	}
	strByte, err := encrypt.DePwdCode(signKey)
	if err != nil {
		golog.Error("cookie解密失败", err)
		return
	}
	if userID != string(strByte) {
		golog.Debug("用户校验失败")
		return
	}
	golog.Debug("当前用户校验成功", userID)
	ctx.Next()
}

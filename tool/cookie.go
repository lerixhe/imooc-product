package tool

import (
	"net/http"

	"github.com/kataras/iris"
)

//设置全局cookie
func GlobalCookie(ctx iris.Context, name string, value string, timeout int) {
	ctx.SetCookie(
		&http.Cookie{
			Name:   name,
			Value:  value,
			Path:   "/",
			MaxAge: timeout,
		},
	)
}

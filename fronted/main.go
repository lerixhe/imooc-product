package main

import (
	"context"
	"imooc-product/common"
	"imooc-product/fronted/middleware"
	"imooc-product/fronted/web/controllers"
	"imooc-product/repositories"
	"imooc-product/services"
	"time"

	"github.com/kataras/iris/sessions"

	"github.com/kataras/golog"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

func main() {
	//1.创建iris 实例
	app := iris.New()
	//2.设置错误模式，在mvc模式下提示错误
	app.Logger().SetLevel("debug")
	//3.注册模板
	tmplate := iris.HTML("./web/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(tmplate)
	//4.设置静态资源目标
	app.StaticWeb("/public", "./web/public")
	//出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错！"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})
	//连接数据库
	db, err := common.NewMysqlConn()
	if err != nil {
		// log.Error(err)
		golog.Error(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 设置session
	sessions := sessions.New(
		sessions.Config{
			// 标记用户会话的纯session,能再未登录的情况下，确定一个用户
			// Cookie:  "helloworld",
			Expires: 60 * time.Hour,
		},
	)
	// sessions := &sessions.Sessions{}
	//5.注册控制器
	userRepository := repositories.NewUserManagerRepository("user", db)
	userSerivce := services.NewUserService(userRepository)
	userParty := app.Party("/user")
	user := mvc.New(userParty)
	user.Register(ctx, userSerivce, sessions.Start) //带上session
	user.Handle(new(controllers.UserController))

	// 取得所需服务实例
	productRespository := repositories.NewProductManager("product", db)
	productService := services.NewProductService(productRespository)
	orderRepository := repositories.NewOrderMangerRepository("user_order", db)
	orderService := services.NewOrderService(orderRepository)
	// 取得路由组实例，并使用中间件服务
	productParty := app.Party("/product")
	productParty.Use(middleware.AuthConProduct)
	// 根据路由组实例，获得mvc实例
	product := mvc.New(productParty)
	// 将服务实例注入这个mvc实例，并使用对应的contrller实例处理
	product.Register(ctx, productService, orderService, sessions.Start)
	product.Handle(new(controllers.ProductControllers))

	//6.启动服务
	app.Run(
		iris.Addr("localhost:8082"),
		// iris.WithoutVersionChecker,
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)

}

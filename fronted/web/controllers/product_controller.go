package controllers

import (
	"imooc-product/datamodels"
	"imooc-product/services"
	"strconv"

	"github.com/kataras/golog"
	"github.com/kataras/iris/mvc"

	"github.com/kataras/iris/sessions"

	"github.com/kataras/iris"
)

type ProductControllers struct {
	Ctx iris.Context
	// 商品服务
	ProductService services.IProductService
	// 用户购买商品，会创建订单，用到订单服务，注意提前注册
	OrderService services.IOrderService
	Session      *sessions.Session
}

func (p *ProductControllers) GetDetail() mvc.View {
	sessionKey := p.Ctx.GetCookie("userlogin")
	userID := p.Session.Get(sessionKey)
	if userID.(int64) == 0 {
		golog.Debug("用户校验失败")
		p.Ctx.Redirect("/user/login")
		return mvc.View{}
	}
	golog.Debug("当前用户", userID)
	product, err := p.ProductService.GetProductByID(1)
	if err != nil {
		golog.Error(err)
		p.Ctx.Redirect("/product/error")
		return mvc.View{
			Name: "shared/error.html",
		}
	}
	golog.Debug(product)
	return mvc.View{
		Layout: "shared/productLayout.html",
		Name:   "product/view.html",
		Data: iris.Map{
			"product":  product,
			"userName": userID,
		},
	}
}
func (p *ProductControllers) GetOrder() mvc.View {
	productIDstr := p.Ctx.URLParam("productID")
	productID, err := strconv.ParseInt(productIDstr, 10, 64)
	if err != nil {
		golog.Error(err)
		return mvc.View{Name: "shared/error.html",
			Data: iris.Map{
				"Message": "数据格式错误",
			},
		}
	}
	cookie := p.Ctx.GetCookie("userlogin")
	userIDx := p.Session.Get(cookie)
	userID := userIDx.(int64)
	product, err := p.ProductService.GetProductByID(productID)
	if err != nil {
		golog.Error(err)
		return mvc.View{Name: "shared/error.html",
			Data: iris.Map{
				"Message": "数据查询错误",
			},
		}

	}
	var orderID int64
	var showMessage string
	// 判断商品库存
	if product.ProductNum > 0 {
		// p.Ctx.BeginTransaction()
		// 减库存
		product.ProductNum--
		err := p.ProductService.UpdateProduct(product)
		if err != nil {
			golog.Error(err)
			product.ProductNum++
			return mvc.View{Name: "shared/error.html",
				Data: iris.Map{
					"Message": "库存更新失败",
				},
			}

		}
		golog.Debug("库存更新成功！订单还在创建中,商品ID:", productID)
		// 创订单
		order := &datamodels.Order{
			UserId:      userID,
			ProductId:   productID,
			OrderStatus: datamodels.OrderSuccess,
		}
		orderID, err = p.OrderService.InsertOrder(order)
		if err != nil {
			golog.Error(err)
			return mvc.View{Name: "shared/error.html",
				Data: iris.Map{
					"Message": "订单创建失败",
				},
			}

		}
		golog.Debug("订单创建成功！订单ID", orderID)
		showMessage = "抢购成功！请尽快付款"
	} else {
		showMessage = "库存不足"
	}
	return mvc.View{
		Layout: "shared/productLayout.html",
		Name:   "product/result.html",
		Data: iris.Map{
			"orderID":     orderID,
			"showMessage": showMessage,
		},
	}
}

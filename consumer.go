package main

import (
	"imooc-product/common"
	"imooc-product/rabbitmq"
	"imooc-product/repositories"
	"imooc-product/services"

	"github.com/kataras/golog"
)

func main() {
	golog.SetLevel("debug")
	db, err := common.NewMysqlConn()
	if err != nil {
		golog.Error("数据库连接失败", err)
		return
	}
	orderRepository := repositories.NewOrderMangerRepository("user_order", db)
	orderService := services.NewOrderService(orderRepository)
	productRepository := repositories.NewProductManager("product", db)
	productService := services.NewProductService(productRepository)

	// rabbit消费端
	rabbitmq := rabbitmq.NewSimpleRabbitMQ("imoocProduct")
	rabbitmq.ConsumeSimple(orderService, productService)
}

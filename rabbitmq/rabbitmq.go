package rabbitmq

import (
	"encoding/json"
	"imooc-product/datamodels"
	"imooc-product/services"

	"github.com/kataras/golog"
	"github.com/streadway/amqp"
)

// uri  amqp://账号：密码@地址：端口/vhost
const MQURL = "amqp://imoocuser:imoocuser@94.191.18.219:5672/imooc"

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	// 队列名称
	QueueName string
	// 交换机
	Exchange string
	// key
	Key string
	// 连接信息
	Mqurl string
}

// 创建结构体实例的函数
func NewRabbitMQ(queueName, exchange, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		Mqurl:     MQURL,
	}
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	if err != nil {
		golog.Error(err)
		return nil
	}
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	if err != nil {
		golog.Error(err)
		return nil
	}
	return rabbitmq
}

// 断开channel和connecion
func (r *RabbitMQ) Destory() {
	r.conn.Close()
	r.channel.Close()
}

// RabbitMQ的简单工作模式
func NewSimpleRabbitMQ(queueName string) *RabbitMQ {
	// 简单模式是：默认交换机，空key
	return NewRabbitMQ(queueName, "", "")
}

// 简单模式下的消息生产代码
func (r *RabbitMQ) PublishSimple(message string) error {
	// 申请队列,该操作是幂等的，无需担心重复申请带来的副作用
	_, err := r.channel.QueueDeclare(
		r.QueueName, // 队列名称name
		false,       // 是否持久化 durable
		false,       // 如果未使用，是否自动删除autoDelete
		false,       // 是否具有排他性，其他人不能访问队列exclusive
		false,       // 是否阻塞 noWait
		nil,         //额外属性

	)
	if err != nil {
		golog.Error(err)
		return err
	}
	err = r.channel.Publish(
		r.Exchange,
		r.QueueName, //routing key,一般是目标队列名
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		golog.Error(err)
		return err
	}
	return nil
}

// 简单模式下消费消息代码
func (r *RabbitMQ) ConsumeSimple(orderService services.IOrderService, productService services.IProductService) {
	// 申请队列,该操作是幂等的，无需担心重复申请带来的副作用
	_, err := r.channel.QueueDeclare(
		r.QueueName, // 队列名称name
		false,       // 是否持久化 durable
		false,       // 如果未使用，是否自动删除autoDelete
		false,       // 是否具有排他性，其他人不能访问队列exclusive
		false,       // 是否阻塞 noWait
		nil,         //额外属性
	)
	if err != nil {
		return
	}
	r.channel.Qos(
		1,     //消费者1次消费的最大数量
		0,     //服务器传递的最大容量
		false, //是否全局可用
	)
	msgs, err := r.channel.Consume(
		r.QueueName,
		"",
		false, //关闭自动应答，消费完一个再来第二个。
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return
	}

	forever := make(chan bool)
	// 多线程消费消息
	go func() {
		for d := range msgs {
			message := new(datamodels.Message)
			err := json.Unmarshal(d.Body, message)
			if err != nil {
				golog.Error(err)
				return
			}
			golog.Debugf("Receive a message:商品id:%d,用户id:%d", message.ProductID, message.UserId)
			orderID, err := orderService.InsertOrderByMessage(message)
			if err != nil {
				golog.Error(err)
			}
			err = productService.SubNumOne(message.ProductID)
			if err != nil {
				golog.Error(err)
			}
			golog.Debugf("订单创建完成,订单号%d，库存扣减成功", orderID)
			d.Ack(false)
		}
	}()
	golog.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

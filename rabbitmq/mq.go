package rabbitmq

import (
	"fmt"
	"path"
	"runtime"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/streadway/amqp"
	"os"
)

var (
	poolCount = 8 //连接池数量
	URL = "amqp://admin:admin@localhost:5672/"
)

var (
	logger log.Logger
	pool   = make(chan *RabbitMQ, poolCount)
)

func init() {
	logger = log.NewLogfmtLogger(os.Stdout)
	log.With(logger, "component", "RabbitMQ")
	newRabbitPool()
}

/**
Receiver 观察者模式需要的接口
观察者用于接收指定的queue到来的数据
*/
type Receiver interface {
	QueueName() string          // 获取接收者需要监听的队列
	RouterKey() string          // 这个队列绑定的路由
	OnError(err error)          // 处理遇到的错误，当RabbitMQ对象发生了错误，他需要告诉接收者处理错误
	OnReceive(body []byte) bool // 处理收到的消息, 这里需要告知RabbitMQ对象消息是否处理成功
}

// RabbitMQ 用于管理和维护rabbitmq的对象
type RabbitMQ struct {
	source       *amqp.Connection
	channel      *amqp.Channel
	exchangeName string     // exchange的名称
	exchangeType string     // exchange的类型
	exchangeArgs amqp.Table //exchange的额外参数
	receivers    []Receiver //消费者列表
}

//初始化连接池
func newRabbitPool() {
	logger.Log(
		"level", "info",
		"method", "newRabbitPool()",
		"msg", "正在初始化rabbitmq连接池。。。",
	)
	for i := 0; i < poolCount; i++ {
		rabbitmq := &RabbitMQ{}
		conn, err := amqp.Dial(URL)

		if err != nil {
			panic(err)
		} else {
			rabbitmq.source = conn
		}
		ch, err := conn.Channel()
		if err != nil {
			panic(err)
		} else {
			rabbitmq.channel = ch
		}
		pool <- rabbitmq
	}
}

func (mq *RabbitMQ) Publish(exname, routekey string, publishing amqp.Publishing) (success bool) {

	//mq.prepareExchange()
	//发送消息
	err := mq.channel.Publish(
		exname,   // exchange     交换器名称，使用默认
		routekey, // routing key    路由键，这里为队列名称
		false,    // mandatory
		false,
		publishing, //消息的详细信息
	)
	if err != nil {
		pc, file, line, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		logger.Log(
			"level", "error",
			"method", f.Name(),
			"file", path.Base(file),
			"line", line,
			"msg", err.Error(),
		)
		success = false
	} else {
		success = true
	}
	return
}

//根据情况合适的创建客户端
func NewRabbitMQ(exname, extype string, args amqp.Table) (*RabbitMQ, error) {
	var rbmq = &RabbitMQ{}
	//如果连接池中没有连接了，就重新申请连接返回
	if len(pool) == 0 {
		return CreateNewRabbitMQ(exname, extype, args)
	} else { //如果连接池中有链接，就将传来的参数保存到连接中返回
		rbmq = <-pool
		rbmq.exchangeArgs = args
		rbmq.exchangeType = extype
		rbmq.exchangeName = exname
		rbmq.receivers = make([]Receiver, 0)
	}
	return rbmq, nil
}

//直接重新创建一个客户端
func CreateNewRabbitMQ(exname, extype string, args amqp.Table) (*RabbitMQ, error) {
	var rbmq = &RabbitMQ{}
	logger.Log(
		"level", "info",
		"method", "NewRabbitMQ()",
		"msg", "连接数过多,正在创建新连接",
	)
	rbmq.exchangeName = exname
	rbmq.exchangeType = extype
	rbmq.exchangeArgs = args
	rbmq.receivers = make([]Receiver, 0)
	conn, err := amqp.Dial(URL)

	if err != nil {
		return nil, err
	} else {
		rbmq.source = conn
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	} else {
		rbmq.channel = ch
	}
	return rbmq, nil
}

// RegisterReceiver 注册一个用于接收指定队列指定路由的数据接收者
func (mq *RabbitMQ) RegisterReceiver(receiver Receiver) bool {
	for _, v := range mq.receivers {
		if v.QueueName() == receiver.QueueName() && v.RouterKey() == receiver.RouterKey() {
			return false
		}
	}
	mq.receivers = append(mq.receivers, receiver)
	return true
}

// prepareExchange 准备rabbitmq的Exchange
func (mq *RabbitMQ) prepareExchange() error {

	// 申明Exchange
	err := mq.channel.ExchangeDeclare(
		mq.exchangeName, // exchange
		mq.exchangeType, // type
		true,            // durable
		true,            // autoDelete
		false,           // internal
		false,           // noWait
		mq.exchangeArgs, // args
	)

	if nil != err {
		fmt.Println(err)
		return err
	}
	return nil
}

//重新连接channel
func (mq *RabbitMQ) Restart() error {
	conn, err := amqp.Dial(URL)
	if err != nil {
		return err
	}
	mq.source = conn
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	mq.channel = ch
	return nil
}

//销毁连接对象
func (mq *RabbitMQ) Close() {
	//如果连接池满了，说明该链接是后来创建的，直接关闭掉该链接就行
	if len(pool) == poolCount {
		logger.Log(
			"level", "info",
			"method", "Close()",
			"msg", "正在关闭连接。。。",
		)
		if mq.channel != nil {
			mq.channel.Close()
		}
		if mq.source != nil {
			mq.source.Close()
		}
	} else {
		//如果不是后来创建的，就将连接中的信息清除，放回连接池中
		mq.clear()
		pool <- mq
		logger.Log(
			"level", "info",
			"method", "Close()",
			"msg", "正在将连接放回连接池。。。",
		)
	}
}

//将连接中的业务信息删除
func (mq *RabbitMQ) clear() {
	mq.exchangeName = ""
	mq.exchangeType = ""
	mq.receivers = []Receiver{}
	mq.exchangeArgs = amqp.Table{}
}

// Listen 监听指定路由发来的消息
// 这里需要针对每一个接收者启动一个goroutine来执行listen
// 该方法负责从每一个接收者监听的队列中获取数据，并负责重试
func (mq *RabbitMQ) Listen(receiver Receiver) {

	mq.prepareExchange()
	// 这里获取每个接收者需要监听的队列和路由
	queueName := receiver.QueueName()
	routerKey := receiver.RouterKey()
	_, err := mq.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when usused
		false,     // exclusive(排他性队列)
		false,     // no-wait
		nil,       // arguments
	)
	if nil != err {
		err = mq.Restart()
		if err != nil {
			// 当队列初始化失败的时候，需要告诉这个接收者相应的错误
			receiver.OnError(fmt.Errorf("初始化队列 %s 失败: %s", queueName, err.Error()))
			return
		}
	}
	// 将Queue绑定到 Exchange上去
	err = mq.channel.QueueBind(
		queueName,       // queue name
		routerKey,       // routing key
		mq.exchangeName, // exchange
		false,           // no-wait
		nil,
	)
	if nil != err {
		receiver.OnError(fmt.Errorf("绑定队列 [%s - %s] 到Exchanges失败: %s", queueName, routerKey, err.Error()))
		return
	}
	// 获取消费通道
	//prefetchSize：0 prefetchSize maximum amount of content (measured in* octets) that the server will deliver, 0 if unlimited
	//prefetchCount：会告诉RabbitMQ不要同时给一个消费者推送多于N个消息，即一旦有N个消息还没有ack，则该consumer将block掉，直到有消息ack
	//global：true\false 是否将上面设置应用于channel，简单点说，就是上面限制是channel级别的还是consumer级别
	//备注：据说prefetchSize 和global这两项，rabbitmq没有实现，暂且不研究
	mq.channel.Qos(10, 0, false) // 确保rabbitmq会一个一个发消息
	msgs, err := mq.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if nil != err {
		receiver.OnError(fmt.Errorf("获取队列 %s 的消费通道失败: %s", queueName, err.Error()))
		return
	}

	// 使用callback消费数据
	for msg := range msgs {
		// 当接收者消息处理失败的时候，
		// 比如网络问题导致的数据库连接失败，连接失败等等这种
		// 通过重试可以成功的操作，那么这个时候是需要重试的
		// 直到数据处理成功后再返回，然后才会回复rabbitmq ack
		for !receiver.OnReceive(msg.Body) {
			pc, file, line, _ := runtime.Caller(1)
			f := runtime.FuncForPC(pc)
			logger.Log(
				"level", "error",
				"method", f.Name(),
				"file", path.Base(file),
				"line", line,
				"msg", "receiver 数据处理失败，将要重试",
			)
			time.Sleep(1 * time.Second)
		}
		// 确认收到本条消息, multiple必须为false
		msg.Ack(false)
	}
}

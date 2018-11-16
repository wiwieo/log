package rabbitmq

import (
	"testing"

	"fmt"
	"github.com/streadway/amqp"
	"time"
)

func TestNewRabbitMQ(t *testing.T) {
	for i := 0; i < 10; i++ {
		go tcreate()
	}
	time.Sleep(4 * time.Second)

}

func tcreate() {
	a, _ := NewRabbitMQ("message_delay", "x-delayed-message", amqp.Table(map[string]interface{}{"x-delayed-type": "direct"}))
	time.Sleep(2 * time.Second)
	defer a.Close()
}

func TestListen(t *testing.T) {
	a, _ := NewRabbitMQ("message_delay", "x-delayed-message", amqp.Table(map[string]interface{}{"x-delayed-type": "direct"}))
	err := a.channel.QueueBind(
		"message_delay",
		"message_delay",
		a.exchangeName,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("1........")
	msgs, err := a.channel.Consume(
		"message_delay", // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	fmt.Println("2........")
	for msga := range msgs {
		// 当接收者消息处理失败的时候，
		// 比如网络问题导致的数据库连接失败，连接失败等等这种
		// 通过重试可以成功的操作，那么这个时候是需要重试的
		// 直到数据处理成功后再返回，然后才会回复rabbitmq ack
		fmt.Println("3........")
		msg := string(msga.Body[:])
		fmt.Println(msg)
		// 确认收到本条消息, multiple必须为false
		msga.Ack(false)
	}
	fmt.Println("4........")
	//c := NewConsumer("message_delay","")
	//a.RegisterReceiver(c)
	//a.Start()
}

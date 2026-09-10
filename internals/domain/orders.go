package domain

type OrderQueueSetup struct {
	ExchangeName string
	TypeExchange string
	RouteKeyName string
	QueueName    string
}

func NewOrderQueueSetup(exchange, typeExchg, routekey, queue string) OrderQueueSetup {
	return OrderQueueSetup{
		ExchangeName: exchange,
		TypeExchange: typeExchg,
		RouteKeyName: routekey,
		QueueName:    queue,
	}
}

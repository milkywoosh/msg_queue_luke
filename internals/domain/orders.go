package domain

type QueueSetup struct {
	ExchangeName string
	TypeExchange string
	RouteKeyName string
	QueueName    string
}

func NewQueueSetup(exchange, typeExchg, routekey, queue string) QueueSetup {
	return QueueSetup{
		ExchangeName: exchange,
		TypeExchange: typeExchg,
		RouteKeyName: routekey,
		QueueName:    queue,
	}
}

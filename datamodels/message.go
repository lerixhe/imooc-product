package datamodels

type Message struct {
	ProductID int64
	UserId    int64
}

func NewMessage(productID, userId int64) *Message {
	return &Message{productID, userId}
}

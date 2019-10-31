package datamodels

type Order struct {
	ID          int64 `sql:"ID"`
	UserId      int64 `sql:"userID"`
	ProductId   int64 `sql:"productID"`
	OrderStatus int64 `sql:"orderStatus"`
}

const (
	// 进行中
	OrderWait    = iota
	OrderSuccess //成功
	OrderFailed  //失败
)

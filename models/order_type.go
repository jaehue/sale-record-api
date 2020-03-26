package models

type OrderType struct {
	CanMakeSaleRecord bool
	Sequence          int
	Value             string
	Type              string
}

var orderTypes = []OrderType{
	OrderType{Sequence: 1, CanMakeSaleRecord: false, Value: "SaleOrderProcessing", Type: "ORDER"},
	OrderType{Sequence: 2, CanMakeSaleRecord: false, Value: "SaleOrderCancel", Type: "ORDER"},
	OrderType{Sequence: 3, CanMakeSaleRecord: true, Value: "SaleOrderFinished", Type: "ORDER"},
	OrderType{Sequence: 4, CanMakeSaleRecord: true, Value: "StockDistributed", Type: "ORDER"},
	OrderType{Sequence: 5, CanMakeSaleRecord: true, Value: "SaleShippingWaiting", Type: "ORDER"},
	OrderType{Sequence: 6, CanMakeSaleRecord: true, Value: "SaleShippingProcessing", Type: "ORDER"},
	OrderType{Sequence: 7, CanMakeSaleRecord: true, Value: "SaleShippingFinished", Type: "ORDER"},
	OrderType{Sequence: 8, CanMakeSaleRecord: true, Value: "BuyerReceivedConfirmed", Type: "ORDER"},
	OrderType{Sequence: 9, CanMakeSaleRecord: true, Value: "SaleOrderSuccess", Type: "ORDER"},
	OrderType{Sequence: 10, CanMakeSaleRecord: false, Value: "RefundOrderRegistered", Type: "REFUND"},
	OrderType{Sequence: 11, CanMakeSaleRecord: false, Value: "SellerRefundAgree", Type: "REFUND"},
	OrderType{Sequence: 12, CanMakeSaleRecord: false, Value: "RefundOrderCancel", Type: "REFUND"},
	OrderType{Sequence: 13, CanMakeSaleRecord: false, Value: "RefundOrderProcessing", Type: "REFUND"},
	OrderType{Sequence: 14, CanMakeSaleRecord: false, Value: "RefundShippingWaiting", Type: "REFUND"},
	OrderType{Sequence: 15, CanMakeSaleRecord: false, Value: "RefundShippingProcessing", Type: "REFUND"},
	OrderType{Sequence: 16, CanMakeSaleRecord: false, Value: "RefundShippingFinished", Type: "REFUND"},
	OrderType{Sequence: 17, CanMakeSaleRecord: false, Value: "RefundRequisiteApprovals", Type: "REFUND"},
	OrderType{Sequence: 18, CanMakeSaleRecord: true, Value: "RefundOrderSuccess", Type: "REFUND"},
	OrderType{Sequence: 18, CanMakeSaleRecord: false, Value: "Undefined", Type: ""},
}

func FindOrderType(oType string) OrderType {
	for _, orderType := range orderTypes {
		if orderType.Value == oType {
			return orderType
		}
	}
	return OrderType{Sequence: 18, CanMakeSaleRecord: false, Value: "Undefined", Type: ""}
}

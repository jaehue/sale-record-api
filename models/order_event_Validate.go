package models

import (
	"context"
	"errors"

	"github.com/pangpanglabs/goutils/number"
)

type OrderEventValidate struct{}

func (orderEvent *OrderEvent) OrderValidate(ctx context.Context) error {
	if orderEvent.CreatedId == 0 {
		return errors.New(string(CreatedIdError))
	}
	if orderEvent.SaleType == "POS" && orderEvent.SalesmanId == 0 {
		return errors.New(string(POSSalesmanIdError))
	}
	if orderEvent.StoreId == 0 {
		return errors.New(string(StoreIdError))
	}

	if err := orderEvent.OrderOfferValidate(); err != nil {
		return err
	}

	if err := orderEvent.OrderItemValidate(); err != nil {
		return err
	}

	return nil
}

func (orderEvent *OrderEvent) OrderItemValidate() error {
	mileagePrice := orderEvent.MileagePrice
	mileage := orderEvent.Mileage
	for _, orderItemEvent := range orderEvent.Items {
		if orderItemEvent.FeeRate == 0 {
			return errors.New(string(ItemFeeRateError))
		}
		mileagePrice = number.ToFixed(mileagePrice-orderItemEvent.MileagePrice, nil)
		mileage = number.ToFixed(mileage-orderItemEvent.Mileage, nil)
	}
	if mileagePrice != float64(0) || mileage != float64(0) {
		return errors.New(string(MileageError))
	}

	return nil
}

func (orderEvent *OrderEvent) OrderOfferValidate() error {

	// for _, offer := range orderEvent.Offers {

	// }
	return nil
}

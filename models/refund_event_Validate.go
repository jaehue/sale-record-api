package models

import (
	"context"
	"errors"

	"github.com/pangpanglabs/goutils/number"
)

type RefundEventValidate struct{}

func (refundEvent *RefundEvent) RefundValidate(ctx context.Context) error {
	if refundEvent.CreatedId == 0 {
		return errors.New(string(CreatedIdError))
	}
	if refundEvent.RefundType == "POS" && refundEvent.SalesmanId == 0 {
		return errors.New(string(POSSalesmanIdError))
	}
	if refundEvent.StoreId == 0 {
		return errors.New(string(StoreIdError))
	}

	if err := refundEvent.RefundOfferValidate(); err != nil {
		return err
	}

	if err := refundEvent.RefundItemValidate(); err != nil {
		return err
	}

	return nil
}

func (refundEvent *RefundEvent) RefundItemValidate() error {
	mileagePrice := refundEvent.MileagePrice
	mileage := refundEvent.Mileage
	for _, refundItemEvent := range refundEvent.Items {
		if refundItemEvent.FeeRate == 0 {
			return errors.New(string(ItemFeeRateError))
		}
		mileagePrice = number.ToFixed(mileagePrice-refundItemEvent.MileagePrice, nil)
		mileage = number.ToFixed(mileage-refundItemEvent.Mileage, nil)
	}
	if mileagePrice != float64(0) || mileage != float64(0) {
		return errors.New(string(MileageError))
	}

	return nil
}

func (refundEvent *RefundEvent) RefundOfferValidate() error {

	// for _, offer := range refundEvent.Offers {

	// }
	return nil
}

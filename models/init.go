package models

import "github.com/go-xorm/xorm"

func Init(db *xorm.Engine) error {
	return db.Sync2(new(AssortedSaleRecord),
		new(AssortedSaleRecordDtl),
		new(AppliedSaleRecordCartOffer),
		new(AppliedSaleRecordItemOffer),
		new(SaleRecordSuccess),
		new(AssortedSaleRecordPayment),
		new(SaleRecordItemAppliedCartOffer),
	)
}

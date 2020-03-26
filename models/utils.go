package models

import "github.com/hublabs/common/auth"

// TODO: github.com/hublabs/common/auth에 아래 로직 반영 & tenantCode 처리
type UserClaim auth.UserClaim

func (u UserClaim) isCustomer() bool {
	// return u.Iss == auth.IssMembership
	return u.Issuer == "membership"
}
func (u UserClaim) customerId() int64 {
	return 0
}
func (u UserClaim) tenantCode() string {
	return "hublabs"
}
func (u UserClaim) channelId() int64 {
	return 0
}

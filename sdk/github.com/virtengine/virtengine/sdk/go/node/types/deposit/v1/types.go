package v1

type HasDeposit interface {
	GetDeposit() Deposit
}

type HasDeposits interface {
	GetDeposits() Deposits
}

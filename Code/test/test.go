package main

import (
	"fmt"
	"sync/atomic"
)

type Account struct {
	Owner   string
	balance float64
}

func (a *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("invalid amount")
	}
	a.balance += amount
	return nil
}

func (a *Account) Withdraw(amount float64) error {
	if amount > a.balance {
		return fmt.Errorf("insufficient funds")
	}
	a.balance -= amount
	return nil
}

func (a Account) Balance() float64 { return a.balance }

func main() {
	var requests atomic.Int64
	requests.Add(1)
	requests.Add(6)
	requests.Add(9)
	requests.Add(10)
	requests.Add(1)
	fmt.Println(requests.Load())
}

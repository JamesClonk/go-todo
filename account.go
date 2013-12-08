package main

import "sort"
import "strings"

type Account struct {
	Id       int    `db:"ID"`
	Name     string `db:"NAME"`
	Email    string `db:"EMAIL"`
	Password string `db:"PASSWORD"`
	Salt     string `db:"SALT"`
	Role     string `db:"ROLE"`
	LastAuth int    `db:"LAST_AUTH"`
}

type Accounts []Account

type accountSort struct {
	accounts Accounts
	by       func(a1, a2 *Account) bool
}

func (a *accountSort) Len() int {
	return len(a.accounts)
}

func (a *accountSort) Swap(l, r int) {
	a.accounts[l], a.accounts[r] = a.accounts[r], a.accounts[l]
}

func (a *accountSort) Less(l, r int) bool {
	return a.by(&a.accounts[l], &a.accounts[r])
}

func (a Accounts) sortBy(by func(a1, a2 *Account) bool) *Accounts {
	ts := &accountSort{
		accounts: a,
		by:       by,
	}
	sort.Sort(ts)
	return &a
}

func (a *Accounts) SortByName(order string) *Accounts {
	a.sortBy(func(a1, a2 *Account) bool {
		if order == "DESC" {
			return strings.ToLower(a1.Name) > strings.ToLower(a2.Name)
		}
		return strings.ToLower(a1.Name) < strings.ToLower(a2.Name)
	})
	return a
}

func (a *Accounts) SortByEmail(order string) *Accounts {
	a.sortBy(func(a1, a2 *Account) bool {
		if order == "DESC" {
			return strings.ToLower(a1.Email) > strings.ToLower(a2.Email)
		}
		return strings.ToLower(a1.Email) < strings.ToLower(a2.Email)
	})
	return a
}

package main

import "testing"
import "strings"

func Test_account_SortBy(t *testing.T) {
	var as1 = Accounts{
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
	}

	var as2 = Accounts{
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
	}

	// sort by domain name in lowercase
	as1.sortBy(func(a1, a2 *Account) bool {
		i1 := strings.Index(a1.Email, "@")
		i2 := strings.Index(a2.Email, "@")
		return strings.ToLower(a1.Email[i1:]) < strings.ToLower(a2.Email[i2:])
	})

	for i, a1 := range as1 {
		if a1 != as2[i] {
			t.Errorf("SortBy is not as expected: [%v], instead of [%v]", as1, as2)
			return
		}
	}
}

func Test_account_SortByName(t *testing.T) {
	var as1 = Accounts{
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
	}

	var as2 = Accounts{
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
	}

	var as3 = Accounts{
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
	}

	as1.SortByName("ASC")
	for i, a1 := range as1 {
		if a1 != as2[i] {
			t.Errorf("SortByName ASC is not as expected: [%v], instead of [%v]", as1, as2)
			return
		}
	}

	as1.SortByName("DESC")
	for i, a1 := range as1 {
		if a1 != as3[i] {
			t.Errorf("SortByName DESC is not as expected: [%v], instead of [%v]", as1, as3)
			return
		}
	}
}

func Test_account_SortByEmail(t *testing.T) {
	var as1 = Accounts{
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
	}

	var as2 = Accounts{
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
	}

	var as3 = Accounts{
		Account{3, "ozzie", "ozzie@abrakadabra", "abcd", "789", "User", 1234567892},
		Account{1, "JamesClonk", "JamesClonk@developer", "abcd", "123", "Admin", 1234567890},
		Account{2, "Clude", "clude@CLUDE", "abcd", "456", "User", 1234567891},
	}

	as1.SortByEmail("ASC")
	for i, a1 := range as1 {
		if a1 != as2[i] {
			t.Errorf("SortByEmail ASC is not as expected: [%v], instead of [%v]", as1, as2)
			return
		}
	}

	as1.SortByEmail("DESC")
	for i, a1 := range as1 {
		if a1 != as3[i] {
			t.Errorf("SortByEmail DESC is not as expected: [%v], instead of [%v]", as1, as3)
			return
		}
	}
}

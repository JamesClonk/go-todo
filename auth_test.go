package main

import "fmt"
import "time"
import "strconv"
import "testing"
import "net/http"

func _authSetup(t *testing.T, accountId int) (*Account, *string, *string, *string, *string, error) {
	account, err := GetAccountById(accountId)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	id := strconv.Itoa(account.Id)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	salt, err := GenerateRandomString()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	token := GenerateToken(timestamp, *salt, account.Password)

	return account, &id, &timestamp, salt, &token, nil
}

func Test_auth_setup(t *testing.T) {
	isLogging = true
	_storage_setup(t)
}

func Test_auth_GenerateToken(t *testing.T) {
	token := GenerateToken("1234567890", "abcdefghijklmnopqrstuvwxyz", "ZYXWVUTSRQPONMLKJIHGFEDCBA")
	expectedToken := "d3ccc18e527b739efa69810994cccf73235003932e657d91d1ac418027cdb9c379377e5afb30d3ae8b7b3c56c002464f3d23151b84543b18ef5ba5449208aa2b"
	if token != expectedToken {
		t.Errorf("Token is not as expected: [%v], instead of [%v]", token, expectedToken)
	}
}

func Test_auth_HashPassword(t *testing.T) {
	passwordHash := HashPassword("ZYXWVUTSRQPONMLKJIHGFEDCBA", "test123")
	expectedHash := "9fcf3fe763849dc5c8dfff8cf0620a29973ab8bcec80ffa1e884b3bf2e9f83e4fb704100476d3a9f29473f726f309e0ff8a8a1225c2eb47087b358f01c5ff53f"
	if passwordHash != expectedHash {
		t.Errorf("Password hash is not as expected: [%v], instead of [%v]", passwordHash, expectedHash)
	}
}

func Test_auth_GenerateRandomString(t *testing.T) {
	string1, err := GenerateRandomString()
	if err != nil {
		t.Error(err)
		return
	}

	string2, err := GenerateRandomString()
	if err != nil {
		t.Error(err)
		return
	}

	if *string1 == *string2 {
		t.Errorf("Random strings are identical: [%v] = [%v]", string1, string2)
	}
}

func Test_auth_Authenticate(t *testing.T) {
	account, id, timestamp, salt, token, err := _authSetup(t, 1)
	if err != nil {
		t.Error(err)
		return
	}

	// ============================================ Valid ============================================
	request, err := http.NewRequest("GET", "http://localhost:8008/does not matter here/?rId="+*id+"&rTimestamp="+*timestamp+"&rSalt="+*salt+"&rToken="+*token, nil)
	if err != nil {
		t.Error(err)
		return
	}

	authId, err := Authenticate(request)
	if err != nil {
		t.Error(err)
		return
	}
	if *authId != account.Id {
		t.Errorf("AccountId not authenticated: [%v] should be [%v]", authId, account.Id)
	}

	// ============================================ Expired ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/auth/?rId="+*id+"&rTimestamp="+*timestamp+"1&rSalt="+*salt+"&rToken="+*token, nil)
	if err != nil {
		t.Error(err)
		return
	}

	authId, err = Authenticate(request)
	if err != nil {
		t.Error(err)
		return
	}
	if authId != nil {
		t.Errorf("AccountId should not be authenticated: [%v] should be [nil]", authId)
	}

	// ============================================ Nonexisting Id ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/auth/?rId=5&rTimestamp="+*timestamp+"&rSalt="+*salt+"&rToken="+*token, nil)
	if err != nil {
		t.Error(err)
		return
	}

	authId, err = Authenticate(request)
	if err == nil {
		t.Error("Expected error!")
	}
	if authId != nil {
		t.Errorf("AccountId should not be authenticated: [%v] should be [nil]", authId)
	}

	// ============================================ Invalid Id ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/auth/?rId=abc&rTimestamp="+*timestamp+"&rSalt="+*salt+"&rToken="+*token, nil)
	if err != nil {
		t.Error(err)
		return
	}

	authId, err = Authenticate(request)
	if err == nil {
		t.Error("Expected parsing error!")
	}
	if authId != nil {
		t.Errorf("AccountId should not be authenticated: [%v] should be [nil]", authId)
	}

	// ============================================ Invalid Timestamp ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/auth/?rId="+*id+"&rTimestamp=1abc23&rSalt="+*salt+"&rToken="+*token, nil)
	if err != nil {
		t.Error(err)
		return
	}

	authId, err = Authenticate(request)
	if err == nil {
		t.Error("Expected parsing error!")
	}
	if authId != nil {
		t.Errorf("AccountId should not be authenticated: [%v] should be [nil]", authId)
	}

	// ============================================ Invalid Token ============================================
	request, err = http.NewRequest("GET", "http://localhost:8008/auth/?rId="+*id+"&rTimestamp="+*timestamp+"&rSalt="+*salt+"&rToken=1"+*token, nil)
	if err != nil {
		t.Error(err)
		return
	}

	authId, err = Authenticate(request)
	if err != nil {
		t.Error(err)
		return
	}
	if authId != nil {
		t.Errorf("AccountId should not be authenticated: [%v] should be [nil]", authId)
	}
}

func Test_auth_cleanup(t *testing.T) {
	_storage_cleanup()
}

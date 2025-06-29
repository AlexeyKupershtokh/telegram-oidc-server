package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type UserData struct {
	ID        int64
	FirstName string
	LastName  string
	Username  string
	PhotoURL  string
	AuthDate  int64
	Hash      string
}

type Verifier interface {
	ParseData(query url.Values) (UserData, error)
	VerifyHash(data UserData, botToken string) bool
}

type DefaultVerifier struct{}

func (v *DefaultVerifier) ParseData(query url.Values) (UserData, error) {
	id, err := strconv.ParseInt(query.Get("id"), 10, 64)
	if err != nil {
		return UserData{}, err
	}

	authDate, err := strconv.ParseInt(query.Get("auth_date"), 10, 64)
	if err != nil {
		return UserData{}, err
	}

	return UserData{
		ID:        id,
		FirstName: query.Get("first_name"),
		LastName:  query.Get("last_name"),
		Username:  query.Get("username"),
		PhotoURL:  query.Get("photo_url"),
		AuthDate:  authDate,
		Hash:      query.Get("hash"),
	}, nil
}

func (v *DefaultVerifier) VerifyHash(data UserData, botToken string) bool {
	secret := sha256.Sum256([]byte(botToken))
	kv := map[string]string{
		"id":         fmt.Sprintf("%d", data.ID),
		"first_name": data.FirstName,
		"auth_date":  fmt.Sprintf("%d", data.AuthDate),
	}

	if data.LastName != "" {
		kv["last_name"] = data.LastName
	}
	if data.Username != "" {
		kv["username"] = data.Username
	}
	if data.PhotoURL != "" {
		kv["photo_url"] = data.PhotoURL
	}

	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var checkStrings []string
	for _, k := range keys {
		checkStrings = append(checkStrings, fmt.Sprintf("%s=%s", k, kv[k]))
	}
	checkString := strings.Join(checkStrings, "\n")

	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(checkString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return expectedHash == data.Hash
}

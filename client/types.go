package client

import (
	"net/http"
)

type MastodonClientOptions struct {
	Server      string
	AccessToken string
}

type MastodonClientImpl struct {
	Opts           *MastodonClientOptions
	HttpClient     *http.Client
	UserAgent      string
	currentAccount Account
}

type Account struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Acct     string `json:"acct"`
}

type Status struct {
	Id                 string  `json:"id"`
	Url                string  `json:"url"`
	Account            Account `json:"account"`
	CreatedAt          string  `json:"created_at"`
	Content            string  `json:"content"`
	Visibility         string  `json:"visibility"`
	RebloggedByMe      bool    `json:"reblogged"`
	Reblog             *Status `json:"reblog"`
	InReplyToAccountId string  `json:"in_reply_to_account_id"`
}

type Search struct {
	Accounts []Account `json:"accounts"`
	Statuses []Status  `json:"statuses"`
}

type ExcludeFunc func(status Status) bool

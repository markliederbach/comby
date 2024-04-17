package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"
)

const (
	EndpointSearch            = "/api/v2/search"
	EndpointVerifyCredentials = "/api/v1/accounts/verify_credentials"
	EndpointAccountStatuses   = "/api/v1/accounts/%s/statuses"
	EndpointBoostStatus       = "/api/v1/statuses/%s/reblog"
)

// https://docs.joinmastodon.org/methods/statuses/#boost

func NewMastodonClient(ctx context.Context, options *MastodonClientOptions) (*MastodonClientImpl, error) {
	client := &MastodonClientImpl{
		Opts:       options,
		HttpClient: http.DefaultClient,
		UserAgent:  "markliederbach/comby",
	}

	_, err := client.GetCurrentAccount(ctx)
	if err != nil {
		return &MastodonClientImpl{}, err
	}
	return client, nil
}

func (c *MastodonClientImpl) GetCurrentAccount(ctx context.Context) (Account, error) {
	if c.currentAccount.Id != "" {
		return c.currentAccount, nil
	}
	resp, err := c.call(ctx, http.MethodGet, EndpointVerifyCredentials, url.Values{})
	if err != nil {
		return Account{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Account{}, errors.New(resp.Status)
	}

	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return Account{}, err
	}

	c.currentAccount = account
	return account, nil
}

func (c *MastodonClientImpl) GetAccount(ctx context.Context, username string) (Account, error) {
	params := url.Values{}
	params.Set("q", username)
	params.Set("limit", "1")
	resp, err := c.call(ctx, http.MethodGet, EndpointSearch, params)
	if err != nil {
		return Account{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Account{}, errors.New(resp.Status)
	}

	var results Search
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return Account{}, err
	}

	if len(results.Accounts) == 0 {
		return Account{}, fmt.Errorf("no accounts found for %s", username)
	}
	return results.Accounts[0], nil
}

func (c *MastodonClientImpl) BoostStatus(ctx context.Context, status Status) (Status, error) {
	endpoint := fmt.Sprintf(EndpointBoostStatus, status.Id)
	resp, err := c.call(ctx, http.MethodPost, endpoint, url.Values{})
	if err != nil {
		return Status{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Status{}, errors.New(resp.Status)
	}

	boostedStatus := Status{}
	if err := json.NewDecoder(resp.Body).Decode(&boostedStatus); err != nil {
		return Status{}, err
	}
	return boostedStatus, nil
}

func (c *MastodonClientImpl) GetAccountStatuses(ctx context.Context, account Account, filters ...ExcludeFunc) ([]Status, error) {
	endpoint := fmt.Sprintf(EndpointAccountStatuses, account.Id)
	params := url.Values{}
	params.Set("limit", "20")
	// params.Set("exclude_reblogs", "true")
	// params.Set("exclude_replies", "true")
	resp, err := c.call(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return []Status{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return []Status{}, errors.New(resp.Status)
	}
	statuses := []Status{}
	if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
		return []Status{}, err
	}

	for _, filter := range filters {
		filteredStatuses := []Status{}
		for _, status := range statuses {
			if filter(status) {
				filteredStatuses = append(filteredStatuses, status)
			}
		}
		statuses = filteredStatuses

	}
	return statuses, nil
}

func (c *MastodonClientImpl) call(ctx context.Context, method, uri string, params url.Values) (*http.Response, error) {
	parsedUrl, err := url.Parse(c.Opts.Server)
	if err != nil {
		return &http.Response{}, err
	}

	parsedUrl.Path = path.Join(parsedUrl.Path, uri)
	parsedUrl.RawQuery = params.Encode()

	var req *http.Request
	switch method {
	case http.MethodGet:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, parsedUrl.String(), nil)
		if err != nil {
			return &http.Response{}, err
		}
	case http.MethodPost:
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, parsedUrl.String(), nil)
		if err != nil {
			return &http.Response{}, err
		}
	default:
		return &http.Response{}, fmt.Errorf("unsupported method: %s", method)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Opts.AccessToken))
	req.Header.Set("User-Agent", c.UserAgent)

	var resp *http.Response
	backoff := time.Second
	for {
		resp, err = c.HttpClient.Do(req)
		if err != nil {
			return &http.Response{}, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if backoff > 5*time.Minute {
				return resp, errors.New(resp.Status)
			}

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return resp, ctx.Err()
			}

			backoff = time.Duration(float64(backoff) * 1.5)
			continue
		}

		break
	}
	return resp, nil
}

func (c *MastodonClientImpl) ExcludeStatusOlderThan(d time.Duration) ExcludeFunc {
	return func(status Status) bool {
		t, err := time.Parse(time.RFC3339, status.CreatedAt)
		if err != nil {
			return false
		}
		return time.Since(t) < d
	}
}

func (c *MastodonClientImpl) ExcludeRepliesToOwnStatuses() ExcludeFunc {
	return func(status Status) bool {
		return status.InReplyToAccountId != status.Account.Id
	}
}

func (c *MastodonClientImpl) ExcludeAllReplies() ExcludeFunc {
	return func(status Status) bool {
		return status.InReplyToAccountId == ""
	}
}

func (c *MastodonClientImpl) ExcludeBoosts() ExcludeFunc {
	return func(status Status) bool {
		return status.Reblog == nil
	}
}

func (c *MastodonClientImpl) ExcludeAlreadyBoosted() ExcludeFunc {
	return func(status Status) bool {
		return !status.RebloggedByMe
	}
}

func (c *MastodonClientImpl) ExcludeMatchingRegex(patterns ...string) ExcludeFunc {
	return func(status Status) bool {
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, status.Content); matched {
				return false
			}
		}
		return true
	}
}

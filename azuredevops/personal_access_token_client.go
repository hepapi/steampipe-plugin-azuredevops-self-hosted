package azuredevops

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/delegatedauthorization"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/tokenadmin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

const (
	patPageSize   = 100
	patMaxPages   = 100
	patAPIVersion = "6.0-preview.1"
)

// personalAccessTokenPage is one page of PAT metadata. The secret fields from
// the API are cleared before a row is built.
type personalAccessTokenPage struct {
	ContinuationToken string
	Value             []delegatedauthorization.SessionToken
}

type rawPersonalAccessTokenPage struct {
	ContinuationToken json.RawMessage                       `json:"continuationToken"`
	Value             []delegatedauthorization.SessionToken `json:"value"`
}

// personalAccessTokenLister lists PATs for one connection. A server-root
// connection whose token-admin route is not registered retries each collection
// until one accepts the call, then keeps that collection.
type personalAccessTokenLister struct {
	connection      *azuredevops.Connection
	settled         bool
	triedCollection bool
}

func (l *personalAccessTokenLister) list(ctx context.Context, d *plugin.QueryData, subject, continuation string) (*personalAccessTokenPage, error) {
	page, err := fetchPersonalAccessTokenPage(ctx, l.connection, subject, continuation)
	if err == nil {
		l.settled = true
		return page, nil
	}
	if l.settled || l.triedCollection || !patRouteMissing(d, err) {
		return nil, err
	}

	l.triedCollection = true
	original := l.connection.BaseUrl
	collections, ferr := tokenAdminCollectionURLs(ctx, d)
	if ferr != nil {
		return nil, ferr
	}
	lastErr := err
	for _, next := range collections {
		if sameBaseURL(next, original) {
			continue
		}
		shifted, cerr := connectionForURL(d, next)
		if cerr != nil {
			return nil, cerr
		}
		l.connection = shifted
		page, lastErr = fetchPersonalAccessTokenPage(ctx, shifted, subject, continuation)
		if lastErr == nil {
			l.settled = true
			return page, nil
		}
		// A registered route that answers 404 is this user, not a missing API.
		// Keep the collection and let the caller skip the user.
		if !patLocationMissing(lastErr) || !patRouteAbsent(lastErr) {
			return nil, lastErr
		}
	}
	return nil, lastErr
}

func fetchPersonalAccessTokenPage(ctx context.Context, connection *azuredevops.Connection, subject, continuation string) (*personalAccessTokenPage, error) {
	callCtx, cancel := patCallContext(ctx)
	defer cancel()
	client, err := tokenadmin.NewClient(callCtx, connection)
	if err != nil {
		if apiNotRegistered(err) {
			return fetchUnregisteredPersonalAccessTokens(callCtx, connection, subject, continuation)
		}
		return nil, err
	}
	page, err := fetchPersonalAccessTokensByClient(callCtx, client, subject, continuation)
	if err != nil && apiNotRegistered(err) {
		return fetchUnregisteredPersonalAccessTokens(callCtx, connection, subject, continuation)
	}
	return page, err
}

// patLocationError marks a failure from the raw route used when the v7
// token-admin location id is not registered. A 404 in that case can mean the
// route itself is missing. A 404 from the registered client means the user is
// unknown.
type patLocationError struct {
	err error
}

func (e *patLocationError) Error() string {
	if e == nil || e.err == nil {
		return "personal access token route is not registered"
	}
	return e.err.Error()
}

func (e *patLocationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func fetchUnregisteredPersonalAccessTokens(ctx context.Context, connection *azuredevops.Connection, subject, continuation string) (*personalAccessTokenPage, error) {
	page, err := fetchPersonalAccessTokensByRoute(ctx, connection, subject, continuation)
	if err != nil {
		return nil, &patLocationError{err: err}
	}
	return page, nil
}

func fetchPersonalAccessTokensByClient(ctx context.Context, client tokenadmin.Client, subject, continuation string) (*personalAccessTokenPage, error) {
	isPublic := false
	pageSize := patPageSize
	args := tokenadmin.ListPersonalAccessTokensArgs{
		SubjectDescriptor: &subject,
		PageSize:          &pageSize,
		IsPublic:          &isPublic,
	}
	if continuation != "" {
		args.ContinuationToken = &continuation
	}
	page, err := client.ListPersonalAccessTokens(ctx, args)
	if err != nil || page == nil {
		return nil, err
	}
	result := &personalAccessTokenPage{}
	if page.ContinuationToken != nil && *page.ContinuationToken != uuid.Nil {
		result.ContinuationToken = page.ContinuationToken.String()
	}
	if page.Value != nil {
		result.Value = *page.Value
	}
	return result, nil
}

// fetchPersonalAccessTokensByRoute calls the on-prem route directly. Azure
// DevOps Server often does not register the v7 token-admin location id, and
// the documented server API version is 6.0-preview.1.
func fetchPersonalAccessTokensByRoute(ctx context.Context, connection *azuredevops.Connection, subject, continuation string) (*personalAccessTokenPage, error) {
	endpoint := personalAccessTokenEndpoint(connection.BaseUrl, subject, continuation, patPageSize)
	httpClient := connection.GetClientByUrl(connection.BaseUrl)
	req, err := httpClient.CreateRequestMessage(ctx, http.MethodGet, endpoint, "", nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.SendRequest(req)
	if err != nil {
		return nil, err
	}
	var raw rawPersonalAccessTokenPage
	if err := httpClient.UnmarshalBody(resp, &raw); err != nil {
		return nil, err
	}
	return &personalAccessTokenPage{
		ContinuationToken: continuationTokenString(raw.ContinuationToken),
		Value:             raw.Value,
	}, nil
}

func personalAccessTokenEndpoint(base, subject, continuation string, pageSize int) string {
	query := url.Values{}
	query.Set("api-version", patAPIVersion)
	query.Set("isPublic", "false")
	query.Set("pageSize", strconv.Itoa(pageSize))
	if continuation != "" {
		query.Set("continuationToken", continuation)
	}
	return strings.TrimRight(base, "/") + "/_apis/tokenadmin/personalaccesstokens/" + url.PathEscape(subject) + "?" + query.Encode()
}

func continuationTokenString(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}
	return strings.Trim(trimmed, `"`)
}

func patRouteMissing(d *plugin.QueryData, err error) bool {
	if !patLocationMissing(err) || !patRouteAbsent(err) {
		return false
	}
	config := GetConfig(connectionConfig(d))
	return resolveCollectionPlan(config).Multi && qualString(d, "collection_url") == ""
}

func patLocationMissing(err error) bool {
	var location *patLocationError
	return errors.As(err, &location)
}

func patRouteAbsent(err error) bool {
	return apiNotRegistered(err) || isNotFound(err)
}

func tokenAdminCollectionURLs(ctx context.Context, d *plugin.QueryData) ([]string, error) {
	config := GetConfig(connectionConfig(d))
	plan := resolveCollectionPlan(config)
	if !plan.Multi {
		return nil, nil
	}
	discovered, err := discoverProjectCollections(ctx, d, plan.ServerURL)
	if err != nil {
		return nil, err
	}
	selected, err := selectCollections(plan.ServerURL, discovered, plan.Allow, plan.Deny)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(selected))
	for _, collection := range selected {
		if collection.URL != "" {
			urls = append(urls, collection.URL)
		}
	}
	return urls, nil
}

func connectionForURL(d *plugin.QueryData, rawURL string) (*azuredevops.Connection, error) {
	token := personalAccessToken(GetConfig(connectionConfig(d)))
	if token == "" || strings.TrimSpace(rawURL) == "" {
		return nil, errors.New("'organization_url' and 'personal_access_token' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}
	return azuredevops.NewPatConnection(rawURL, token), nil
}

func sameBaseURL(left, right string) bool {
	return strings.TrimRight(left, "/") == strings.TrimRight(right, "/")
}

func patStatus(err error) (int, bool) {
	var wrapped *azuredevops.WrappedError
	if errors.As(err, &wrapped) && wrapped != nil && wrapped.StatusCode != nil {
		return *wrapped.StatusCode, true
	}
	var value azuredevops.WrappedError
	if errors.As(err, &value) && value.StatusCode != nil {
		return *value.StatusCode, true
	}
	return 0, false
}

func skipPatUser(err error) bool {
	code, ok := patStatus(err)
	return ok && (code == http.StatusNotFound || code == http.StatusBadRequest)
}

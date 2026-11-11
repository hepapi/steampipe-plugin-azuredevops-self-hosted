package azuredevops

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
)

// readIdentities calls the Identity ReadIdentities API. Azure DevOps Server
// 2022.1 serves /_apis/identities but does not register the v7 location id
// the generated client looks up, so a missing location falls back to that route.
func readIdentities(ctx context.Context, connection *azuredevops.Connection, args identity.ReadIdentitiesArgs) (*[]identity.Identity, error) {
	client, err := identity.NewClient(ctx, connection)
	if err != nil {
		if !apiNotRegistered(err) {
			return nil, err
		}
		return readIdentitiesByRoute(ctx, connection, args)
	}

	items, err := client.ReadIdentities(ctx, args)
	if err != nil && apiNotRegistered(err) {
		return readIdentitiesByRoute(ctx, connection, args)
	}
	return items, err
}

func apiNotRegistered(err error) bool {
	var location *azuredevops.LocationIdNotRegisteredError
	var area *azuredevops.ResourceAreaIdNotRegisteredError
	return errors.As(err, &location) || errors.As(err, &area)
}

func readIdentitiesByRoute(ctx context.Context, connection *azuredevops.Connection, args identity.ReadIdentitiesArgs) (*[]identity.Identity, error) {
	endpoint := strings.TrimRight(connection.BaseUrl, "/") + "/_apis/identities?" + identityQueryValues(args).Encode()
	httpClient := connection.GetClientByUrl(connection.BaseUrl)
	req, err := httpClient.CreateRequestMessage(ctx, http.MethodGet, endpoint, "", nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.SendRequest(req)
	if err != nil {
		return nil, err
	}

	var items []identity.Identity
	if err := httpClient.UnmarshalCollectionBody(resp, &items); err != nil {
		return nil, err
	}
	return &items, nil
}

// identityQueryValues builds the ReadIdentities query. api-version 6.0 is the
// version this Server collection accepts on /_apis/identities.
func identityQueryValues(args identity.ReadIdentitiesArgs) url.Values {
	query := url.Values{}
	query.Set("api-version", "6.0")
	if args.Descriptors != nil {
		query.Set("descriptors", *args.Descriptors)
	}
	if args.IdentityIds != nil {
		query.Set("identityIds", *args.IdentityIds)
	}
	if args.SubjectDescriptors != nil {
		query.Set("subjectDescriptors", *args.SubjectDescriptors)
	}
	if args.SearchFilter != nil {
		query.Set("searchFilter", *args.SearchFilter)
	}
	if args.FilterValue != nil {
		query.Set("filterValue", *args.FilterValue)
	}
	if args.QueryMembership != nil {
		query.Set("queryMembership", string(*args.QueryMembership))
	}
	if args.Properties != nil {
		query.Set("properties", *args.Properties)
	}
	return query
}

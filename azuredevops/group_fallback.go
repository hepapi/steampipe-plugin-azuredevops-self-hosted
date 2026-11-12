package azuredevops

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// graphGroupAPIVersions are the preview versions Azure DevOps Server accepts on
// /_apis/graph/groups. The generated client asks for 7.1-preview.1 by location
// id ebbe6af8-0b91-4c13-8cf1-777c14858188, and Server 2022.1 often does not
// register that location.
var graphGroupAPIVersions = []string{"6.0-preview.1", "7.0-preview.1", "7.1-preview.1"}

// identityGroupProperties are the Identity profile fields mapped onto
// principal_name, mail_address, domain, and description when Graph is unavailable.
const identityGroupProperties = "Account,Mail,Domain,Description,ScopeId"

// groupListItem is one SQL row. Graph fields stay on the embedded group.
// MembershipActive and Memberships are set when those values already came from
// the Identity service, so the hydrates do not call Graph again.
type groupListItem struct {
	graph.GraphGroup
	MembershipActive *bool
	Memberships      *[]graph.GraphMembership
}

func streamGraphGroupsByRoute(ctx context.Context, d *plugin.QueryData, connection *azuredevops.Connection) error {
	var version string
	var continuation string
	seenTokens := map[string]struct{}{}
	for {
		page, used, err := fetchGraphGroupPage(ctx, connection, continuation, version)
		if err != nil {
			return err
		}
		version = used
		if page == nil || page.GraphGroups == nil {
			return nil
		}
		for _, group := range *page.GraphGroups {
			d.StreamListItem(ctx, groupListItem{GraphGroup: group})
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
		continuation = firstContinuation(page.ContinuationToken)
		if continuation == "" {
			return nil
		}
		if _, exists := seenTokens[continuation]; exists {
			return nil
		}
		seenTokens[continuation] = struct{}{}
	}
}

func fetchGraphGroupPage(ctx context.Context, connection *azuredevops.Connection, continuation, version string) (*graph.PagedGraphGroups, string, error) {
	versions := graphGroupAPIVersions
	if version != "" {
		versions = []string{version}
	}
	var last error
	for _, candidate := range versions {
		page, err := readGraphGroupPage(ctx, connection, graphGroupsEndpoint(connection.BaseUrl, continuation, candidate))
		if err == nil {
			return page, candidate, nil
		}
		last = err
		if version != "" || isNotFound(err) || !unsupportedAPIVersion(err) {
			return nil, "", err
		}
	}
	return nil, "", last
}

func getGraphGroupByRoute(ctx context.Context, connection *azuredevops.Connection, descriptor string) (*graph.GraphGroup, error) {
	var last error
	for _, version := range graphGroupAPIVersions {
		group, err := readGraphGroup(ctx, connection, graphGroupEndpoint(connection.BaseUrl, descriptor, version))
		if err == nil {
			return group, nil
		}
		last = err
		if isNotFound(err) || !unsupportedAPIVersion(err) {
			return nil, err
		}
	}
	return nil, last
}

func readGraphGroupPage(ctx context.Context, connection *azuredevops.Connection, endpoint string) (*graph.PagedGraphGroups, error) {
	httpClient := connection.GetClientByUrl(connection.BaseUrl)
	req, err := httpClient.CreateRequestMessage(ctx, http.MethodGet, endpoint, "", nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.SendRequest(req)
	if err != nil {
		return nil, err
	}
	token := resp.Header.Get("X-MS-ContinuationToken")
	var groups []graph.GraphGroup
	if err := httpClient.UnmarshalCollectionBody(resp, &groups); err != nil {
		return nil, err
	}
	page := &graph.PagedGraphGroups{GraphGroups: &groups}
	if strings.TrimSpace(token) != "" {
		page.ContinuationToken = &[]string{strings.TrimSpace(token)}
	}
	return page, nil
}

func readGraphGroup(ctx context.Context, connection *azuredevops.Connection, endpoint string) (*graph.GraphGroup, error) {
	httpClient := connection.GetClientByUrl(connection.BaseUrl)
	req, err := httpClient.CreateRequestMessage(ctx, http.MethodGet, endpoint, "", nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.SendRequest(req)
	if err != nil {
		return nil, err
	}
	var group graph.GraphGroup
	if err := httpClient.UnmarshalBody(resp, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

func graphGroupsEndpoint(base, continuation, apiVersion string) string {
	query := url.Values{}
	query.Set("api-version", apiVersion)
	if continuation != "" {
		query.Set("continuationToken", continuation)
	}
	return strings.TrimRight(base, "/") + "/_apis/graph/groups?" + query.Encode()
}

func graphGroupEndpoint(base, descriptor, apiVersion string) string {
	query := url.Values{}
	query.Set("api-version", apiVersion)
	return strings.TrimRight(base, "/") + "/_apis/graph/groups/" + url.PathEscape(descriptor) + "?" + query.Encode()
}

// streamGroupsFromIdentity lists groups by walking Valid Users. Azure DevOps
// Server 2022.1 does not register the Graph groups location, and the REST route
// is missing on some collections. The Identity API has no unfiltered group list,
// so the walk follows direct membership of the built-in Valid Users group and
// keeps every container it finds.
func streamGroupsFromIdentity(ctx context.Context, d *plugin.QueryData, connection *azuredevops.Connection) error {
	roots, err := listCollectionGroups(ctx, connection)
	if err != nil {
		return err
	}

	queued := map[string]struct{}{}
	emitted := map[string]struct{}{}
	var queue []string
	for _, group := range roots {
		descriptor := strings.TrimSpace(stringValue(group.Descriptor))
		if descriptor == "" {
			continue
		}
		if _, exists := queued[descriptor]; exists {
			continue
		}
		queued[descriptor] = struct{}{}
		queue = append(queue, descriptor)
	}

	direct := identity.QueryMembershipValues.Direct
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
		end := patIdentityChunk
		if end > len(queue) {
			end = len(queue)
		}
		batch := append([]string(nil), queue[:end]...)
		queue = queue[end:]
		joined := strings.Join(batch, ",")
		properties := identityGroupProperties
		callCtx, cancel := patCallContext(ctx)
		items, err := readIdentities(callCtx, connection, identity.ReadIdentitiesArgs{
			Descriptors:     &joined,
			QueryMembership: &direct,
			Properties:      &properties,
		})
		cancel()
		if err != nil {
			return err
		}
		var identities []identity.Identity
		if items != nil {
			identities = *items
		}
		groups, next := classifyCollectionGroups(identities, queued, emitted)
		queue = append(queue, next...)
		for _, group := range groups {
			d.StreamListItem(ctx, group)
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
	}
	return nil
}

func listCollectionGroups(ctx context.Context, connection *azuredevops.Connection) ([]identity.Identity, error) {
	groups, err := listValidUserGroups(ctx, connection)
	if err != nil && validUsersGroupNotFound(err) {
		return nil, errGroupRootsNotFound
	}
	return groups, err
}

var errGroupRootsNotFound = groupRootError("azuredevopssh_group could not find Team Foundation Valid Users or Project Collection Valid Users")

type groupRootError string

func (e groupRootError) Error() string {
	return string(e)
}

func classifyCollectionGroups(items []identity.Identity, queued, emitted map[string]struct{}) ([]groupListItem, []string) {
	var groups []groupListItem
	var next []string
	for _, item := range items {
		row, ok := graphGroupFromIdentity(item)
		if !ok {
			continue
		}
		descriptor := groupDescriptor(row)
		if descriptor == "" {
			continue
		}
		legacy := strings.TrimSpace(stringValue(row.LegacyDescriptor))
		if _, exists := emitted[descriptor]; exists {
			continue
		}
		if legacy != "" {
			if _, exists := emitted[legacy]; exists {
				continue
			}
		}
		emitted[descriptor] = struct{}{}
		if legacy != "" {
			emitted[legacy] = struct{}{}
		}
		groups = append(groups, row)
	}
	for _, item := range items {
		if item.IsContainer == nil || !*item.IsContainer || item.Members == nil {
			continue
		}
		for _, member := range *item.Members {
			member = strings.TrimSpace(member)
			if member == "" {
				continue
			}
			if _, queuedAlready := queued[member]; queuedAlready {
				continue
			}
			if _, emittedAlready := emitted[member]; emittedAlready {
				continue
			}
			queued[member] = struct{}{}
			next = append(next, member)
		}
	}
	return groups, next
}

func graphGroupFromIdentity(item identity.Identity) (groupListItem, bool) {
	if item.IsContainer == nil || !*item.IsContainer {
		return groupListItem{}, false
	}
	legacy := strings.TrimSpace(stringValue(item.Descriptor))
	descriptor := strings.TrimSpace(stringValue(item.SubjectDescriptor))
	if descriptor == "" {
		descriptor = legacy
	}
	if descriptor == "" {
		return groupListItem{}, false
	}

	principal := identityProperty(item, "Account", "Mail")
	origin := identityProperty(item, "Origin")
	if origin == "" {
		origin = originFromDescriptor(descriptor)
	}
	originID := identityProperty(item, "ObjectId", "SID", "Puid")
	if originID == "" {
		originID = uuidString(item.Id)
	}
	domain := identityProperty(item, "ScopeId", "Domain")
	kind := "group"
	row := groupListItem{
		GraphGroup: graph.GraphGroup{
			Descriptor:       optionalString(descriptor),
			LegacyDescriptor: optionalString(legacy),
			DisplayName:      optionalString(identityDisplayName(item)),
			PrincipalName:    optionalString(principal),
			MailAddress:      optionalString(identityProperty(item, "Mail")),
			Domain:           optionalString(domain),
			Origin:           optionalString(strings.ToLower(origin)),
			OriginId:         optionalString(originID),
			Description:      optionalString(identityProperty(item, "Description")),
			SubjectKind:      &kind,
		},
		MembershipActive: item.IsActive,
	}
	if item.MemberOf != nil {
		memberships := membershipsFromMemberOf(descriptor, item.MemberOf)
		row.Memberships = &memberships
	}
	return row, true
}

func identityGroupForDescriptor(ctx context.Context, connection *azuredevops.Connection, descriptor string) (*identity.Identity, error) {
	direct := identity.QueryMembershipValues.Direct
	properties := identityGroupProperties
	callCtx, cancel := patCallContext(ctx)
	items, err := readIdentities(callCtx, connection, identity.ReadIdentitiesArgs{
		SubjectDescriptors: &descriptor,
		QueryMembership:    &direct,
		Properties:         &properties,
	})
	cancel()
	if item, ok := firstIdentity(items); ok && err == nil {
		return item, nil
	}
	firstErr := err
	callCtx, cancel = patCallContext(ctx)
	items, err = readIdentities(callCtx, connection, identity.ReadIdentitiesArgs{
		Descriptors:     &descriptor,
		QueryMembership: &direct,
		Properties:      &properties,
	})
	cancel()
	if item, ok := firstIdentity(items); ok && err == nil {
		return item, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, firstErr
}

func groupFromHydrate(h *plugin.HydrateData) groupListItem {
	if h == nil || h.Item == nil {
		return groupListItem{}
	}
	switch item := h.Item.(type) {
	case groupListItem:
		return item
	case *groupListItem:
		if item == nil {
			return groupListItem{}
		}
		return *item
	case graph.GraphGroup:
		return groupListItem{GraphGroup: item}
	case *graph.GraphGroup:
		if item == nil {
			return groupListItem{}
		}
		return groupListItem{GraphGroup: *item}
	default:
		return groupListItem{}
	}
}

func groupDescriptor(group groupListItem) string {
	if descriptor := strings.TrimSpace(stringValue(group.Descriptor)); descriptor != "" {
		return descriptor
	}
	return strings.TrimSpace(stringValue(group.LegacyDescriptor))
}

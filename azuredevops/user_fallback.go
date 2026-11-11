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

// graphUserAPIVersions are the preview versions Azure DevOps Server accepts on
// /_apis/graph/users. The generated client asks for 7.1-preview.1 by location
// id, and Server 2022.1 often does not register that location.
var graphUserAPIVersions = []string{"6.0-preview.1", "7.0-preview.1", "7.1-preview.1"}

// identityUserProperties are the Identity profile fields mapped onto
// principal_name, mail_address, and domain when Graph is unavailable.
const identityUserProperties = "Account,Mail,Domain"

// userListItem is one SQL row. Graph fields stay on the embedded user.
// MembershipActive and Memberships are set when those values already came from
// the Identity service, so the hydrates do not call Graph again.
type userListItem struct {
	graph.GraphUser
	MembershipActive *bool
	Memberships      *[]graph.GraphMembership
}

func streamGraphUsersByRoute(ctx context.Context, d *plugin.QueryData, connection *azuredevops.Connection) error {
	var version string
	var continuation string
	seenTokens := map[string]struct{}{}
	for {
		page, used, err := fetchGraphUserPage(ctx, connection, continuation, version)
		if err != nil {
			return err
		}
		version = used
		if page == nil || page.GraphUsers == nil {
			return nil
		}
		for _, user := range *page.GraphUsers {
			d.StreamListItem(ctx, userListItem{GraphUser: user})
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

func fetchGraphUserPage(ctx context.Context, connection *azuredevops.Connection, continuation, version string) (*graph.PagedGraphUsers, string, error) {
	versions := graphUserAPIVersions
	if version != "" {
		versions = []string{version}
	}
	var last error
	for _, candidate := range versions {
		page, err := readGraphUserPage(ctx, connection, graphUsersEndpoint(connection.BaseUrl, continuation, candidate))
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

func getGraphUserByRoute(ctx context.Context, connection *azuredevops.Connection, descriptor string) (*graph.GraphUser, error) {
	var last error
	for _, version := range graphUserAPIVersions {
		user, err := readGraphUser(ctx, connection, graphUserEndpoint(connection.BaseUrl, descriptor, version))
		if err == nil {
			return user, nil
		}
		last = err
		if isNotFound(err) || !unsupportedAPIVersion(err) {
			return nil, err
		}
	}
	return nil, last
}

func readGraphUserPage(ctx context.Context, connection *azuredevops.Connection, endpoint string) (*graph.PagedGraphUsers, error) {
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
	var users []graph.GraphUser
	if err := httpClient.UnmarshalCollectionBody(resp, &users); err != nil {
		return nil, err
	}
	page := &graph.PagedGraphUsers{GraphUsers: &users}
	if strings.TrimSpace(token) != "" {
		page.ContinuationToken = &[]string{strings.TrimSpace(token)}
	}
	return page, nil
}

func readGraphUser(ctx context.Context, connection *azuredevops.Connection, endpoint string) (*graph.GraphUser, error) {
	httpClient := connection.GetClientByUrl(connection.BaseUrl)
	req, err := httpClient.CreateRequestMessage(ctx, http.MethodGet, endpoint, "", nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.SendRequest(req)
	if err != nil {
		return nil, err
	}
	var user graph.GraphUser
	if err := httpClient.UnmarshalBody(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func graphUsersEndpoint(base, continuation, apiVersion string) string {
	query := url.Values{}
	query.Set("api-version", apiVersion)
	if continuation != "" {
		query.Set("continuationToken", continuation)
	}
	return strings.TrimRight(base, "/") + "/_apis/graph/users?" + query.Encode()
}

func graphUserEndpoint(base, descriptor, apiVersion string) string {
	query := url.Values{}
	query.Set("api-version", apiVersion)
	return strings.TrimRight(base, "/") + "/_apis/graph/users/" + url.PathEscape(descriptor) + "?" + query.Encode()
}

func graphRouteMissing(err error) bool {
	return apiNotRegistered(err) || isNotFound(err)
}

func unsupportedAPIVersion(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "api version") || strings.Contains(message, "api-version")
}

func firstContinuation(tokens *[]string) string {
	if tokens == nil || len(*tokens) == 0 {
		return ""
	}
	return strings.TrimSpace((*tokens)[0])
}

// streamUsersFromIdentity lists users by walking Valid Users. Azure DevOps
// Server 2022.1 does not register the Graph users location, and the REST route
// is missing on some collections. The Identity API has no unfiltered user list,
// so the walk follows direct membership of the built-in Valid Users group.
func streamUsersFromIdentity(ctx context.Context, d *plugin.QueryData, connection *azuredevops.Connection) error {
	groups, err := listCollectionUserGroups(ctx, connection)
	if err != nil {
		return err
	}

	queued := map[string]struct{}{}
	emitted := map[string]struct{}{}
	var queue []string
	for _, group := range groups {
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
		properties := identityUserProperties
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
		users, next := classifyCollectionUsers(identities, queued, emitted)
		queue = append(queue, next...)
		for _, user := range users {
			d.StreamListItem(ctx, user)
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
	}
	return nil
}

func listCollectionUserGroups(ctx context.Context, connection *azuredevops.Connection) ([]identity.Identity, error) {
	groups, err := listValidUserGroups(ctx, connection)
	if err != nil && validUsersGroupNotFound(err) {
		return nil, errUserGroupsNotFound
	}
	return groups, err
}

var errUserGroupsNotFound = userGroupError("azuredevopssh_user could not find Team Foundation Valid Users or Project Collection Valid Users")

type userGroupError string

func (e userGroupError) Error() string {
	return string(e)
}

func classifyCollectionUsers(items []identity.Identity, queued, emitted map[string]struct{}) ([]userListItem, []string) {
	var users []userListItem
	var next []string
	for _, item := range items {
		row, ok := graphUserFromIdentity(item)
		if !ok {
			continue
		}
		descriptor := userDescriptor(row)
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
		users = append(users, row)
	}
	for _, item := range items {
		if _, ok := graphUserFromIdentity(item); ok || item.Members == nil {
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
	return users, next
}

func graphUserFromIdentity(item identity.Identity) (userListItem, bool) {
	if item.IsContainer != nil && *item.IsContainer {
		return userListItem{}, false
	}
	legacy := strings.TrimSpace(stringValue(item.Descriptor))
	descriptor := strings.TrimSpace(stringValue(item.SubjectDescriptor))
	if descriptor == "" {
		descriptor = legacy
	}
	if descriptor == "" {
		return userListItem{}, false
	}

	principal := identityProperty(item, "Account", "Mail")
	origin := identityProperty(item, "Origin")
	if origin == "" {
		origin = originFromDescriptor(descriptor)
	}
	alias := identityProperty(item, "SamAccountName", "DirectoryAlias")
	if alias == "" {
		alias = accountAlias(principal)
	}
	kind := "user"
	row := userListItem{
		GraphUser: graph.GraphUser{
			Descriptor:       optionalString(descriptor),
			LegacyDescriptor: optionalString(legacy),
			DisplayName:      optionalString(identityDisplayName(item)),
			PrincipalName:    optionalString(principal),
			MailAddress:      optionalString(identityProperty(item, "Mail")),
			Domain:           optionalString(identityProperty(item, "Domain")),
			Origin:           optionalString(strings.ToLower(origin)),
			DirectoryAlias:   optionalString(alias),
			OriginId:         optionalString(identityProperty(item, "ObjectId", "SID", "Puid")),
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

func membershipsFromMemberOf(member string, memberOf *[]string) []graph.GraphMembership {
	if memberOf == nil {
		return []graph.GraphMembership{}
	}
	var memberships []graph.GraphMembership
	for _, container := range *memberOf {
		container = strings.TrimSpace(container)
		if container == "" {
			continue
		}
		containerDescriptor := container
		memberDescriptor := member
		memberships = append(memberships, graph.GraphMembership{
			ContainerDescriptor: &containerDescriptor,
			MemberDescriptor:    &memberDescriptor,
		})
	}
	if memberships == nil {
		return []graph.GraphMembership{}
	}
	return memberships
}

func userFromHydrate(h *plugin.HydrateData) userListItem {
	if h == nil || h.Item == nil {
		return userListItem{}
	}
	switch item := h.Item.(type) {
	case userListItem:
		return item
	case *userListItem:
		if item == nil {
			return userListItem{}
		}
		return *item
	case graph.GraphUser:
		return userListItem{GraphUser: item}
	case *graph.GraphUser:
		if item == nil {
			return userListItem{}
		}
		return userListItem{GraphUser: *item}
	default:
		return userListItem{}
	}
}

func userDescriptor(user userListItem) string {
	if descriptor := strings.TrimSpace(stringValue(user.Descriptor)); descriptor != "" {
		return descriptor
	}
	return strings.TrimSpace(stringValue(user.LegacyDescriptor))
}

func originFromDescriptor(descriptor string) string {
	switch {
	case strings.HasPrefix(descriptor, "aad."):
		return "aad"
	case strings.HasPrefix(descriptor, "msa."):
		return "msa"
	case strings.HasPrefix(descriptor, "win."):
		return "ad"
	case strings.HasPrefix(descriptor, "svc."), strings.HasPrefix(descriptor, "imp."):
		return "vsts"
	default:
		return ""
	}
}

func accountAlias(principal string) string {
	principal = strings.TrimSpace(principal)
	slash := strings.LastIndex(principal, `\`)
	if slash < 0 || slash == len(principal)-1 {
		return ""
	}
	return principal[slash+1:]
}

func identityProperty(item identity.Identity, names ...string) string {
	props, ok := item.Properties.(map[string]interface{})
	if !ok || len(props) == 0 {
		return ""
	}
	folded := make(map[string]interface{}, len(props))
	for key, value := range props {
		folded[strings.ToLower(key)] = value
	}
	for _, name := range names {
		value := propertyValue(folded[strings.ToLower(name)])
		if value != "" {
			return value
		}
	}
	return ""
}

func propertyValue(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case map[string]interface{}:
		if nested, ok := value["$value"]; ok {
			return propertyValue(nested)
		}
		if nested, ok := value["value"]; ok {
			return propertyValue(nested)
		}
	}
	return ""
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

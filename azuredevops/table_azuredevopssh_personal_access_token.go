package azuredevops

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/delegatedauthorization"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

const (
	patIdentityChunk = 20
	patCallTimeout   = time.Minute
)

var patValidUserGroups = []string{
	"Team Foundation Valid Users",
	"Project Collection Valid Users",
}

type patUser struct {
	SubjectDescriptor string
	DisplayName       string
}

// personalAccessTokenRow is the SQL row. It has no secret fields. token and
// alternateToken on the API payload are cleared before the copy.
type personalAccessTokenRow struct {
	AuthorizationID     string
	DisplayName         string
	Scope               string
	ValidFrom           *time.Time
	ValidTo             *time.Time
	IsValid             *bool
	UserID              string
	SubjectDescriptor   string
	UserDisplayName     string
	AccessID            string
	ClientID            string
	HostAuthorizationID string
	Source              string
	TargetAccounts      []string
}

func tableAzureDevOpsPersonalAccessToken(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_personal_access_token",
		Description: "Retrieve personal access token metadata, including expiration, without the secret value.",
		List: &plugin.ListConfig{
			Hydrate: listPersonalAccessTokens,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "subject_descriptor", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "authorization_id",
				Description: "Unique identifier of the personal access token.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "display_name",
				Description: "Name given to the personal access token when it was created.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "scope",
				Description: "Space-separated OAuth scopes granted to the token.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "valid_from",
				Description: "When the token was created.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "valid_to",
				Description: "When the token expires.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "is_valid",
				Description: "True when the token has not expired and has not been revoked.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "user_id",
				Description: "Storage key of the user who owns the token.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "subject_descriptor",
				Description: "Subject descriptor of the user who owns the token. Set this to list one user.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "user_display_name",
				Description: "Display name of the user who owns the token.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "access_id",
				Description: "Access identifier associated with the token.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "client_id",
				Description: "Client identifier associated with the token.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "host_authorization_id",
				Description: "Host authorization identifier associated with the token.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "source",
				Description: "Source of the session token.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "target_accounts",
				Description: "Organizations the token is valid for. Empty when the token applies to every organization the user can access.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DisplayName"),
			},
		}),
	}
}

func listPersonalAccessTokens(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_personal_access_token.listPersonalAccessTokens", "connection_error", err)
		return nil, err
	}

	lister := &personalAccessTokenLister{connection: connection}
	seenTokens := map[string]struct{}{}
	if subject := qualString(d, "subject_descriptor"); subject != "" {
		err = streamOneUserTokens(ctx, d, lister, patUserForSubject(ctx, connection, subject), seenTokens)
		if err != nil {
			plugin.Logger(ctx).Error("azuredevopssh_personal_access_token.listPersonalAccessTokens", "api_error", err)
		}
		return nil, err
	}

	err = streamOrganizationTokens(ctx, d, connection, lister, seenTokens)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_personal_access_token.listPersonalAccessTokens", "identity_error", err)
	}
	return nil, err
}

func streamOneUserTokens(ctx context.Context, d *plugin.QueryData, lister *personalAccessTokenLister, user patUser, seenTokens map[string]struct{}) error {
	err := streamPersonalAccessTokens(ctx, d, lister, user, seenTokens)
	if skipPatUser(err) {
		plugin.Logger(ctx).Warn("azuredevopssh_personal_access_token.listPersonalAccessTokens", "subject_descriptor", user.SubjectDescriptor, "api_error", err)
		return nil
	}
	return err
}

func streamPersonalAccessTokens(ctx context.Context, d *plugin.QueryData, lister *personalAccessTokenLister, user patUser, seen map[string]struct{}) error {
	continuation := ""
	for pageNumber := 0; pageNumber < patMaxPages; pageNumber++ {
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
		page, err := lister.list(ctx, d, user.SubjectDescriptor, continuation)
		if err != nil {
			return err
		}
		if page == nil {
			return nil
		}
		for i := range page.Value {
			row := personalAccessTokenRowFrom(&page.Value[i], user)
			if row.AuthorizationID != "" {
				if _, exists := seen[row.AuthorizationID]; exists {
					continue
				}
				seen[row.AuthorizationID] = struct{}{}
			}
			d.StreamListItem(ctx, row)
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
		if page.ContinuationToken == "" || page.ContinuationToken == continuation {
			return nil
		}
		continuation = page.ContinuationToken
	}
	return nil
}

func streamOrganizationTokens(ctx context.Context, d *plugin.QueryData, connection *azuredevops.Connection, lister *personalAccessTokenLister, seenTokens map[string]struct{}) error {
	err := streamValidUserTokens(ctx, d, connection, lister, seenTokens)
	if err == nil || !shouldRetryUsersOnCollection(d, err) {
		return err
	}

	collections, ferr := tokenAdminCollectionURLs(ctx, d)
	if ferr != nil {
		return ferr
	}
	lastErr := err
	foundAny := false
	for _, next := range collections {
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.RowsRemaining(ctx) == 0 {
			return nil
		}
		if sameBaseURL(next, connection.BaseUrl) {
			continue
		}
		shifted, cerr := connectionForURL(d, next)
		if cerr != nil {
			return cerr
		}
		if !lister.settled {
			lister.connection = shifted
		}
		walkErr := streamValidUserTokens(ctx, d, shifted, lister, seenTokens)
		if walkErr != nil {
			lastErr = walkErr
			if apiNotRegistered(walkErr) || isNotFound(walkErr) || validUsersGroupNotFound(walkErr) {
				continue
			}
			return walkErr
		}
		foundAny = true
	}
	if !foundAny {
		return lastErr
	}
	return nil
}

// streamValidUserTokens walks direct group membership and lists tokens as soon
// as a user is found. Expanded membership of Valid Users is one request that
// can sit until the identity service has walked the whole deployment, so the
// query never reaches its first row.
func streamValidUserTokens(ctx context.Context, d *plugin.QueryData, connection *azuredevops.Connection, lister *personalAccessTokenLister, seenTokens map[string]struct{}) error {
	groups, err := listValidUserGroups(ctx, connection)
	if err != nil {
		return err
	}

	seen := map[string]struct{}{}
	var queue []string
	for _, group := range groups {
		descriptor := strings.TrimSpace(stringValue(group.Descriptor))
		if descriptor == "" {
			continue
		}
		if _, exists := seen[descriptor]; exists {
			continue
		}
		seen[descriptor] = struct{}{}
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
		callCtx, cancel := patCallContext(ctx)
		items, err := readIdentities(callCtx, connection, identity.ReadIdentitiesArgs{
			Descriptors:     &joined,
			QueryMembership: &direct,
		})
		cancel()
		if err != nil {
			return err
		}
		var identities []identity.Identity
		if items != nil {
			identities = *items
		}
		users, next := classifyPatMembers(identities, seen)
		queue = append(queue, next...)
		for _, user := range users {
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
			if err := streamOneUserTokens(ctx, d, lister, user, seenTokens); err != nil {
				return err
			}
		}
	}
	return nil
}

func classifyPatMembers(items []identity.Identity, seen map[string]struct{}) ([]patUser, []string) {
	var users []patUser
	var next []string
	for _, item := range items {
		subject, ok := patSubject(item)
		if !ok {
			continue
		}
		if _, exists := seen[subject]; exists {
			continue
		}
		seen[subject] = struct{}{}
		if descriptor := strings.TrimSpace(stringValue(item.Descriptor)); descriptor != "" {
			seen[descriptor] = struct{}{}
		}
		users = append(users, patUser{
			SubjectDescriptor: subject,
			DisplayName:       identityDisplayName(item),
		})
	}
	for _, item := range items {
		if _, ok := patSubject(item); ok || item.Members == nil {
			continue
		}
		for _, member := range *item.Members {
			member = strings.TrimSpace(member)
			if member == "" {
				continue
			}
			if _, exists := seen[member]; exists {
				continue
			}
			seen[member] = struct{}{}
			next = append(next, member)
		}
	}
	return users, next
}

func shouldRetryUsersOnCollection(d *plugin.QueryData, err error) bool {
	if err == nil {
		return false
	}
	config := GetConfig(connectionConfig(d))
	if !resolveCollectionPlan(config).Multi || qualString(d, "collection_url") != "" {
		return false
	}
	return apiNotRegistered(err) || isNotFound(err) || validUsersGroupNotFound(err)
}

func validUsersGroupNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "could not find Team Foundation Valid Users")
}

func listValidUserGroups(ctx context.Context, connection *azuredevops.Connection) ([]identity.Identity, error) {
	filter := "LocalGroupName"
	none := identity.QueryMembershipValues.None
	for _, name := range patValidUserGroups {
		groupName := name
		callCtx, cancel := patCallContext(ctx)
		items, err := readIdentities(callCtx, connection, identity.ReadIdentitiesArgs{
			SearchFilter:    &filter,
			FilterValue:     &groupName,
			QueryMembership: &none,
		})
		cancel()
		if err != nil {
			return nil, err
		}
		groups := containerIdentities(items)
		if len(groups) > 0 {
			return groups, nil
		}
	}
	return nil, fmt.Errorf("azuredevopssh_personal_access_token could not find Team Foundation Valid Users or Project Collection Valid Users")
}

func containerIdentities(items *[]identity.Identity) []identity.Identity {
	if items == nil {
		return nil
	}
	var groups []identity.Identity
	for _, item := range *items {
		if item.IsContainer != nil && *item.IsContainer && strings.TrimSpace(stringValue(item.Descriptor)) != "" {
			groups = append(groups, item)
		}
	}
	return groups
}

func patUserForSubject(ctx context.Context, connection *azuredevops.Connection, subject string) patUser {
	user := patUser{SubjectDescriptor: subject}
	item, err := identityForSubject(ctx, connection, subject)
	if err != nil {
		plugin.Logger(ctx).Warn("azuredevopssh_personal_access_token.listPersonalAccessTokens", "subject_descriptor", subject, "identity_error", err)
		return user
	}
	if item != nil {
		user.DisplayName = identityDisplayName(*item)
	}
	return user
}

func identityForSubject(ctx context.Context, connection *azuredevops.Connection, subject string) (*identity.Identity, error) {
	none := identity.QueryMembershipValues.None
	callCtx, cancel := patCallContext(ctx)
	items, err := readIdentities(callCtx, connection, identity.ReadIdentitiesArgs{
		SubjectDescriptors: &subject,
		QueryMembership:    &none,
	})
	cancel()
	if item, ok := firstIdentity(items); ok && err == nil {
		return item, nil
	}
	firstErr := err
	callCtx, cancel = patCallContext(ctx)
	items, err = readIdentities(callCtx, connection, identity.ReadIdentitiesArgs{
		Descriptors:     &subject,
		QueryMembership: &none,
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

func firstIdentity(items *[]identity.Identity) (*identity.Identity, bool) {
	if items == nil || len(*items) == 0 {
		return nil, false
	}
	item := (*items)[0]
	return &item, true
}

func patUsersFromIdentities(items []identity.Identity) []patUser {
	return dedupePatUsers(patUsersFromIdentitySlice(items))
}

func patUsersFromIdentitySlice(items []identity.Identity) []patUser {
	var users []patUser
	for _, item := range items {
		subject, ok := patSubject(item)
		if !ok {
			continue
		}
		users = append(users, patUser{
			SubjectDescriptor: subject,
			DisplayName:       identityDisplayName(item),
		})
	}
	return users
}

func dedupePatUsers(users []patUser) []patUser {
	seen := map[string]struct{}{}
	var unique []patUser
	for _, user := range users {
		if user.SubjectDescriptor == "" {
			continue
		}
		if _, exists := seen[user.SubjectDescriptor]; exists {
			continue
		}
		seen[user.SubjectDescriptor] = struct{}{}
		unique = append(unique, user)
	}
	return unique
}

func patSubject(item identity.Identity) (string, bool) {
	if item.IsContainer != nil && *item.IsContainer {
		return "", false
	}
	if subject := strings.TrimSpace(stringValue(item.SubjectDescriptor)); subject != "" {
		return subject, true
	}
	if descriptor := strings.TrimSpace(stringValue(item.Descriptor)); descriptor != "" {
		return descriptor, true
	}
	return "", false
}

func identityDisplayName(item identity.Identity) string {
	if name := strings.TrimSpace(stringValue(item.ProviderDisplayName)); name != "" {
		return name
	}
	return strings.TrimSpace(stringValue(item.CustomDisplayName))
}

func personalAccessTokenRowFrom(token *delegatedauthorization.SessionToken, user patUser) personalAccessTokenRow {
	row := personalAccessTokenRow{
		SubjectDescriptor: user.SubjectDescriptor,
		UserDisplayName:   user.DisplayName,
	}
	if token == nil {
		return row
	}
	token.Token = nil
	token.AlternateToken = nil
	row.AuthorizationID = uuidString(token.AuthorizationId)
	row.DisplayName = stringValue(token.DisplayName)
	row.Scope = stringValue(token.Scope)
	row.ValidFrom = azureTime(token.ValidFrom)
	row.ValidTo = azureTime(token.ValidTo)
	row.IsValid = token.IsValid
	row.UserID = uuidString(token.UserId)
	row.AccessID = uuidString(token.AccessId)
	row.ClientID = uuidString(token.ClientId)
	row.HostAuthorizationID = uuidString(token.HostAuthorizationId)
	row.Source = stringValue(token.Source)
	row.TargetAccounts = uuidStrings(token.TargetAccounts)
	return row
}

func uuidString(id *uuid.UUID) string {
	if id == nil || *id == uuid.Nil {
		return ""
	}
	return id.String()
}

func uuidStrings(ids *[]uuid.UUID) []string {
	if ids == nil {
		return nil
	}
	values := make([]string, 0, len(*ids))
	for _, id := range *ids {
		if id == uuid.Nil {
			continue
		}
		values = append(values, id.String())
	}
	return values
}

func patCallContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, patCallTimeout)
}

func azureTime(value *azuredevops.Time) *time.Time {
	if value == nil {
		return nil
	}
	stamp := value.Time
	return &stamp
}

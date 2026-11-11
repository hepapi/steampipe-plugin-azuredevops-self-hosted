package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsUser(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_user",
		Description: "Retrieve information about your users.",
		List: &plugin.ListConfig{
			Hydrate: listUsers,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("descriptor"),
			Hydrate:    getUser,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "principal_name",
				Description: "This is the PrincipalName of this graph member from the source provider. The source provider may change this field over time and it is not guaranteed to be immutable for the life of the graph member by VSTS.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "display_name",
				Description: "This is the non-unique display name of the graph subject. To change this field, you must alter its value in the source provider.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "membership_state",
				Description: "When true, the membership is active.",
				Type:        proto.ColumnType_BOOL,
				Hydrate:     getUserMembershipState,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "domain",
				Description: "This represents the name of the container of origin for a graph member. (For MSA this is Windows Live ID, for AD the name of the domain, for AAD the tenantID of the directory, for VSTS users the ScopeId, etc).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "origin",
				Description: "The type of source provider for the origin identifier (ex:AD, AAD, MSA).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "directory_alias",
				Description: "The short, generally unique name for the user in the backing directory. For AAD users, this corresponds to the mail nickname, which is often but not necessarily similar to the part of the user's mail address before the @ sign. For GitHub users, this corresponds to the GitHub user handle.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "descriptor",
				Description: "The descriptor is the primary way to reference the graph subject while the system is running. This field will uniquely identify the same graph subject across both Accounts and Organizations.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_deleted_in_origin",
				Description: "When true, the user has been deleted in the identity provider.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "legacy_descriptor",
				Description: "The legacy descriptor is here in case you need to access old version IMS using identity descriptor.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "mail_address",
				Description: "The email address of record for a given graph member. This may be different than the principal name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "meta_type",
				Description: "The meta type of the user in the origin, such as member, guest, etc.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "origin_id",
				Description: "The unique identifier from the system of origin. Typically a sid, object id or Guid. Linking and unlinking operations can cause this value to change for a user because the user is not backed by a different provider and has a different unique id in the new provider.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "subject_kind",
				Description: "This field identifies the type of the graph subject (ex: Group, Scope, User).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "This url is the full route to the source resource of this graph subject.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "links",
				Description: "This field contains zero or more interesting links about the graph subject. These links may be invoked to obtain additional relationships or more detailed information about this graph subject.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "memberships",
				Description: "Get all the memberships where this descriptor is a member in the relationship.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getUserMemberships,
				Transform:   transform.FromValue(),
			},

			/// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("DisplayName"),
			},
		}),
	}
}

func listUsers(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_user.listUsers", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevopssh_user.listUsers", "client_error", err)
		return nil, err
	}
	if err == nil {
		err = streamGraphUsers(ctx, d, client)
		if err == nil || !apiNotRegistered(err) {
			if err != nil {
				plugin.Logger(ctx).Error("azuredevopssh_user.listUsers", "api_error", err)
			}
			return nil, err
		}
	}

	routeErr := streamGraphUsersByRoute(ctx, d, connection)
	if routeErr == nil || !graphRouteMissing(routeErr) {
		if routeErr != nil {
			plugin.Logger(ctx).Error("azuredevopssh_user.listUsers", "api_error", routeErr)
		}
		return nil, routeErr
	}

	plugin.Logger(ctx).Warn("azuredevopssh_user.listUsers", "fallback", "identity", "reason", routeErr)
	if err = streamUsersFromIdentity(ctx, d, connection); err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_user.listUsers", "api_error", err)
		return nil, err
	}
	return nil, nil
}

func streamGraphUsers(ctx context.Context, d *plugin.QueryData, client graph.Client) error {
	input := graph.ListUsersArgs{}
	seenTokens := map[string]struct{}{}
	for {
		users, err := client.ListUsers(ctx, input)
		if err != nil {
			return err
		}
		if users == nil || users.GraphUsers == nil {
			return nil
		}
		for _, user := range *users.GraphUsers {
			d.StreamListItem(ctx, userListItem{GraphUser: user})
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
		token := firstContinuation(users.ContinuationToken)
		if token == "" {
			return nil
		}
		if _, exists := seenTokens[token]; exists {
			return nil
		}
		seenTokens[token] = struct{}{}
		input.ContinuationToken = &token
	}
}

func getUser(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	descriptor := d.EqualsQuals["descriptor"].GetStringValue()
	if descriptor == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUser", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUser", "client_error", err)
		return nil, err
	}
	if err == nil {
		output, getErr := client.GetUser(ctx, graph.GetUserArgs{UserDescriptor: &descriptor})
		if getErr == nil && output != nil {
			return userListItem{GraphUser: *output}, nil
		}
		if getErr != nil && !apiNotRegistered(getErr) {
			if absentFromCollection(d, getErr) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("azuredevopssh_user.getUser", "api_error", getErr)
			return nil, getErr
		}
	}

	output, routeErr := getGraphUserByRoute(ctx, connection, descriptor)
	if routeErr == nil && output != nil {
		return userListItem{GraphUser: *output}, nil
	}
	if routeErr != nil && !graphRouteMissing(routeErr) {
		if absentFromCollection(d, routeErr) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_user.getUser", "api_error", routeErr)
		return nil, routeErr
	}

	item, idErr := identityForSubject(ctx, connection, descriptor)
	if idErr != nil {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUser", "api_error", idErr)
		return nil, idErr
	}
	if item == nil {
		return nil, nil
	}
	row, ok := graphUserFromIdentity(*item)
	if !ok {
		return nil, nil
	}
	return row, nil
}

func getUserMembershipState(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	user := userFromHydrate(h)
	if user.MembershipActive != nil {
		return user.MembershipActive, nil
	}
	descriptor := userDescriptor(user)
	if descriptor == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUserMembershipState", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUserMembershipState", "client_error", err)
		return nil, err
	}
	if err == nil {
		output, stateErr := client.GetMembershipState(ctx, graph.GetMembershipStateArgs{SubjectDescriptor: &descriptor})
		if stateErr == nil {
			if output == nil {
				return nil, nil
			}
			return output.Active, nil
		}
		if !apiNotRegistered(stateErr) {
			plugin.Logger(ctx).Error("azuredevopssh_user.getUserMembershipState", "api_error", stateErr)
			return nil, stateErr
		}
	}

	item, idErr := identityForSubject(ctx, connection, descriptor)
	if idErr != nil {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUserMembershipState", "api_error", idErr)
		return nil, idErr
	}
	if item == nil {
		return nil, nil
	}
	return item.IsActive, nil
}

func getUserMemberships(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	user := userFromHydrate(h)
	if user.Memberships != nil {
		return user.Memberships, nil
	}
	descriptor := userDescriptor(user)
	if descriptor == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUserMemberships", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUserMemberships", "client_error", err)
		return nil, err
	}
	if err == nil {
		output, listErr := client.ListMemberships(ctx, graph.ListMembershipsArgs{SubjectDescriptor: &descriptor})
		if listErr == nil {
			return output, nil
		}
		if !apiNotRegistered(listErr) {
			plugin.Logger(ctx).Error("azuredevopssh_user.getUserMemberships", "api_error", listErr)
			return nil, listErr
		}
	}

	item, idErr := identityForMembership(ctx, connection, descriptor)
	if idErr != nil {
		plugin.Logger(ctx).Error("azuredevopssh_user.getUserMemberships", "api_error", idErr)
		return nil, idErr
	}
	memberships := membershipsFromMemberOf(descriptor, nil)
	if item != nil {
		memberships = membershipsFromMemberOf(descriptor, item.MemberOf)
	}
	return &memberships, nil
}

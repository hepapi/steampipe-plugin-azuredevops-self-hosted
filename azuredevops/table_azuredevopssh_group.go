package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGroup(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_group",
		Description: "Retrieve information about your groups.",
		List: &plugin.ListConfig{
			Hydrate: listGroups,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("descriptor"),
			Hydrate:    getGroup,
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
				Hydrate:     getGroupMembershipState,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "domain",
				Description: "This represents the name of the container of origin for a graph member. (For MSA this is Windows Live ID, for AD the name of the domain, for AAD the tenantID of the directory, for VSTS groups the ScopeId, etc).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "origin",
				Description: "The type of source provider for the origin identifier (ex:AD, AAD, MSA).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "A short phrase to help human readers disambiguate groups with similar names.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "descriptor",
				Description: "The descriptor is the primary way to reference the graph subject while the system is running. This field will uniquely identify the same graph subject across both Accounts and Organizations.",
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
				Hydrate:     getGroupMemberships,
				Transform:   transform.FromValue(),
			},

			/// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("PrincipalName"),
			},
		}),
	}
}

func listGroups(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_group.listGroups", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevopssh_group.listGroups", "client_error", err)
		return nil, err
	}
	if err == nil {
		err = streamGraphGroups(ctx, d, client)
		if err == nil || !apiNotRegistered(err) {
			if err != nil {
				plugin.Logger(ctx).Error("azuredevopssh_group.listGroups", "api_error", err)
			}
			return nil, err
		}
	}

	routeErr := streamGraphGroupsByRoute(ctx, d, connection)
	if routeErr == nil || !graphRouteMissing(routeErr) {
		if routeErr != nil {
			plugin.Logger(ctx).Error("azuredevopssh_group.listGroups", "api_error", routeErr)
		}
		return nil, routeErr
	}

	plugin.Logger(ctx).Warn("azuredevopssh_group.listGroups", "fallback", "identity", "reason", routeErr)
	if err = streamGroupsFromIdentity(ctx, d, connection); err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_group.listGroups", "api_error", err)
		return nil, err
	}
	return nil, nil
}

func streamGraphGroups(ctx context.Context, d *plugin.QueryData, client graph.Client) error {
	input := graph.ListGroupsArgs{}
	seenTokens := map[string]struct{}{}
	for {
		groups, err := client.ListGroups(ctx, input)
		if err != nil {
			return err
		}
		if groups == nil || groups.GraphGroups == nil {
			return nil
		}
		for _, group := range *groups.GraphGroups {
			d.StreamListItem(ctx, groupListItem{GraphGroup: group})
			if d.RowsRemaining(ctx) == 0 {
				return nil
			}
		}
		token := firstContinuation(groups.ContinuationToken)
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

func getGroup(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	descriptor := d.EqualsQuals["descriptor"].GetStringValue()
	if descriptor == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroup", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroup", "client_error", err)
		return nil, err
	}
	if err == nil {
		output, getErr := client.GetGroup(ctx, graph.GetGroupArgs{GroupDescriptor: &descriptor})
		if getErr == nil && output != nil {
			return groupListItem{GraphGroup: *output}, nil
		}
		if getErr != nil && !apiNotRegistered(getErr) {
			if absentFromCollection(d, getErr) {
				return nil, nil
			}
			plugin.Logger(ctx).Error("azuredevopssh_group.getGroup", "api_error", getErr)
			return nil, getErr
		}
	}

	output, routeErr := getGraphGroupByRoute(ctx, connection, descriptor)
	if routeErr == nil && output != nil {
		return groupListItem{GraphGroup: *output}, nil
	}
	if routeErr != nil && !graphRouteMissing(routeErr) {
		if absentFromCollection(d, routeErr) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroup", "api_error", routeErr)
		return nil, routeErr
	}

	item, idErr := identityGroupForDescriptor(ctx, connection, descriptor)
	if idErr != nil {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroup", "api_error", idErr)
		return nil, idErr
	}
	if item == nil {
		return nil, nil
	}
	row, ok := graphGroupFromIdentity(*item)
	if !ok {
		return nil, nil
	}
	return row, nil
}

func getGroupMembershipState(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	group := groupFromHydrate(h)
	if group.MembershipActive != nil {
		return group.MembershipActive, nil
	}
	descriptor := groupDescriptor(group)
	if descriptor == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroupMembershipState", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroupMembershipState", "client_error", err)
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
			plugin.Logger(ctx).Error("azuredevopssh_group.getGroupMembershipState", "api_error", stateErr)
			return nil, stateErr
		}
	}

	item, idErr := identityForSubject(ctx, connection, descriptor)
	if idErr != nil {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroupMembershipState", "api_error", idErr)
		return nil, idErr
	}
	if item == nil {
		return nil, nil
	}
	return item.IsActive, nil
}

func getGroupMemberships(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	group := groupFromHydrate(h)
	if group.Memberships != nil {
		return group.Memberships, nil
	}
	descriptor := groupDescriptor(group)
	if descriptor == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroupMemberships", "connection_error", err)
		return nil, err
	}

	client, err := graph.NewClient(ctx, connection)
	if err != nil && !apiNotRegistered(err) {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroupMemberships", "client_error", err)
		return nil, err
	}
	if err == nil {
		output, listErr := client.ListMemberships(ctx, graph.ListMembershipsArgs{SubjectDescriptor: &descriptor})
		if listErr == nil {
			return output, nil
		}
		if !apiNotRegistered(listErr) {
			plugin.Logger(ctx).Error("azuredevopssh_group.getGroupMemberships", "api_error", listErr)
			return nil, listErr
		}
	}

	item, idErr := identityForMembership(ctx, connection, descriptor)
	if idErr != nil {
		plugin.Logger(ctx).Error("azuredevopssh_group.getGroupMemberships", "api_error", idErr)
		return nil, idErr
	}
	memberships := membershipsFromMemberOf(descriptor, nil)
	if item != nil {
		memberships = membershipsFromMemberOf(descriptor, item.MemberOf)
	}
	return &memberships, nil
}

package azuredevops

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// identityRow carries the search quals so Steampipe's post-filter keeps rows
// whose API payload does not include filter_value or search_filter.
type identityRow struct {
	identity.Identity
	SearchFilter string
	FilterValue  string
}

func tableAzureDevOpsIdentity(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_identity",
		Description: "Retrieve Azure DevOps users and groups from the Identity service.",
		List: &plugin.ListConfig{
			Hydrate: listIdentities,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "filter_value", Require: plugin.Optional},
				{Name: "search_filter", Require: plugin.Optional},
				{Name: "descriptor", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getIdentity,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "Identity identifier (storage key).",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "descriptor",
				Description: "Legacy identity descriptor.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "subject_descriptor",
				Description: "Graph subject descriptor for this identity, when the server provides one.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "provider_display_name",
				Description: "Display name from the source identity provider.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "custom_display_name",
				Description: "Custom display name, when one is set.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "is_active",
				Description: "True when the identity is active in the collection.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_container",
				Description: "True when the identity is a group.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "search_filter",
				Description: "Identity search mode used for this row: AccountName, DisplayName, MailAddress, General, or LocalGroupName.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "filter_value",
				Description: "Search text used for this row. The Identity API rejects an empty value and '*'.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "member_ids",
				Description: "Storage keys of direct members. Populated for groups.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("MemberIds"),
			},
			{
				Name:        "members",
				Description: "Descriptors of direct members. Populated for groups.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Members"),
			},
			{
				Name:        "member_of",
				Description: "Descriptors of groups this identity belongs to.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("MemberOf"),
			},
			{
				Name:        "properties",
				Description: "Identity properties returned by the server.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ProviderDisplayName"),
			},
		}),
	}
}

func listIdentities(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	args, rowQuals, err := identityListArgs(d)
	if err != nil {
		return nil, err
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_identity.listIdentities", "connection_error", err)
		return nil, err
	}

	items, err := readIdentities(ctx, connection, args)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_identity.listIdentities", "api_error", err)
		return nil, err
	}
	if items == nil {
		return nil, nil
	}

	for _, item := range *items {
		d.StreamListItem(ctx, identityRow{
			Identity:     item,
			SearchFilter: rowQuals.searchFilter,
			FilterValue:  rowQuals.filterValue,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getIdentity(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id := d.EqualsQuals["id"].GetStringValue()
	if id == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_identity.getIdentity", "connection_error", err)
		return nil, err
	}

	direct := identity.QueryMembershipValues.Direct
	items, err := readIdentities(ctx, connection, identity.ReadIdentitiesArgs{
		IdentityIds:     &id,
		QueryMembership: &direct,
	})
	if err != nil {
		if absentFromCollection(d, err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_identity.getIdentity", "api_error", err)
		return nil, err
	}
	if items == nil || len(*items) == 0 {
		return nil, nil
	}
	return identityRow{Identity: (*items)[0]}, nil
}

type identityListQuals struct {
	searchFilter string
	filterValue  string
}

func identityListArgs(d *plugin.QueryData) (identity.ReadIdentitiesArgs, identityListQuals, error) {
	filterValue := qualString(d, "filter_value")
	searchFilter := qualString(d, "search_filter")
	descriptor := qualString(d, "descriptor")
	id := qualString(d, "id")

	if filterValue == "" && descriptor == "" && id == "" {
		return identity.ReadIdentitiesArgs{}, identityListQuals{}, fmt.Errorf("azuredevopssh_identity requires filter_value, descriptor, or id")
	}
	if searchFilter != "" && filterValue == "" {
		return identity.ReadIdentitiesArgs{}, identityListQuals{}, fmt.Errorf("azuredevopssh_identity search_filter requires filter_value")
	}

	direct := identity.QueryMembershipValues.Direct
	args := identity.ReadIdentitiesArgs{QueryMembership: &direct}
	quals := identityListQuals{searchFilter: searchFilter, filterValue: filterValue}

	switch {
	case id != "":
		args.IdentityIds = &id
	case descriptor != "":
		args.Descriptors = &descriptor
	default:
		if searchFilter == "" {
			searchFilter = "General"
			quals.searchFilter = searchFilter
		}
		args.SearchFilter = &searchFilter
		args.FilterValue = &filterValue
	}
	return args, quals, nil
}

package azuredevops

import (
	"context"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/memberentitlementmanagement"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// memberEntitlementRow keeps the optional filter qual on the row so a filtered
// query is not dropped by Steampipe after the API call.
type memberEntitlementRow struct {
	memberentitlementmanagement.MemberEntitlement2
	Filter string
}

func tableAzureDevOpsMemberEntitlement(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_member_entitlement",
		Description: "Retrieve member entitlements and licenses for an Azure DevOps organization.",
		List: &plugin.ListConfig{
			Hydrate: listMemberEntitlements,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Optional},
				{Name: "filter", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getMemberEntitlement,
		},
		Columns: commonColumns(append(entitlementColumns(),
			&plugin.Column{
				Name:        "display_name",
				Description: "Display name of the member.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Member.DisplayName"),
			},
			&plugin.Column{
				Name:        "descriptor",
				Description: "Graph descriptor of the member.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Member.Descriptor"),
			},
			&plugin.Column{
				Name:        "principal_name",
				Description: "Principal name of the member.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Member.PrincipalName"),
			},
			&plugin.Column{
				Name:        "mail_address",
				Description: "Mail address of the member.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Member.MailAddress"),
			},
			&plugin.Column{
				Name:        "member",
				Description: "Member reference returned by the entitlement API.",
				Type:        proto.ColumnType_JSON,
			},
			&plugin.Column{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Member.DisplayName"),
			},
		)),
	}
}

func entitlementColumns() []*plugin.Column {
	return []*plugin.Column{
		{
			Name:        "id",
			Description: "Identity id of the member.",
			Type:        proto.ColumnType_STRING,
		},
		{
			Name:        "account_license_type",
			Description: "Account license type, such as Express or Stakeholder.",
			Type:        proto.ColumnType_STRING,
			Transform:   transform.FromField("AccessLevel.AccountLicenseType"),
		},
		{
			Name:        "license_display_name",
			Description: "Display name of the assigned license.",
			Type:        proto.ColumnType_STRING,
			Transform:   transform.FromField("AccessLevel.LicenseDisplayName"),
		},
		{
			Name:        "licensing_source",
			Description: "Source of the license, such as account or MSDN.",
			Type:        proto.ColumnType_STRING,
			Transform:   transform.FromField("AccessLevel.LicensingSource"),
		},
		{
			Name:        "assignment_source",
			Description: "How the license was assigned.",
			Type:        proto.ColumnType_STRING,
			Transform:   transform.FromField("AccessLevel.AssignmentSource"),
		},
		{
			Name:        "license_status",
			Description: "License status for the member.",
			Type:        proto.ColumnType_STRING,
			Transform:   transform.FromField("AccessLevel.Status"),
		},
		{
			Name:        "status_message",
			Description: "License status message.",
			Type:        proto.ColumnType_STRING,
			Transform:   transform.FromField("AccessLevel.StatusMessage"),
		},
		{
			Name:        "date_created",
			Description: "When the member was added to the collection.",
			Type:        proto.ColumnType_TIMESTAMP,
			Transform:   transform.FromField("DateCreated.Time"),
		},
		{
			Name:        "last_accessed_date",
			Description: "When the member last accessed the collection.",
			Type:        proto.ColumnType_TIMESTAMP,
			Transform:   transform.FromField("LastAccessedDate.Time"),
		},
		{
			Name:        "filter",
			Description: "Optional entitlement search filter passed to the API, such as name eq 'ada'.",
			Type:        proto.ColumnType_STRING,
		},
		{
			Name:        "group_assignments",
			Description: "Group entitlements this member belongs to.",
			Type:        proto.ColumnType_JSON,
		},
		{
			Name:        "project_entitlements",
			Description: "Project permissions for this member.",
			Type:        proto.ColumnType_JSON,
		},
	}
}

func listMemberEntitlements(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	filter := qualString(d, "filter")
	id := qualString(d, "id")
	items, err := searchMemberEntitlements(ctx, d, "azuredevops_member_entitlement.listMemberEntitlements", filter)
	if err != nil || items == nil {
		return nil, err
	}

	for _, item := range *items {
		if id != "" && (item.Id == nil || item.Id.String() != id) {
			continue
		}
		d.StreamListItem(ctx, memberEntitlementRow{MemberEntitlement2: item, Filter: filter})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getMemberEntitlement(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	id := d.EqualsQuals["id"].GetStringValue()
	if _, err := uuid.Parse(id); err != nil {
		return nil, nil
	}
	items, err := searchMemberEntitlements(ctx, d, "azuredevops_member_entitlement.getMemberEntitlement", "")
	if err != nil || items == nil {
		return nil, err
	}
	for _, item := range *items {
		if item.Id != nil && item.Id.String() == id {
			return memberEntitlementRow{MemberEntitlement2: item}, nil
		}
	}
	return nil, nil
}

func searchMemberEntitlements(ctx context.Context, d *plugin.QueryData, hydrate, filter string) (*[]memberentitlementmanagement.MemberEntitlement2, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "connection_error", err)
		return nil, err
	}
	client, err := memberentitlementmanagement.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "client_error", err)
		return nil, err
	}

	selectAll := memberentitlementmanagement.UserEntitlementPropertyValues.All
	args := memberentitlementmanagement.SearchMemberEntitlementsArgs{Select: &selectAll}
	if filter != "" {
		args.Filter = &filter
	}
	items, err := client.SearchMemberEntitlements(ctx, args)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "api_error", err)
		return nil, err
	}
	return items, nil
}

package azuredevops

import (
	"context"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/memberentitlementmanagement"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// userEntitlementRow keeps the optional filter qual on the row so a filtered
// query is not dropped by Steampipe after the API call.
type userEntitlementRow struct {
	memberentitlementmanagement.UserEntitlement
	Filter string
}

func tableAzureDevOpsUserEntitlement(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_user_entitlement",
		Description: "Retrieve user entitlements and licenses for an Azure DevOps organization.",
		List: &plugin.ListConfig{
			Hydrate: listUserEntitlements,
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "id", Require: plugin.Optional},
				{Name: "filter", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getUserEntitlement,
		},
		Columns: commonColumns(append(entitlementColumns(),
			&plugin.Column{
				Name:        "display_name",
				Description: "Display name of the user.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("User.DisplayName"),
			},
			&plugin.Column{
				Name:        "descriptor",
				Description: "Graph descriptor of the user.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("User.Descriptor"),
			},
			&plugin.Column{
				Name:        "principal_name",
				Description: "Principal name of the user.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("User.PrincipalName"),
			},
			&plugin.Column{
				Name:        "mail_address",
				Description: "Mail address of the user.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("User.MailAddress"),
			},
			&plugin.Column{
				Name:        "user",
				Description: "User reference returned by the entitlement API.",
				Type:        proto.ColumnType_JSON,
			},
			&plugin.Column{
				Name:        "extensions",
				Description: "Extensions assigned to the user.",
				Type:        proto.ColumnType_JSON,
			},
			&plugin.Column{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("User.DisplayName"),
			},
		)),
	}
}

func listUserEntitlements(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_user_entitlement.listUserEntitlements", "connection_error", err)
		return nil, err
	}
	client, err := memberentitlementmanagement.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_user_entitlement.listUserEntitlements", "client_error", err)
		return nil, err
	}

	selectAll := memberentitlementmanagement.UserEntitlementPropertyValues.All
	args := memberentitlementmanagement.SearchUserEntitlementsArgs{Select: &selectAll}
	filter := qualString(d, "filter")
	if filter != "" {
		args.Filter = &filter
	}
	id := qualString(d, "id")

	page, err := client.SearchUserEntitlements(ctx, args)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_user_entitlement.listUserEntitlements", "api_error", err)
		return nil, err
	}
	if page == nil || page.Members == nil {
		return nil, nil
	}

	for _, item := range *page.Members {
		if id != "" && (item.Id == nil || item.Id.String() != id) {
			continue
		}
		d.StreamListItem(ctx, userEntitlementRow{UserEntitlement: item, Filter: filter})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getUserEntitlement(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	rawID := d.EqualsQuals["id"].GetStringValue()
	userID, err := uuid.Parse(rawID)
	if err != nil {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_user_entitlement.getUserEntitlement", "connection_error", err)
		return nil, err
	}
	client, err := memberentitlementmanagement.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_user_entitlement.getUserEntitlement", "client_error", err)
		return nil, err
	}

	item, err := client.GetUserEntitlement(ctx, memberentitlementmanagement.GetUserEntitlementArgs{UserId: &userID})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_user_entitlement.getUserEntitlement", "api_error", err)
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	return userEntitlementRow{UserEntitlement: *item}, nil
}

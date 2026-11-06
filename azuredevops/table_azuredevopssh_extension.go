package azuredevops

import (
	"context"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/extensionmanagement"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsExtension(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_extension",
		Description: "Retrieve extensions installed in the Azure DevOps collection.",
		List: &plugin.ListConfig{
			Hydrate: listExtensions,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.KeyColumnSlice{
				{Name: "publisher_id", Require: plugin.Required},
				{Name: "extension_id", Require: plugin.Required},
			},
			Hydrate: getExtension,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "extension_id",
				Description: "Extension identifier, unique for its publisher.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "extension_name",
				Description: "Display name of the extension.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "publisher_id",
				Description: "Publisher identifier.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "publisher_name",
				Description: "Publisher display name.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "version",
				Description: "Installed version.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "registration_id",
				Description: "Registration identifier shared by every version of the extension.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "flags",
				Description: "Extension flags.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "install_state_flags",
				Description: "Installation state flags.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("InstallState.Flags"),
			},
			{
				Name:        "last_published",
				Description: "When the extension was last published to the gallery.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("LastPublished.Time"),
			},
			{
				Name:        "scopes",
				Description: "OAuth scopes required by the extension.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ExtensionName"),
			},
		}),
	}
}

func listExtensions(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := newExtensionClient(ctx, d, "azuredevopssh_extension.listExtensions")
	if err != nil {
		return nil, err
	}
	extensions, err := client.GetInstalledExtensions(ctx, extensionmanagement.GetInstalledExtensionsArgs{
		IncludeDisabledExtensions: boolPtr(true),
		IncludeErrors:             boolPtr(true),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_extension.listExtensions", "api_error", err)
		return nil, err
	}
	if extensions == nil {
		return nil, nil
	}
	for _, item := range *extensions {
		d.StreamListItem(ctx, item)
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}
	return nil, nil
}

func getExtension(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	publisherID := qualString(d, "publisher_id")
	extensionID := qualString(d, "extension_id")
	if publisherID == "" || extensionID == "" {
		return nil, nil
	}
	client, err := newExtensionClient(ctx, d, "azuredevopssh_extension.getExtension")
	if err != nil {
		return nil, err
	}
	item, err := client.GetInstalledExtensionByName(ctx, extensionmanagement.GetInstalledExtensionByNameArgs{
		PublisherName: &publisherID,
		ExtensionName: &extensionID,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_extension.getExtension", "api_error", err)
		return nil, err
	}
	return item, nil
}

func newExtensionClient(ctx context.Context, d *plugin.QueryData, hydrate string) (extensionmanagement.Client, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "connection_error", err)
		return nil, err
	}
	client, err := extensionmanagement.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error(hydrate, "client_error", err)
		return nil, err
	}
	return client, nil
}

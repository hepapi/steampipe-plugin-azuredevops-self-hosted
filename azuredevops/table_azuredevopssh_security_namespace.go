package azuredevops

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsSecurityNamespace(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_security_namespace",
		Description: "Retrieve security namespaces and the permission bits they define.",
		List: &plugin.ListConfig{
			Hydrate: listSecurityNamespaces,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "namespace_id", Require: plugin.Optional},
				{Name: "name", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("namespace_id"),
			Hydrate:    getSecurityNamespace,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "namespace_id",
				Description: "The unique identifier for this security namespace.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "name",
				Description: "The non-localized name of this namespace.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "display_name",
				Description: "The localized display name of this namespace.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "dataspace_category",
				Description: "Dataspace category that describes where security information for this namespace is stored.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "separator_value",
				Description: "Character used to split security tokens into elements. Empty when the namespace is not hierarchical.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "element_length",
				Description: "Fixed element length used to split security tokens, or -1 when tokens are not fixed-length.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "structure_value",
				Description: "Integer describing the structure of the security namespace.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "is_remotable",
				Description: "True if another service can proxy this security namespace.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "use_token_translator",
				Description: "True if the security service expects a token translator plugin for this namespace.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "extension_type",
				Description: "Extension type loaded from the plugins directory to extend this namespace.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "read_permission",
				Description: "Permission bits required to read security data in this namespace.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "write_permission",
				Description: "Permission bits required to modify security data in this namespace.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "system_bit_mask",
				Description: "Bits reserved by the system store.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "actions",
				Description: "Actions this namespace secures. Each action has a name, display name, and a power-of-two bit.",
				Type:        proto.ColumnType_JSON,
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

func listSecurityNamespaces(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_security_namespace.listSecurityNamespaces", "connection_error", err)
		return nil, err
	}
	client := security.NewClient(ctx, connection)

	input := security.QuerySecurityNamespacesArgs{}
	if id := qualString(d, "namespace_id"); id != "" {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid security namespace id %q: %w", id, err)
		}
		input.SecurityNamespaceId = &parsed
	}

	namespaces, err := client.QuerySecurityNamespaces(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_security_namespace.listSecurityNamespaces", "api_error", err)
		return nil, err
	}
	if namespaces == nil {
		return nil, nil
	}

	nameQual := qualString(d, "name")
	idQual := qualString(d, "namespace_id")
	for _, namespace := range *namespaces {
		if idQual != "" && (namespace.NamespaceId == nil || !strings.EqualFold(namespace.NamespaceId.String(), idQual)) {
			continue
		}
		if nameQual != "" && (namespace.Name == nil || !strings.EqualFold(*namespace.Name, nameQual)) {
			continue
		}
		d.StreamListItem(ctx, namespace)
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

func getSecurityNamespace(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if h.Item != nil {
		if namespace, ok := h.Item.(security.SecurityNamespaceDescription); ok {
			return namespace, nil
		}
	}

	id := qualString(d, "namespace_id")
	if id == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid security namespace id %q: %w", id, err)
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_security_namespace.getSecurityNamespace", "connection_error", err)
		return nil, err
	}
	client := security.NewClient(ctx, connection)

	namespaces, err := client.QuerySecurityNamespaces(ctx, security.QuerySecurityNamespacesArgs{
		SecurityNamespaceId: &parsed,
	})
	if err != nil {
		if absentFromCollection(d, err) {
			return nil, nil
		}
		plugin.Logger(ctx).Error("azuredevopssh_security_namespace.getSecurityNamespace", "namespace_id", id, "api_error", err)
		return nil, err
	}
	if namespaces == nil {
		return nil, nil
	}
	nameQual := qualString(d, "name")
	for _, namespace := range *namespaces {
		if namespace.NamespaceId != nil && strings.EqualFold(namespace.NamespaceId.String(), id) {
			if nameQual != "" && (namespace.Name == nil || !strings.EqualFold(*namespace.Name, nameQual)) {
				return nil, nil
			}
			return namespace, nil
		}
	}
	if len(*namespaces) == 1 {
		namespace := (*namespaces)[0]
		if nameQual != "" && (namespace.Name == nil || !strings.EqualFold(*namespace.Name, nameQual)) {
			return nil, nil
		}
		return namespace, nil
	}
	return nil, nil
}

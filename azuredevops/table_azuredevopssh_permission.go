package azuredevops

import (
	"context"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsPermission(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_permission",
		Description: "Retrieve access control entries (ACEs) from Azure DevOps security namespaces.",
		List: &plugin.ListConfig{
			ParentHydrate: listSecurityNamespaces,
			Hydrate:       listPermissions,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "namespace_id", Require: plugin.Optional},
				{Name: "token", Require: plugin.Optional},
				{Name: "descriptor", Require: plugin.Optional},
			},
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "namespace_id",
				Description: "Security namespace this access control entry belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "namespace_name",
				Description: "Non-localized name of the security namespace.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "namespace_display_name",
				Description: "Localized display name of the security namespace.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "token",
				Description: "Security token this entry applies to. Token format depends on the namespace.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "descriptor",
				Description: "Identity descriptor this entry grants or denies permissions for.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "inherit_permissions",
				Description: "True if this token inherits permissions from its parent token.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "allow",
				Description: "Explicit allow permission bits for this identity on this token.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "deny",
				Description: "Explicit deny permission bits for this identity on this token.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "allow_actions",
				Description: "Namespace action names whose bits are set in allow.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "deny_actions",
				Description: "Namespace action names whose bits are set in deny.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "extended_info",
				Description: "Inherited and effective allow/deny bits for this identity on this token.",
				Type:        proto.ColumnType_JSON,
			},

			/// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Descriptor"),
			},
		}),
	}
}

// Permission is one access control entry, flattened so each identity on a token is its own row.
type Permission struct {
	NamespaceId          string
	NamespaceName        string
	NamespaceDisplayName string
	Token                string
	Descriptor           string
	InheritPermissions   *bool
	Allow                *int
	Deny                 *int
	AllowActions         []string
	DenyActions          []string
	ExtendedInfo         *security.AceExtendedInformation
}

func listPermissions(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	namespace, ok := h.Item.(security.SecurityNamespaceDescription)
	if !ok || namespace.NamespaceId == nil {
		return nil, nil
	}
	namespaceID := namespace.NamespaceId.String()
	if qual := qualString(d, "namespace_id"); qual != "" && !strings.EqualFold(qual, namespaceID) {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_permission.listPermissions", "connection_error", err)
		return nil, err
	}
	client := security.NewClient(ctx, connection)

	input := security.QueryAccessControlListsArgs{
		SecurityNamespaceId: namespace.NamespaceId,
		IncludeExtendedInfo: types.Bool(true),
	}
	// A token qual selects that token. Recurse is left unset so child tokens
	// are not mixed into a query that asked for one token.
	if token := qualString(d, "token"); token != "" {
		input.Token = types.String(token)
	}
	if descriptor := qualString(d, "descriptor"); descriptor != "" {
		input.Descriptors = types.String(descriptor)
	}

	acls, err := client.QueryAccessControlLists(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_permission.listPermissions", "namespace_id", namespaceID, "api_error", err)
		// A query scoped to one namespace should surface the API error.
		// An unscoped inventory skips namespaces the caller cannot read.
		if qualString(d, "namespace_id") != "" {
			return nil, err
		}
		return nil, nil
	}
	if acls == nil {
		return nil, nil
	}

	namespaceName := ""
	if namespace.Name != nil {
		namespaceName = *namespace.Name
	}
	namespaceDisplayName := ""
	if namespace.DisplayName != nil {
		namespaceDisplayName = *namespace.DisplayName
	}
	descriptorQual := qualString(d, "descriptor")

	for _, acl := range *acls {
		if acl.AcesDictionary == nil {
			continue
		}
		token := ""
		if acl.Token != nil {
			token = *acl.Token
		}
		for key, ace := range *acl.AcesDictionary {
			descriptor := key
			if ace.Descriptor != nil && *ace.Descriptor != "" {
				descriptor = *ace.Descriptor
			}
			if descriptorQual != "" && !strings.EqualFold(descriptor, descriptorQual) {
				continue
			}
			d.StreamListItem(ctx, Permission{
				NamespaceId:          namespaceID,
				NamespaceName:        namespaceName,
				NamespaceDisplayName: namespaceDisplayName,
				Token:                token,
				Descriptor:           descriptor,
				InheritPermissions:   acl.InheritPermissions,
				Allow:                ace.Allow,
				Deny:                 ace.Deny,
				AllowActions:         permissionActionNames(namespace.Actions, ace.Allow),
				DenyActions:          permissionActionNames(namespace.Actions, ace.Deny),
				ExtendedInfo:         ace.ExtendedInfo,
			})
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}

	return nil, nil
}

// permissionActionNames returns the namespace action names covered by bits.
// Action bits are powers of two, so a bit is included when it is fully set.
func permissionActionNames(actions *[]security.ActionDefinition, bits *int) []string {
	if actions == nil || bits == nil || *bits == 0 {
		return nil
	}
	names := make([]string, 0)
	for _, action := range *actions {
		if action.Bit == nil || action.Name == nil || *action.Bit == 0 {
			continue
		}
		if *bits&*action.Bit == *action.Bit {
			names = append(names, *action.Name)
		}
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

package azuredevops

import (
	"context"
	"encoding/json"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsGitRepositoryBranchPolicy(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevopssh_git_repository_branch_policy",
		Description: "Retrieve information about your Git repository branch policies.",
		List: &plugin.ListConfig{
			ParentHydrate: listProjects,
			Hydrate:       listGitRepositoryBranchPolicies,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "policy_type_id", Require: plugin.Optional},
				{Name: "is_enabled", Require: plugin.Optional},
				{Name: "is_blocking", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "project_id"}),
			Hydrate:    getGitRepositoryBranchPolicy,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The policy configuration ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this policy belongs to.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "revision",
				Description: "The policy configuration revision ID.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "is_enabled",
				Description: "Indicates whether the policy is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_blocking",
				Description: "Indicates whether the policy is blocking.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_deleted",
				Description: "Indicates whether the policy has been (soft) deleted.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "is_enterprise_managed",
				Description: "If set, this policy requires 'Manage Enterprise Policies' permission to create, edit, or delete.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "created_date",
				Description: "The date and time when the policy was created.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CreatedDate.Time"),
			},
			{
				Name:        "created_by",
				Description: "A reference to the identity that created the policy.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "policy_type_id",
				Description: "The policy type ID.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type.Id"),
			},
			{
				Name:        "policy_type_display_name",
				Description: "The display name of the policy type.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type.DisplayName"),
			},
			{
				Name:        "url",
				Description: "The URL where the policy configuration can be retrieved.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "links",
				Description: "The links to other objects related to this object.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "settings",
				Description: "The policy configuration settings.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "scope",
				Description: "The scope entries (branch/repository) this policy applies to.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getBranchPolicyScope,
				Transform:   transform.FromValue(),
			},

			/// Steampipe standard columns
			{
				Name:        "title",
				Description: "Title of the resource.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Type.DisplayName"),
			},
		}),
	}
}

// BranchPolicy wraps a policy.PolicyConfiguration with its project ID.
type BranchPolicy struct {
	policy.PolicyConfiguration
	ProjectId string
}

// policyScopeEntry represents a single scope entry within a policy's settings.
type policyScopeEntry struct {
	RefName      string  `json:"refName"`
	MatchKind    string  `json:"matchKind"`
	RepositoryId *string `json:"repositoryId"`
}

// policySettings is the typed view of a policy's settings JSON.
type policySettings struct {
	Scope []policyScopeEntry `json:"scope"`
}

func listGitRepositoryBranchPolicies(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	project := h.Item.(core.TeamProjectReference)
	projectId := d.EqualsQuals["project_id"].GetStringValue()

	// check if the provided project_id is not matching with the parentHydrate
	if projectId != "" && projectId != project.Id.String() {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository_branch_policy.listGitRepositoryBranchPolicies", "connection_error", err)
		return nil, err
	}

	client, err := policy.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository_branch_policy.listGitRepositoryBranchPolicies", "client_error", err)
		return nil, err
	}

	input := policy.GetPolicyConfigurationsArgs{
		Project: types.String(project.Id.String()),
	}

	response, err := client.GetPolicyConfigurations(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository_branch_policy.listGitRepositoryBranchPolicies", "api_error", err)
		return nil, err
	}

	for _, configuration := range response.Value {
		d.StreamListItem(ctx, BranchPolicy{
			PolicyConfiguration: configuration,
			ProjectId:           project.Id.String(),
		})

		// Context can be cancelled due to manual cancellation or the limit has been hit
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

func getGitRepositoryBranchPolicy(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	policyId := d.EqualsQuals["id"].GetInt64Value()
	projectId := d.EqualsQuals["project_id"].GetStringValue()

	// Check if projectId is empty
	if projectId == "" {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository_branch_policy.getGitRepositoryBranchPolicy", "connection_error", err)
		return nil, err
	}

	client, err := policy.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository_branch_policy.getGitRepositoryBranchPolicy", "client_error", err)
		return nil, err
	}

	input := policy.GetPolicyConfigurationArgs{
		Project:         types.String(projectId),
		ConfigurationId: types.Int(int(policyId)),
	}

	configuration, err := client.GetPolicyConfiguration(ctx, input)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevopssh_git_repository_branch_policy.getGitRepositoryBranchPolicy", "api_error", err)
		return nil, err
	}

	return BranchPolicy{PolicyConfiguration: *configuration, ProjectId: projectId}, nil
}

// getBranchPolicyScope extracts the raw scope array (JSON) from the policy settings.
func getBranchPolicyScope(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	item := h.Item.(BranchPolicy)
	if item.Settings == nil {
		return nil, nil
	}

	raw, err := json.Marshal(item.Settings)
	if err != nil {
		return nil, nil
	}

	var parsed policySettings
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, nil
	}

	return parsed.Scope, nil
}

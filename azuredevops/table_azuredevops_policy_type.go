package azuredevops

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAzureDevOpsPolicyType(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azuredevops_policy_type",
		Description: "Retrieve information about policy types available in a project.",
		List: &plugin.ListConfig{
			ParentHydrate: listProjects,
			Hydrate:       listPolicyTypes,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "project_id", Require: plugin.Optional},
				{Name: "id", Require: plugin.Optional},
			},
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"id", "project_id"}),
			Hydrate:    getPolicyType,
		},
		Columns: commonColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The policy type ID.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "project_id",
				Description: "ID of the project this policy type was retrieved from.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "display_name",
				Description: "Display name of the policy type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "description",
				Description: "Detailed description of the policy type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "url",
				Description: "The URL where the policy type can be retrieved.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "links",
				Description: "The links to other objects related to this object.",
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

// PolicyTypeRow is a policy type plus the project it was listed from.
// Policy types are project-scoped in the REST API even when the same id
// exists in every project.
type PolicyTypeRow struct {
	policy.PolicyType
	ProjectId string
}

func listPolicyTypes(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	project, ok := h.Item.(core.TeamProjectReference)
	if !ok || project.Id == nil {
		return nil, nil
	}
	projectID := project.Id.String()
	if qual := qualString(d, "project_id"); qual != "" && !strings.EqualFold(qual, projectID) {
		return nil, nil
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_type.listPolicyTypes", "connection_error", err)
		return nil, err
	}
	client, err := policy.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_type.listPolicyTypes", "client_error", err)
		return nil, err
	}

	typesList, err := client.GetPolicyTypes(ctx, policy.GetPolicyTypesArgs{
		Project: types.String(projectID),
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_type.listPolicyTypes", "project_id", projectID, "api_error", err)
		return nil, err
	}
	if typesList == nil {
		return nil, nil
	}

	idQual := qualString(d, "id")
	for _, policyType := range *typesList {
		if idQual != "" && (policyType.Id == nil || !strings.EqualFold(policyType.Id.String(), idQual)) {
			continue
		}
		d.StreamListItem(ctx, PolicyTypeRow{
			PolicyType: policyType,
			ProjectId:  projectID,
		})
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	return nil, nil
}

func getPolicyType(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	if h.Item != nil {
		if row, ok := h.Item.(PolicyTypeRow); ok {
			return row, nil
		}
	}

	projectID := qualString(d, "project_id")
	id := qualString(d, "id")
	if projectID == "" || id == "" {
		return nil, nil
	}
	typeID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid policy type id %q: %w", id, err)
	}

	connection, err := getConnection(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_type.getPolicyType", "connection_error", err)
		return nil, err
	}
	client, err := policy.NewClient(ctx, connection)
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_type.getPolicyType", "client_error", err)
		return nil, err
	}

	policyType, err := client.GetPolicyType(ctx, policy.GetPolicyTypeArgs{
		Project: types.String(projectID),
		TypeId:  &typeID,
	})
	if err != nil {
		plugin.Logger(ctx).Error("azuredevops_policy_type.getPolicyType", "project_id", projectID, "id", id, "api_error", err)
		return nil, err
	}
	if policyType == nil {
		return nil, nil
	}

	return PolicyTypeRow{PolicyType: *policyType, ProjectId: projectID}, nil
}

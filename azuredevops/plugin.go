package azuredevops

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

// Plugin creates this (azuredevops) plugin
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             "steampipe-plugin-azuredevopssh",
		DefaultTransform: transform.FromCamel(),
		ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
			NewInstance: ConfigInstance,
		},
		ConnectionKeyColumns: []plugin.ConnectionKeyColumn{
			{
				Name:    "organization",
				Hydrate: getOrganization,
			},
		},
		TableMap: map[string]*plugin.Table{
			"azuredevops_build":                        tableAzureDevOpsBuild(ctx),
			"azuredevops_build_definition":             tableAzureDevOpsBuildDefinition(ctx),
			"azuredevops_dashboard":                    tableAzureDevOpsDashboard(ctx),
			"azuredevops_git_commit":                   tableAzureDevOpsGitCommit(ctx),
			"azuredevops_git_pull_request":             tableAzureDevOpsGitPullRequest(ctx),
			"azuredevops_git_pull_request_comment":     tableAzureDevOpsGitPullRequestComment(ctx),
			"azuredevops_git_pull_request_reviewer":    tableAzureDevOpsGitPullRequestReviewer(ctx),
			"azuredevops_git_pull_request_thread":      tableAzureDevOpsGitPullRequestThread(ctx),
			"azuredevops_git_ref":                      tableAzureDevOpsGitRef(ctx),
			"azuredevops_git_repository":               tableAzureDevOpsGitRepository(ctx),
			"azuredevops_git_repository_branch":        tableAzureDevOpsGitRepositoryBranch(ctx),
			"azuredevops_git_repository_branch_policy": tableAzureDevOpsGitRepositoryBranchPolicy(ctx),
			"azuredevops_git_tag":                      tableAzureDevOpsGitTag(ctx),
			"azuredevops_group":                        tableAzureDevOpsGroup(ctx),
			"azuredevops_permission":                   tableAzureDevOpsPermission(ctx),
			"azuredevops_pipeline":                     tableAzureDevOpsPipeline(ctx),
			"azuredevops_policy_evaluation":            tableAzureDevOpsPolicyEvaluation(ctx),
			"azuredevops_policy_type":                  tableAzureDevOpsPolicyType(ctx),
			"azuredevops_project":                      tableAzureDevOpsProject(ctx),
			"azuredevops_project_collection":           tableAzureDevOpsProjectCollection(ctx),
			"azuredevops_release":                      tableAzureDevOpsRelease(ctx),
			"azuredevops_security_namespace":           tableAzureDevOpsSecurityNamespace(ctx),
			"azuredevops_serviceendpoint":              tableAzureDevOpsServiceEndpoint(ctx),
			"azuredevops_servicehook_subscription":     tableAzureDevOpsServiceHookSubscription(ctx),
			"azuredevops_team":                         tableAzureDevOpsTeam(ctx),
			"azuredevops_team_member":                  tableAzureDevOpsTeamMember(ctx),
			"azuredevops_user":                         tableAzureDevOpsUser(ctx),
		},
	}
	return p
}

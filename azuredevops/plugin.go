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
		// organization is not a connection key. One connection can list many
		// collections, so `where organization = ...` has to reach the collection
		// matrix. A connection key would drop the connection unless the value
		// matched the single name derived from organization_url.
		TableMap: map[string]*plugin.Table{
			"azuredevopssh_agent":                        tableAzureDevOpsAgent(ctx),
			"azuredevopssh_agent_pool":                   tableAzureDevOpsAgentPool(ctx),
			"azuredevopssh_audit_log":                    tableAzureDevOpsAuditLog(ctx),
			"azuredevopssh_build":                        tableAzureDevOpsBuild(ctx),
			"azuredevopssh_build_artifact":               tableAzureDevOpsBuildArtifact(ctx),
			"azuredevopssh_build_definition":             tableAzureDevOpsBuildDefinition(ctx),
			"azuredevopssh_build_log":                    tableAzureDevOpsBuildLog(ctx),
			"azuredevopssh_build_timeline":               tableAzureDevOpsBuildTimeline(ctx),
			"azuredevopssh_code_search_result":           tableAzureDevOpsCodeSearchResult(ctx),
			"azuredevopssh_dashboard":                    tableAzureDevOpsDashboard(ctx),
			"azuredevopssh_extension":                    tableAzureDevOpsExtension(ctx),
			"azuredevopssh_feed":                         tableAzureDevOpsFeed(ctx),
			"azuredevopssh_git_commit":                   tableAzureDevOpsGitCommit(ctx),
			"azuredevopssh_git_pull_request":             tableAzureDevOpsGitPullRequest(ctx),
			"azuredevopssh_git_pull_request_comment":     tableAzureDevOpsGitPullRequestComment(ctx),
			"azuredevopssh_git_pull_request_reviewer":    tableAzureDevOpsGitPullRequestReviewer(ctx),
			"azuredevopssh_git_pull_request_thread":      tableAzureDevOpsGitPullRequestThread(ctx),
			"azuredevopssh_git_ref":                      tableAzureDevOpsGitRef(ctx),
			"azuredevopssh_git_repository":               tableAzureDevOpsGitRepository(ctx),
			"azuredevopssh_git_repository_branch":        tableAzureDevOpsGitRepositoryBranch(ctx),
			"azuredevopssh_git_repository_branch_policy": tableAzureDevOpsGitRepositoryBranchPolicy(ctx),
			"azuredevopssh_git_tag":                      tableAzureDevOpsGitTag(ctx),
			"azuredevopssh_group":                        tableAzureDevOpsGroup(ctx),
			"azuredevopssh_group_membership":             tableAzureDevOpsGroupMembership(ctx),
			"azuredevopssh_identity":                     tableAzureDevOpsIdentity(ctx),
			"azuredevopssh_member_entitlement":           tableAzureDevOpsMemberEntitlement(ctx),
			"azuredevopssh_package":                      tableAzureDevOpsPackage(ctx),
			"azuredevopssh_package_version":              tableAzureDevOpsPackageVersion(ctx),
			"azuredevopssh_permission":                   tableAzureDevOpsPermission(ctx),
			"azuredevopssh_personal_access_token":        tableAzureDevOpsPersonalAccessToken(ctx),
			"azuredevopssh_pipeline":                     tableAzureDevOpsPipeline(ctx),
			"azuredevopssh_pipeline_approval":            tableAzureDevOpsPipelineApproval(ctx),
			"azuredevopssh_pipeline_check":               tableAzureDevOpsPipelineCheck(ctx),
			"azuredevopssh_pipeline_run":                 tableAzureDevOpsPipelineRun(ctx),
			"azuredevopssh_policy_evaluation":            tableAzureDevOpsPolicyEvaluation(ctx),
			"azuredevopssh_policy_type":                  tableAzureDevOpsPolicyType(ctx),
			"azuredevopssh_project":                      tableAzureDevOpsProject(ctx),
			"azuredevopssh_project_collection":           tableAzureDevOpsProjectCollection(ctx),
			"azuredevopssh_release":                      tableAzureDevOpsRelease(ctx),
			"azuredevopssh_release_approval":             tableAzureDevOpsReleaseApproval(ctx),
			"azuredevopssh_release_definition":           tableAzureDevOpsReleaseDefinition(ctx),
			"azuredevopssh_release_environment":          tableAzureDevOpsReleaseEnvironment(ctx),
			"azuredevopssh_security_namespace":           tableAzureDevOpsSecurityNamespace(ctx),
			"azuredevopssh_serviceendpoint":              tableAzureDevOpsServiceEndpoint(ctx),
			"azuredevopssh_servicehook_subscription":     tableAzureDevOpsServiceHookSubscription(ctx),
			"azuredevopssh_team":                         tableAzureDevOpsTeam(ctx),
			"azuredevopssh_team_member":                  tableAzureDevOpsTeamMember(ctx),
			"azuredevopssh_test_case":                    tableAzureDevOpsTestCase(ctx),
			"azuredevopssh_test_plan":                    tableAzureDevOpsTestPlan(ctx),
			"azuredevopssh_test_point":                   tableAzureDevOpsTestPoint(ctx),
			"azuredevopssh_test_result":                  tableAzureDevOpsTestResult(ctx),
			"azuredevopssh_test_run":                     tableAzureDevOpsTestRun(ctx),
			"azuredevopssh_test_suite":                   tableAzureDevOpsTestSuite(ctx),
			"azuredevopssh_user":                         tableAzureDevOpsUser(ctx),
			"azuredevopssh_user_entitlement":             tableAzureDevOpsUserEntitlement(ctx),
			"azuredevopssh_variable_group":               tableAzureDevOpsVariableGroup(ctx),
			"azuredevopssh_wiki":                         tableAzureDevOpsWiki(ctx),
			"azuredevopssh_wiki_page":                    tableAzureDevOpsWikiPage(ctx),
			"azuredevopssh_work_item":                    tableAzureDevOpsWorkItem(ctx),
		},
	}
	attachCollectionMatrix(p)
	return p
}

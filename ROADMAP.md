# Azure DevOps (Self-Hosted) Steampipe Plugin — Roadmap

This is the roadmap for Hepapi's `azuredevopssh` plugin. The target is **self-hosted Azure DevOps Server 2022.1**. Azure DevOps Services (`https://dev.azure.com/{organization}`) keeps working where the same API exists.

- **Client:** `github.com/microsoft/azure-devops-go-api/azuredevops/v7` (v7.1.0)
- **REST reference:** https://learn.microsoft.com/en-us/rest/api/azure/devops/?view=azure-devops-rest-7.2

Each stage is independently shippable and testable. Checkboxes are marked `[x]` when a table is implemented, documented, and verified against the live collection.

---

## Stage 0 — Foundation (self-hosted + v7 + branch policies)

- [x] Migrate from v6 to v7 go client
- [x] Fix `organization` column derivation for self-hosted URLs (collection name)
- [x] Add `server_url` config + `getServerURL`/`getServerConnection` helpers
- [x] `azuredevopssh_project_collection` — `core.GetProjectCollections` / `GetProjectCollection`
- [x] `azuredevopssh_git_repository_branch_policy` — `policy.GetPolicyConfigurations` / `GetPolicyConfiguration`

---

## Stage 1 — Git & Code (pairs with branch policies)

- [x] `azuredevopssh_git_pull_request` — `git.GetPullRequests`
- [x] `azuredevopssh_git_pull_request_thread` — `git.GetThreads`
- [x] `azuredevopssh_git_pull_request_comment` — `git.GetComments`
- [x] `azuredevopssh_git_pull_request_reviewer` — `git.GetPullRequestReviewers`
- [x] `azuredevopssh_git_commit` — `git.GetCommits`
- [x] `azuredevopssh_git_ref` — `git.GetRefs`
- [x] `azuredevopssh_git_tag` — `git.GetAnnotatedTags`
- [x] Enrich `azuredevopssh_git_repository` (project name, default branch object, `is_disabled`)

---

## Stage 2 — Policy & Security

- [x] `azuredevopssh_policy_type` — `policy.GetPolicyTypes`
- [x] `azuredevopssh_policy_evaluation` — `policy.GetPolicyEvaluations`
- [x] `azuredevopssh_security_namespace` — `security.QuerySecurityNamespaces`
- [x] `azuredevopssh_permission` / ACLs — `security.QueryAccessControlLists`
- [x] `azuredevopssh_servicehook_subscription` — `servicehooks.ListSubscriptions`

---

## Stage 3 — Build & Release (deepen existing)

- [x] `azuredevopssh_build_timeline` — `build.GetBuildTimeline` (requires `project_id` and `build_id`)
- [x] `azuredevopssh_build_log` — `build.GetBuildLogs` (metadata only; requires `project_id` and `build_id`)
- [x] `azuredevopssh_build_artifact` — `build.GetArtifacts` (requires `project_id` and `build_id`)
- [x] `azuredevopssh_release_definition` — `release.GetReleaseDefinitions`
- [x] `azuredevopssh_release_environment` — environments expanded on `release.GetReleases`
- [x] `azuredevopssh_release_approval` — `release.GetApprovals` (pending, approved, rejected, canceled, skipped, reassigned)
- [x] `azuredevopssh_pipeline_run` — `pipelines.ListRuns` (requires `project_id` and `pipeline_id`)
- [x] `azuredevopssh_pipeline_check` — `pipelineschecks.GetCheckConfigurationsOnResource` (requires `project_id`, `resource_type`, and `resource_id`)
- [x] `azuredevopssh_pipeline_approval` — `pipelinesapproval.QueryApprovals` (requires `project_id` and `id`; Server 2022 does not list approvals without an id)

---

## Stage 4 — Work Item Tracking

- [ ] `azuredevopssh_work_item` — `workitemtracking.GetWorkItems` / `GetWorkItemsBatch`
- [ ] `azuredevopssh_work_item_type` — `workitemtracking.GetWorkItemTypes`
- [ ] `azuredevopssh_work_item_comment` — `workitemtracking.GetComments`
- [ ] `azuredevopssh_work_item_revision` — `workitemtracking.GetWorkItemRevisions`
- [ ] `azuredevopssh_work_item_relation` — `workitemtracking.GetWorkItemRelations`
- [ ] `azuredevopssh_work_item_field` — `workitemtracking.GetFields`
- [ ] `azuredevopssh_work_item_query` — `workitemtracking.QueryByWiql`

---

## Stage 5 — Test & Quality

- [x] `azuredevopssh_test_plan` — `testplan`
- [x] `azuredevopssh_test_suite` — `testplan`
- [x] `azuredevopssh_test_case` — `testplan`
- [x] `azuredevopssh_test_run` — `test` (`GetTestRuns`; the `testresults` package is logs and metadata)
- [x] `azuredevopssh_test_result` — `test` (`GetTestResults`)
- [x] `azuredevopssh_test_point` — `testplan`

---

## Stage 6 — Identity, Permissions & Entitlements

- [ ] `azuredevopssh_member_entitlement` — `memberentitlementmanagement` (implemented; this Server 2022.1 collection does not register the API)
- [x] `azuredevopssh_group_membership` — `graph`, with an Identity `queryMembership=direct` fallback when the graph location is not registered
- [ ] `azuredevopssh_user_entitlement` — `memberentitlementmanagement` (implemented; this Server 2022.1 collection does not register the API)
- [x] `azuredevopssh_identity` — `identity` (`/_apis/identities` when the v7 location id is not registered)
- [x] `azuredevopssh_personal_access_token` — `tokenadmin.ListPersonalAccessTokens`. Server calls `/_apis/tokenadmin/personalaccesstokens/{subjectDescriptor}` with api-version 6.0-preview.1 when the v7 location is not registered. Lists name, scope, owner, and expiration for every user in Team Foundation Valid Users or Project Collection Valid Users. The secret is not returned. The calling token must be allowed to administer tokens. SSH keys are excluded.

---

## Stage 7 — Artifacts / Packages

- [x] `azuredevopssh_feed` — `feed`
- [x] `azuredevopssh_package` — `feed` (`GetPackages`; both collection feeds are empty)
- [x] `azuredevopssh_package_version` — `feed` (`GetPackageVersions`; not row-verified, the feeds have no packages)

---

## Stage 8 — Wiki, Search, Audit, Extensions, Agents

- [x] `azuredevopssh_wiki` / `azuredevopssh_wiki_page` — `wiki`
- [x] `azuredevopssh_code_search_result` — `search`
- [ ] `azuredevopssh_audit_log` — `audit` (implemented; this Server 2022.1 collection does not register the API)
- [x] `azuredevopssh_extension` — `extensionmanagement`
- [x] `azuredevopssh_agent_pool` / `azuredevopssh_agent` — `taskagent`
- [x] `azuredevopssh_variable_group` — `taskagent`

---

## Self-hosted caveats

Some packages may be unavailable or limited on Azure DevOps Server 2022.1 depending on edition/licensing. Validate each against the live collection before committing; mark unavailable ones here:

- `feed` / `packagingshared` (Azure Artifacts) — feeds are available on this collection. The two collection feeds have no packages, so package versions were not row-verified. Project feeds need `project_id`.
- `search` — Code Search is installed on this collection. `azuredevopssh_code_search_result` requires `search_text`.
- `audit` — the audit log route is not registered on this Azure DevOps Server 2022.1 collection. `azuredevopssh_audit_log` is implemented for servers that expose it.
- `testplan` — Test Plans may require a separate license.
- `memberentitlementmanagement` — member and user entitlement locations are not registered on this Azure DevOps Server 2022.1 collection. `/_apis/memberentitlements` and `/_apis/userentitlements` return 404. The tables are implemented for servers that expose the API.
- `graph` — group and membership locations are not registered on this collection. `azuredevopssh_group_membership` reads direct membership from the Identity API when the graph location is missing. `azuredevopssh_group` and `azuredevopssh_user` still call graph.
- `identity` — the v7 ReadIdentities location id is not registered. `azuredevopssh_identity` calls `/_apis/identities` (api-version 6.0). A list needs `filter_value`, `descriptor`, or `id`.
- `tokenadmin` — `azuredevopssh_personal_access_token` lists personal access token metadata and does not return the secret. A 401 or 403 means the calling token cannot administer tokens. SSH keys are excluded.

---

## Implementation conventions

- File/table naming: `table_azuredevopssh_<name>.go`, `tableAzureDevOps<Name>(ctx)`.
- Use `commonColumns(...)` for `organization` + `title` standard columns.
- Project-scoped tables: `ParentHydrate: listProjects` + wrapper struct carrying `ProjectId`; filter on `project_id` equals-qual before calling the API.
- Respect `d.RowsRemaining(ctx)` and `d.QueryContext.Limit`; implement continuation-token pagination where the go client supports it.
- Log errors with table/hydrate name as first arg.
- Add matching `docs/tables/*.md` with Steampipe SQL examples.

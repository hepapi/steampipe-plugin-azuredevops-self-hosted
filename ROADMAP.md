# Azure DevOps (Self-Hosted) Steampipe Plugin — Roadmap

This is a living, incremental roadmap for expanding the `steampipe-plugin-azuredevops` fork into a comprehensive observability plugin for **self-hosted Azure DevOps Server 2022.1** (primary target), while keeping **SaaS (`https://dev.azure.com/org`)** working (secondary).

- **Client:** `github.com/microsoft/azure-devops-go-api/azuredevops/v7` (v7.1.0)
- **REST reference:** https://learn.microsoft.com/en-us/rest/api/azure/devops/?view=azure-devops-rest-7.2

Each stage is independently shippable and testable. Checkboxes are marked `[x]` when a table is implemented, documented, and verified against the live collection.

---

## Stage 0 — Foundation (self-hosted + v7 + branch policies)

- [x] Migrate from v6 to v7 go client
- [x] Fix `organization` column derivation for self-hosted URLs (collection name)
- [x] Add `server_url` config + `getServerURL`/`getServerConnection` helpers
- [x] `azuredevops_project_collection` — `core.GetProjectCollections` / `GetProjectCollection`
- [x] `azuredevops_git_repository_branch_policy` — `policy.GetPolicyConfigurations` / `GetPolicyConfiguration`

---

## Stage 1 — Git & Code (pairs with branch policies)

- [x] `azuredevops_git_pull_request` — `git.GetPullRequests`
- [x] `azuredevops_git_pull_request_thread` — `git.GetThreads`
- [x] `azuredevops_git_pull_request_comment` — `git.GetComments`
- [x] `azuredevops_git_pull_request_reviewer` — `git.GetPullRequestReviewers`
- [x] `azuredevops_git_commit` — `git.GetCommits`
- [x] `azuredevops_git_ref` — `git.GetRefs`
- [x] `azuredevops_git_tag` — `git.GetAnnotatedTags`
- [x] Enrich `azuredevops_git_repository` (project name, default branch object, `is_disabled`)

---

## Stage 2 — Policy & Security

- [x] `azuredevops_policy_type` — `policy.GetPolicyTypes`
- [x] `azuredevops_policy_evaluation` — `policy.GetPolicyEvaluations`
- [x] `azuredevops_security_namespace` — `security.QuerySecurityNamespaces`
- [x] `azuredevops_permission` / ACLs — `security.QueryAccessControlLists`
- [x] `azuredevops_servicehook_subscription` — `servicehooks.ListSubscriptions`

---

## Stage 3 — Build & Release (deepen existing)

- [x] `azuredevops_build_timeline` — `build.GetBuildTimeline` (requires `project_id` and `build_id`)
- [x] `azuredevops_build_log` — `build.GetBuildLogs` (metadata only; requires `project_id` and `build_id`)
- [x] `azuredevops_build_artifact` — `build.GetArtifacts` (requires `project_id` and `build_id`)
- [x] `azuredevops_release_definition` — `release.GetReleaseDefinitions`
- [x] `azuredevops_release_environment` — environments expanded on `release.GetReleases`
- [x] `azuredevops_release_approval` — `release.GetApprovals` (pending, approved, rejected, canceled, skipped, reassigned)
- [x] `azuredevops_pipeline_run` — `pipelines.ListRuns` (requires `project_id` and `pipeline_id`)
- [x] `azuredevops_pipeline_check` — `pipelineschecks.GetCheckConfigurationsOnResource` (requires `project_id`, `resource_type`, and `resource_id`)
- [x] `azuredevops_pipeline_approval` — `pipelinesapproval.QueryApprovals` (requires `project_id` and `id`; Server 2022 does not list approvals without an id)

---

## Stage 4 — Work Item Tracking

- [ ] `azuredevops_work_item` — `workitemtracking.GetWorkItems` / `GetWorkItemsBatch`
- [ ] `azuredevops_work_item_type` — `workitemtracking.GetWorkItemTypes`
- [ ] `azuredevops_work_item_comment` — `workitemtracking.GetComments`
- [ ] `azuredevops_work_item_revision` — `workitemtracking.GetWorkItemRevisions`
- [ ] `azuredevops_work_item_relation` — `workitemtracking.GetWorkItemRelations`
- [ ] `azuredevops_work_item_field` — `workitemtracking.GetFields`
- [ ] `azuredevops_work_item_query` — `workitemtracking.QueryByWiql`

---

## Stage 5 — Test & Quality

- [x] `azuredevops_test_plan` — `testplan`
- [x] `azuredevops_test_suite` — `testplan`
- [x] `azuredevops_test_case` — `testplan`
- [ ] `azuredevops_test_run` — `testresults`
- [ ] `azuredevops_test_result` — `testresults`
- [x] `azuredevops_test_point` — `testplan`

---

## Stage 6 — Identity, Permissions & Entitlements

- [ ] `azuredevops_member_entitlement` — `memberentitlementmanagement`
- [ ] `azuredevops_group_membership` — `graph`
- [ ] `azuredevops_user_entitlement` — `memberentitlementmanagement`
- [ ] `azuredevops_identity` — `identity`

---

## Stage 7 — Artifacts / Packages

- [ ] `azuredevops_feed` — `feed`
- [ ] `azuredevops_package` — `feed` / `packagingshared`
- [ ] `azuredevops_package_version` — `feed`

---

## Stage 8 — Wiki, Search, Audit, Extensions, Agents

- [ ] `azuredevops_wiki` / `azuredevops_wiki_page` — `wiki`
- [ ] `azuredevops_code_search_result` — `search`
- [ ] `azuredevops_audit_log` — `audit`
- [ ] `azuredevops_extension` — `extensionmanagement`
- [ ] `azuredevops_agent_pool` / `azuredevops_agent` — `taskagent`
- [ ] `azuredevops_variable_group` — `taskagent` / `distributedtaskcommon`

---

## Self-hosted caveats

Some packages may be unavailable or limited on Azure DevOps Server 2022.1 depending on edition/licensing. Validate each against the live collection before committing; mark unavailable ones here:

- `feed` / `packagingshared` (Azure Artifacts) — may require a licensed Artifacts extension.
- `search` — requires the Code Search extension.
- `audit` — audit log streaming may be limited on self-hosted.
- `testplan` — Test Plans may require a separate license.
- `memberentitlementmanagement` — user entitlement management may differ on self-hosted.

---

## Implementation conventions

- File/table naming: `table_azuredevops_<name>.go`, `tableAzureDevOps<Name>(ctx)`.
- Use `commonColumns(...)` for `organization` + `title` standard columns.
- Project-scoped tables: `ParentHydrate: listProjects` + wrapper struct carrying `ProjectId`; filter on `project_id` equals-qual before calling the API.
- Respect `d.RowsRemaining(ctx)` and `d.QueryContext.Limit`; implement continuation-token pagination where the go client supports it.
- Log errors with table/hydrate name as first arg.
- Add matching `docs/tables/*.md` with `sql+postgres` + `sql+sqlite` examples.

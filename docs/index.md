---
organization: Hepapi
category: ["software development"]
brand_color: "#0090F1"
display_name: "Azure DevOps Server"
short_name: "azuredevopssh"
description: "Steampipe plugin to query a self-hosted Azure DevOps Server."
og_description: "Query a self-hosted Azure DevOps Server with SQL."
engines: ["steampipe"]
---

# Azure DevOps Server + Steampipe

[Steampipe](https://steampipe.io) queries APIs with SQL. This plugin reads a self-hosted [Azure DevOps Server](https://learn.microsoft.com/en-us/azure/devops/server/) collection. Hepapi maintains it for Server 2022.1. It began as a fork of Turbot's `azuredevops` plugin. Tables are named `azuredevopssh_*`.

```sql
select
  id,
  name,
  state,
  visibility
from
  azuredevopssh_project;
```

## Documentation

Table pages live in [tables/](tables/).

### Collections and projects

- [azuredevopssh_project_collection](tables/azuredevopssh_project_collection.md)
- [azuredevopssh_project](tables/azuredevopssh_project.md)
- [azuredevopssh_team](tables/azuredevopssh_team.md)
- [azuredevopssh_team_member](tables/azuredevopssh_team_member.md)
- [azuredevopssh_dashboard](tables/azuredevopssh_dashboard.md)

### Git

- [azuredevopssh_git_repository](tables/azuredevopssh_git_repository.md)
- [azuredevopssh_git_repository_branch](tables/azuredevopssh_git_repository_branch.md)
- [azuredevopssh_git_repository_branch_policy](tables/azuredevopssh_git_repository_branch_policy.md)
- [azuredevopssh_git_commit](tables/azuredevopssh_git_commit.md)
- [azuredevopssh_git_ref](tables/azuredevopssh_git_ref.md)
- [azuredevopssh_git_tag](tables/azuredevopssh_git_tag.md)
- [azuredevopssh_git_pull_request](tables/azuredevopssh_git_pull_request.md)
- [azuredevopssh_git_pull_request_thread](tables/azuredevopssh_git_pull_request_thread.md)
- [azuredevopssh_git_pull_request_comment](tables/azuredevopssh_git_pull_request_comment.md)
- [azuredevopssh_git_pull_request_reviewer](tables/azuredevopssh_git_pull_request_reviewer.md)

### Policy and security

- [azuredevopssh_policy_type](tables/azuredevopssh_policy_type.md)
- [azuredevopssh_policy_evaluation](tables/azuredevopssh_policy_evaluation.md)
- [azuredevopssh_security_namespace](tables/azuredevopssh_security_namespace.md)
- [azuredevopssh_permission](tables/azuredevopssh_permission.md)
- [azuredevopssh_servicehook_subscription](tables/azuredevopssh_servicehook_subscription.md)

### Build and release

- [azuredevopssh_build_definition](tables/azuredevopssh_build_definition.md)
- [azuredevopssh_build](tables/azuredevopssh_build.md)
- [azuredevopssh_build_timeline](tables/azuredevopssh_build_timeline.md)
- [azuredevopssh_build_log](tables/azuredevopssh_build_log.md)
- [azuredevopssh_build_artifact](tables/azuredevopssh_build_artifact.md)
- [azuredevopssh_pipeline](tables/azuredevopssh_pipeline.md)
- [azuredevopssh_pipeline_run](tables/azuredevopssh_pipeline_run.md)
- [azuredevopssh_pipeline_check](tables/azuredevopssh_pipeline_check.md)
- [azuredevopssh_pipeline_approval](tables/azuredevopssh_pipeline_approval.md)
- [azuredevopssh_release_definition](tables/azuredevopssh_release_definition.md)
- [azuredevopssh_release](tables/azuredevopssh_release.md)
- [azuredevopssh_release_environment](tables/azuredevopssh_release_environment.md)
- [azuredevopssh_release_approval](tables/azuredevopssh_release_approval.md)
- [azuredevopssh_serviceendpoint](tables/azuredevopssh_serviceendpoint.md)
- [azuredevopssh_variable_group](tables/azuredevopssh_variable_group.md)

### Work, test, and packages

- [azuredevopssh_work_item](tables/azuredevopssh_work_item.md)
- [azuredevopssh_test_plan](tables/azuredevopssh_test_plan.md)
- [azuredevopssh_test_suite](tables/azuredevopssh_test_suite.md)
- [azuredevopssh_test_case](tables/azuredevopssh_test_case.md)
- [azuredevopssh_test_point](tables/azuredevopssh_test_point.md)
- [azuredevopssh_test_run](tables/azuredevopssh_test_run.md)
- [azuredevopssh_test_result](tables/azuredevopssh_test_result.md)
- [azuredevopssh_feed](tables/azuredevopssh_feed.md)
- [azuredevopssh_package](tables/azuredevopssh_package.md)
- [azuredevopssh_package_version](tables/azuredevopssh_package_version.md)

### Identity, wiki, search, and agents

- [azuredevopssh_identity](tables/azuredevopssh_identity.md)
- [azuredevopssh_group](tables/azuredevopssh_group.md)
- [azuredevopssh_group_membership](tables/azuredevopssh_group_membership.md)
- [azuredevopssh_user](tables/azuredevopssh_user.md)
- [azuredevopssh_member_entitlement](tables/azuredevopssh_member_entitlement.md)
- [azuredevopssh_user_entitlement](tables/azuredevopssh_user_entitlement.md)
- [azuredevopssh_wiki](tables/azuredevopssh_wiki.md)
- [azuredevopssh_wiki_page](tables/azuredevopssh_wiki_page.md)
- [azuredevopssh_code_search_result](tables/azuredevopssh_code_search_result.md)
- [azuredevopssh_audit_log](tables/azuredevopssh_audit_log.md)
- [azuredevopssh_extension](tables/azuredevopssh_extension.md)
- [azuredevopssh_agent_pool](tables/azuredevopssh_agent_pool.md)
- [azuredevopssh_agent](tables/azuredevopssh_agent.md)

Branch-policy views are SQL files in [views/](views/).

## Quick start

### Install

```sh
steampipe plugin install ghcr.io/hepapi/azuredevopssh
```

### Credentials

| Item | Description |
| --- | --- |
| Credentials | A collection URL and a [personal access token](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows). On Azure DevOps Server the URL is a collection, for example `https://azuredevops.example.com/DefaultCollection`. On Azure DevOps Services it is `https://dev.azure.com/{organization}`. |
| Permissions | The token has the same permissions as the user who created it. |
| Radius | One connection is one collection, or one Azure DevOps Services organization. |
| Resolution | 1. Arguments in `~/.steampipe/config/azuredevops.spc`.<br />2. `AZDO_ORG_SERVICE_URL` and `AZDO_PERSONAL_ACCESS_TOKEN`, when those arguments are unset. |

### Configuration

Installing the plugin creates `~/.steampipe/config/azuredevops.spc` with a connection named `azuredevops`:

```hcl
connection "azuredevops" {
  plugin = "ghcr.io/hepapi/azuredevopssh"

  # `organization_url`: collection URL on Azure DevOps Server, or an organization URL on Azure DevOps Services. Required.
  # Server example: https://azuredevops.example.com/DefaultCollection
  # Services example: https://dev.azure.com/test
  # Can also be set with AZDO_ORG_SERVICE_URL.
  # organization_url = "https://azuredevops.example.com/DefaultCollection"

  # `server_url`: Azure DevOps Server root. Optional.
  # Used only by azuredevopssh_project_collection.
  # When unset, derived from organization_url by stripping the last path segment.
  # server_url = "https://azuredevops.example.com/"

  # `personal_access_token`: personal access token. Required.
  # Can also be set with AZDO_PERSONAL_ACCESS_TOKEN.
  # personal_access_token = "..."
}
```

```sh
export AZDO_ORG_SERVICE_URL=https://azuredevops.example.com/DefaultCollection
export AZDO_PERSONAL_ACCESS_TOKEN=...
```

List collections on the server:

```sql
select
  id,
  name,
  url
from
  azuredevopssh_project_collection;
```

## Server 2022.1 caveats

These routes are missing or limited on the Azure DevOps Server 2022.1 collection this plugin is developed against. The tables still exist for servers that expose the API.

- `azuredevopssh_audit_log` — the audit route is not registered.
- `azuredevopssh_member_entitlement` and `azuredevopssh_user_entitlement` — member entitlement locations are not registered, and the REST routes return 404.
- `azuredevopssh_group` and `azuredevopssh_user` call the Graph API, which is not registered on this collection. `azuredevopssh_group_membership` falls back to the Identity API.
- `azuredevopssh_identity` calls `/_apis/identities` (api-version 6.0) because the v7 Read Identities location is not registered. A list needs `filter_value`, `descriptor`, or `id`.
- `azuredevopssh_code_search_result` requires `search_text`. Code Search must be installed.
- `azuredevopssh_pipeline_approval` needs `project_id` and `id`. Server 2022 does not list approvals without an id.
- Test Plans and Azure Artifacts depend on the server edition and license. Package versions need a feed that contains packages.

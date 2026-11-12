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

[Steampipe](https://steampipe.io) queries APIs with SQL. This plugin reads a self-hosted [Azure DevOps Server](https://learn.microsoft.com/en-us/azure/devops/server/). Point `organization_url` at one collection, or at the server root to query every collection the token can list. Hepapi maintains it for Server 2022.1. It began as a fork of Turbot's `azuredevops` plugin. Tables are named `azuredevopssh_*`.

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
- [azuredevopssh_personal_access_token](tables/azuredevopssh_personal_access_token.md)
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
| Credentials | A collection URL, or a server root, and a [personal access token](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows). A collection URL is `https://azuredevops.example.com/DefaultCollection`. A server root is `https://azuredevops.example.com`. On Azure DevOps Services the URL is `https://dev.azure.com/{organization}`. |
| Permissions | The token has the same permissions as the user who created it. A server-root connection needs a token that can list project collections. |
| Radius | One connection is one collection, one server, or one Azure DevOps Services organization. `collections` limits the server to an allow-list. `excluded_collections` drops named collections, which is how you skip TFVC-only collections. A name in both lists is excluded. |
| Resolution | 1. Arguments in `~/.steampipe/config/azuredevops.spc`.<br />2. `AZDO_ORG_SERVICE_URL` and `AZDO_PERSONAL_ACCESS_TOKEN`, when those arguments are unset. |

### Configuration

Installing the plugin creates `~/.steampipe/config/azuredevops.spc` with a connection named `azuredevops`:

```hcl
connection "azuredevops" {
  plugin = "ghcr.io/hepapi/azuredevopssh"

  # `organization_url`: one collection, the server root, or an Azure DevOps Services organization. Required.
  # One collection: https://azuredevops.example.com/DefaultCollection
  # All collections: https://azuredevops.example.com
  # Services example: https://dev.azure.com/test
  # Can also be set with AZDO_ORG_SERVICE_URL.
  # organization_url = "https://azuredevops.example.com/DefaultCollection"

  # `collections`: optional allow-list of collection names. "*" discovers every collection.
  # collections = ["DefaultCollection"]

  # `excluded_collections`: optional deny-list. New collections stay in the scan until named here.
  # excluded_collections = ["LegacyTfs"]

  # `server_url`: Azure DevOps Server root. Optional.
  # Lists project collections. Also the URL for server-scoped tables when the connection is not pinned to one collection.
  # When unset, derived from organization_url by stripping the last path segment.
  # Set this when the server root itself has a path, such as https://azuredevops.example.com/tfs.
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

A server-root `organization_url` discovers project collections and runs each collection-scoped table once per remaining collection. `where organization = 'SomeCollection'` limits that scan to one collection. `collections` and `excluded_collections` are matched without case sensitivity. A name in either list that is not on the server fails the query and reports the names that were discovered. One collection API error still fails the query.

These tables are server-scoped. They run once against the server root and ignore `collections`, `excluded_collections`, and `where organization = '...'`:

- [azuredevopssh_project_collection](tables/azuredevopssh_project_collection.md) lists every collection, including ones the deny-list skips elsewhere.
- [azuredevopssh_agent_pool](tables/azuredevopssh_agent_pool.md) and [azuredevopssh_agent](tables/azuredevopssh_agent.md) read deployment pools. A pool's `auto_provision` flag creates a queue in each collection from that one pool.

On a single-collection URL, server-scoped tables keep using that URL.

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
- `azuredevopssh_group` calls `/_apis/graph/groups` when the Graph location is missing, and lists groups from the Identity service when that route is absent. `azuredevopssh_user` does the same with `/_apis/graph/users`. `azuredevopssh_group_membership` falls back to the Identity API.
- `azuredevopssh_identity` calls `/_apis/identities` (api-version 6.0) because the v7 Read Identities location is not registered. A list needs `filter_value`, `descriptor`, or `id`.
- `azuredevopssh_code_search_result` requires `search_text`. Code Search must be installed.
- `azuredevopssh_pipeline_approval` needs `project_id` and `id`. Server 2022 does not list approvals without an id.
- Test Plans and Azure Artifacts depend on the server edition and license. Package versions need a feed that contains packages.

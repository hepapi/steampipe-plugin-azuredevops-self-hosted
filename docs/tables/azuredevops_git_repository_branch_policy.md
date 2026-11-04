---
title: "Steampipe Table: azuredevops_git_repository_branch_policy - Query Azure DevOps Git Branch Policies using SQL"
description: "Allows users to query Git repository branch policies in Azure DevOps, providing insights into branch protection rules, required reviewers, build validation, status checks, and other policy configurations."
---

# Table: azuredevops_git_repository_branch_policy - Query Azure DevOps Git Branch Policies using SQL

Azure DevOps branch policies enforce code quality and review requirements on Git branches. The `azuredevops_git_repository_branch_policy` table provides insights into these policies, including whether they are enabled or blocking, the policy type (e.g. required reviewers, build validation, minimum approval count, status checks), and the branch/repository scope they apply to.

This table is a **faithful 1:1 view of the policy API**: one row per policy configuration. A single policy may target multiple branches or repositories, so the scope is stored as a JSON array in the `scope` column (also available inside `settings`). Use the [coverage views](../views/branch_policy_coverage.sql) to flatten the scope and answer "does this repo/branch have status checks and branch protection?".

## Table Usage Guide

The `azuredevops_git_repository_branch_policy` table provides insights into branch policies within Git repositories in Azure DevOps. As a DevOps engineer or security reviewer, explore policy-specific details through this table, including the policy type, enabled/blocking status, and the raw scope/settings JSON. Utilize it to audit branch protection coverage and ensure code quality gates are in place.

## Examples

### Basic info
List all branch policies with their type and status.

```sql+postgres
select
  id,
  project_id,
  policy_type_display_name,
  is_enabled,
  is_blocking
from
  azuredevops_git_repository_branch_policy;
```

```sql+sqlite
select
  id,
  project_id,
  policy_type_display_name,
  is_enabled,
  is_blocking
from
  azuredevops_git_repository_branch_policy;
```

### List enabled blocking policies
Identify policies that are both enabled and blocking (i.e. actively gating merges).

```sql+postgres
select
  id,
  policy_type_display_name
from
  azuredevops_git_repository_branch_policy
where
  is_enabled
  and is_blocking;
```

```sql+sqlite
select
  id,
  policy_type_display_name
from
  azuredevops_git_repository_branch_policy
where
  is_enabled
  and is_blocking;
```

### List status check policies
Status checks are a branch policy type whose display name is `Status` (or `Required status checks` on some versions).

```sql+postgres
select
  id,
  project_id,
  policy_type_id,
  settings
from
  azuredevops_git_repository_branch_policy
where
  lower(policy_type_display_name) in ('status', 'required status checks', 'status checks');
```

```sql+sqlite
select
  id,
  project_id,
  policy_type_id,
  settings
from
  azuredevops_git_repository_branch_policy
where
  lower(policy_type_display_name) in ('status', 'required status checks', 'status checks');
```

### Inspect the scope (branches/repositories a policy covers)
The `scope` column is a JSON array; each entry has `refName` (full ref), `repositoryId`, and `matchKind`.

```sql+postgres
select
  id,
  policy_type_display_name,
  scope
from
  azuredevops_git_repository_branch_policy;
```

```sql+sqlite
select
  id,
  policy_type_display_name,
  scope
from
  azuredevops_git_repository_branch_policy;
```

### Inspect raw policy settings
Examine the full policy settings JSON (e.g. required reviewer IDs, build definition ID, status name, filename patterns).

```sql+postgres
select
  id,
  policy_type_display_name,
  settings
from
  azuredevops_git_repository_branch_policy;
```

```sql+sqlite
select
  id,
  policy_type_display_name,
  settings
from
  azuredevops_git_repository_branch_policy;
```

## Coverage views

To audit whether every repository/branch has branch protection and status checks, use the coverage views in [`docs/views/branch_policy_coverage.sql`](../views/branch_policy_coverage.sql):

- `azuredevops_v_branch_policy_scope` — flattens each policy's `scope` JSON array into repository/ref rows.
- `azuredevops_v_branch_policy_kind` — classifies each policy as `status_check` or `branch_protection`.
- `azuredevops_v_branch_coverage` — one row per repository+branch with `has_branch_protection` and `has_status_check` boolean flags (gaps are included, not hidden).
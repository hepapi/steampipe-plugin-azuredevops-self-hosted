---
title: "Steampipe Table: azuredevopssh_policy_type - Query Azure DevOps Policy Types using SQL"
description: "Allows users to query Azure DevOps policy types, including the display name and description of each policy a project can configure."
---

# Table: azuredevopssh_policy_type - Query Azure DevOps Policy Types using SQL

Azure DevOps policy types describe the kinds of policies a project can apply, such as required reviewers, build validation, comment requirements, and status checks. The `azuredevopssh_policy_type` table lists those types per project. Configured policies themselves are in `azuredevopssh_git_repository_branch_policy`.

## Table Usage Guide

The `azuredevopssh_policy_type` table provides the catalog of policy types in each Azure DevOps project. As a DevOps engineer, use it to see which policy kinds a project supports and to resolve a policy type id from `azuredevopssh_git_repository_branch_policy` to a display name and description.

## Examples

### Basic info
List policy types and the project they were read from.

```sql
select
  id,
  project_id,
  display_name,
  description
from
  azuredevopssh_policy_type;
```

### List policy types in one project
Limit the catalog to a single project.

```sql
select
  id,
  display_name,
  description
from
  azuredevopssh_policy_type
where
  project_id = 'bdcdf70b-7757-4253-b36c-33c08ca07dbb';
```

### Find the status check policy type
Look up the policy type used for status checks.

```sql
select
  id,
  project_id,
  display_name
from
  azuredevopssh_policy_type
where
  lower(display_name) in ('status', 'required status checks', 'status checks');
```


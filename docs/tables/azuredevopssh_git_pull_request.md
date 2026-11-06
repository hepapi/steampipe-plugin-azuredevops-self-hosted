---
title: "Steampipe Table: azuredevopssh_git_pull_request - Query Azure DevOps Git Pull Requests using SQL"
description: "Allows users to query Git Pull Requests in Azure DevOps, providing details on each pull request's title, source and target branches, status, reviewers, and merge information."
---

# Table: azuredevopssh_git_pull_request - Query Azure DevOps Git Pull Requests using SQL

Azure DevOps is a service from Microsoft that provides developer services for support teams to plan work, collaborate on code development, and build and deploy applications. Git Pull Requests in Azure DevOps allow teams to review and approve code changes before they are merged into a target branch.

## Table Usage Guide

The `azuredevopssh_git_pull_request` table provides insights into Git Pull Requests within Azure DevOps. As a DevOps engineer, explore pull request details through this table, including author, creation/closing dates, source and target branches, merge status, reviewers, and labels. Utilize it for code review workflows, tracking in-flight changes, and auditing merge activity.

## Examples

### Basic info
Get the general information about pull requests across your repositories, such as their title, status, and merge status.

```sql+postgres
select
  pull_request_id,
  repository_id,
  title,
  status,
  merge_status,
  is_draft,
  creation_date
from
  azuredevopssh_git_pull_request;
```

```sql+sqlite
select
  pull_request_id,
  repository_id,
  title,
  status,
  merge_status,
  is_draft,
  creation_date
from
  azuredevopssh_git_pull_request;
```

### List open pull requests
Find all currently active pull requests that have not yet been completed.

```sql+postgres
select
  pull_request_id,
  title,
  source_ref_name,
  target_ref_name,
  created_by
from
  azuredevopssh_git_pull_request
where
  status = 'active';
```

```sql+sqlite
select
  pull_request_id,
  title,
  source_ref_name,
  target_ref_name,
  created_by
from
  azuredevopssh_git_pull_request
where
  status = 'active';
```

### List draft pull requests
Find pull requests that are still marked as drafts and are not ready to be merged.

```sql+postgres
select
  pull_request_id,
  title,
  created_by,
  creation_date
from
  azuredevopssh_git_pull_request
where
  is_draft;
```

```sql+sqlite
select
  pull_request_id,
  title,
  created_by,
  creation_date
from
  azuredevopssh_git_pull_request
where
  is_draft;
```

### List pull requests in a specific repository
Filter pull requests by repository to focus on a single codebase.

```sql+postgres
select
  pull_request_id,
  title,
  status,
  merge_status
from
  azuredevopssh_git_pull_request
where
  repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21';
```

```sql+sqlite
select
  pull_request_id,
  title,
  status,
  merge_status
from
  azuredevopssh_git_pull_request
where
  repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21';
```

### Find pull requests with merge conflicts
Identify pull requests that have failed to merge, which can help prioritize code review and conflict resolution.

```sql+postgres
select
  pull_request_id,
  title,
  merge_status,
  merge_failure_type,
  merge_failure_message
from
  azuredevopssh_git_pull_request
where
  merge_status = 'conflicts';
```

```sql+sqlite
select
  pull_request_id,
  title,
  merge_status,
  merge_failure_type,
  merge_failure_message
from
  azuredevopssh_git_pull_request
where
  merge_status = 'conflicts';
```
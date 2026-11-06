---
title: "Steampipe Table: azuredevopssh_git_commit - Query Azure DevOps Git Commits using SQL"
description: "Allows users to query Git Commits in Azure DevOps, providing details on each commit's author, message, parents, associated work items, and push information."
---

# Table: azuredevopssh_git_commit - Query Azure DevOps Git Commits using SQL

Azure DevOps is a service from Microsoft that provides developer services for support teams to plan work, collaborate on code development, and build and deploy applications. Git Commits in Azure DevOps represent the individual changes pushed to a repository, capturing author, message, parent history, and associated work items and pushes.

## Table Usage Guide

The `azuredevopssh_git_commit` table provides insights into Git Commits within Azure DevOps. As a DevOps engineer, explore commit details through this table, including the commit message, author, authoring date, parent commits, changes, and linked work items. Utilize it to audit code history, track contributors, and understand change volumes across your repositories.

## Examples

### Basic info
Get the general information about commits, such as their messages and authors.

```sql+postgres
select
  commit_id,
  repository_id,
  comment,
  author_name,
  author_date
from
  azuredevopssh_git_commit;
```

```sql+sqlite
select
  commit_id,
  repository_id,
  comment,
  author_name,
  author_date
from
  azuredevopssh_git_commit;
```

### List recent commits in a repository
Find the most recent commits in a specific repository, ordered by authoring date.

```sql+postgres
select
  commit_id,
  comment,
  author_name,
  author_date
from
  azuredevopssh_git_commit
where
  repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21'
order by
  author_date desc
limit
  20;
```

```sql+sqlite
select
  commit_id,
  comment,
  author_name,
  author_date
from
  azuredevopssh_git_commit
where
  repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21'
order by
  author_date desc
limit
  20;
```

### List commits from a specific author
Track the contribution of a particular developer across your repositories.

```sql+postgres
select
  repository_id,
  commit_id,
  comment,
  author_date
from
  azuredevopssh_git_commit
where
  author_email = 'dev@example.com'
order by
  author_date desc;
```

```sql+sqlite
select
  repository_id,
  commit_id,
  comment,
  author_date
from
  azuredevopssh_git_commit
where
  author_email = 'dev@example.com'
order by
  author_date desc;
```

### Find commits linked to a work item
Identify commits associated with work items, useful for tracing code changes back to tracked work.

```sql+postgres
select
  commit_id,
  repository_id,
  comment,
  work_items
from
  azuredevopssh_git_commit
where
  work_items is not null
  and json_array_length(work_items) > 0;
```

```sql+sqlite
select
  commit_id,
  repository_id,
  comment,
  work_items
from
  azuredevopssh_git_commit
where
  work_items is not null
  and json_array_length(work_items) > 0;
```
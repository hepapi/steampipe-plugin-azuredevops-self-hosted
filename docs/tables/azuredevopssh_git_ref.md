---
title: "Steampipe Table: azuredevopssh_git_ref - Query Azure DevOps Git Refs using SQL"
description: "Allows users to query Git Refs in Azure DevOps, providing details on each ref (branches, tags, and pull request refs), including the object it points to and its associated metadata."
---

# Table: azuredevopssh_git_ref - Query Azure DevOps Git Refs using SQL

Azure DevOps is a service from Microsoft that provides developer services for support teams to plan work, collaborate on code development, and build and deploy applications. Git Refs in Azure DevOps are named pointers to commits, including branches, tags, and pull request refs, that identify the current state of a line of work.

## Table Usage Guide

The `azuredevopssh_git_ref` table provides insights into Git Refs within Azure DevOps. As a DevOps engineer, explore ref details through this table, including the ref name, the object it points to, its creator, and whether it is locked. Utilize it to understand branch and tag layout and to audit ref-level metadata across your repositories.

## Examples

### Basic info
Get the general information about refs, such as their names and the objects they point to.

```sql+postgres
select
  repository_id,
  name,
  object_id,
  peeled_object_id,
  is_locked
from
  azuredevopssh_git_ref;
```

```sql+sqlite
select
  repository_id,
  name,
  object_id,
  peeled_object_id,
  is_locked
from
  azuredevopssh_git_ref;
```

### List branches in a repository
Find all branch refs in a specific repository.

```sql+postgres
select
  name,
  object_id
from
  azuredevopssh_git_ref
where
  repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21'
  and name like 'refs/heads/%';
```

```sql+sqlite
select
  name,
  object_id
from
  azuredevopssh_git_ref
where
  repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21'
  and name like 'refs/heads/%';
```

### Find locked refs
Identify refs that have been locked, which prevents updates to the referenced object.

```sql+postgres
select
  repository_id,
  name,
  object_id,
  is_locked_by
from
  azuredevopssh_git_ref
where
  is_locked;
```

```sql+sqlite
select
  repository_id,
  name,
  object_id,
  is_locked_by
from
  azuredevopssh_git_ref
where
  is_locked;
```

### List refs created by a specific user
Find refs created by a given identity, useful for auditing who created branches or other refs.

```sql+postgres
select
  repository_id,
  name,
  object_id
from
  azuredevopssh_git_ref
where
  creator ->> 'displayName' = 'Jane Doe';
```

```sql+sqlite
select
  repository_id,
  name,
  object_id
from
  azuredevopssh_git_ref
where
  json_extract(creator, '$.displayName') = 'Jane Doe';
```
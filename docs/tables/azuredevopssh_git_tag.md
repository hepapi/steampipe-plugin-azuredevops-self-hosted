---
title: "Steampipe Table: azuredevopssh_git_tag - Query Azure DevOps Git Tags using SQL"
description: "Allows users to query Git Tags in Azure DevOps, providing details on each tag, including its name, the object it points to, its creator, and the tag message for annotated tags."
---

# Table: azuredevopssh_git_tag - Query Azure DevOps Git Tags using SQL

Git Tags in Azure DevOps are named references used to mark specific points in repository history, such as release versions, and can be lightweight or annotated.

## Table Usage Guide

The `azuredevopssh_git_tag` table provides insights into Git Tags within Azure DevOps. As a DevOps engineer, explore tag details through this table, including the tag name, the commit it points to, its creator, and the tag message for annotated tags. Utilize it to audit release markers and understand what each tag in a repository corresponds to.

## Examples

### Basic info
Get the general information about tags, such as their names and the objects they point to.

```sql
select
  repository_id,
  name,
  object_id,
  peeled_object_id,
  message
from
  azuredevopssh_git_tag;
```

### List tags in a repository
Find all tags in a specific repository, ordered by name.

```sql
select
  name,
  object_id,
  message
from
  azuredevopssh_git_tag
where
  repository_id = '256e4f2a-9e1a-4b7f-9a8e-3f7a2d0c9b21'
order by
  name;
```

### List tags with a release message
Find annotated tags that carry a message, which often documents the release they mark.

```sql
select
  name,
  message,
  creator
from
  azuredevopssh_git_tag
where
  message is not null;
```

### List tags created by a specific user
Find tags created by a given identity, useful for auditing who cut release tags.

```sql
select
  repository_id,
  name,
  object_id
from
  azuredevopssh_git_tag
where
  creator ->> 'displayName' = 'Jane Doe';
```

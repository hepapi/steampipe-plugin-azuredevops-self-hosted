---
title: "Steampipe Table: azuredevops_security_namespace - Query Azure DevOps Security Namespaces using SQL"
description: "Allows users to query Azure DevOps security namespaces and the permission actions each namespace defines."
---

# Table: azuredevops_security_namespace - Query Azure DevOps Security Namespaces using SQL

Azure DevOps stores permissions in security namespaces. Each namespace (for example Git repositories, build, or project) defines the actions it protects and the bit mask for each action. The `azuredevops_security_namespace` table lists those namespaces. Access control entries that grant or deny those bits are in `azuredevops_permission`.

## Table Usage Guide

The `azuredevops_security_namespace` table provides the permission model for an Azure DevOps collection. As a security reviewer, use it to discover namespace ids and to decode the action names stored on `azuredevops_permission`.

## Examples

### Basic info
List security namespaces and how many actions each one defines.

```sql+postgres
select
  namespace_id,
  name,
  display_name,
  json_array_length(actions) as action_count
from
  azuredevops_security_namespace
order by
  name;
```

```sql+sqlite
select
  namespace_id,
  name,
  display_name,
  json_array_length(actions) as action_count
from
  azuredevops_security_namespace
order by
  name;
```

### Find the Git repositories namespace
Resolve the namespace id used for repository permissions.

```sql+postgres
select
  namespace_id,
  name,
  display_name,
  read_permission,
  write_permission
from
  azuredevops_security_namespace
where
  name = 'Git Repositories';
```

```sql+sqlite
select
  namespace_id,
  name,
  display_name,
  read_permission,
  write_permission
from
  azuredevops_security_namespace
where
  name = 'Git Repositories';
```

### List the actions in a namespace
Inspect the bit and name of each action a namespace secures.

```sql+postgres
select
  namespace_id,
  name,
  actions
from
  azuredevops_security_namespace
where
  name = 'Git Repositories';
```

```sql+sqlite
select
  namespace_id,
  name,
  actions
from
  azuredevops_security_namespace
where
  name = 'Git Repositories';
```

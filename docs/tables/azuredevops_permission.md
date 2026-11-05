---
title: "Steampipe Table: azuredevops_permission - Query Azure DevOps Access Control Entries using SQL"
description: "Allows users to query Azure DevOps access control entries, including the identity, token, and allow or deny permission bits."
---

# Table: azuredevops_permission - Query Azure DevOps Access Control Entries using SQL

Azure DevOps access control lists bind identities to security tokens inside a security namespace. The `azuredevops_permission` table returns one row per access control entry: the namespace, the token, the identity descriptor, the explicit allow and deny bits, and the action names those bits correspond to.

An unscoped query walks every namespace. Pass `namespace_id` to read one namespace. Namespaces the token cannot read are skipped unless the query is scoped to that namespace, in which case the API error is returned. `allow_actions` and `deny_actions` are decoded from `azuredevops_security_namespace.actions`.

## Table Usage Guide

The `azuredevops_permission` table provides explicit allow and deny entries for Azure DevOps security tokens. As a security reviewer, use it to see who is granted or denied a specific action on a repository, project, or other token.

## Examples

### List entries in one namespace
Scope the query to a namespace id. Replace the id with one from `azuredevops_security_namespace`.

```sql+postgres
select
  namespace_name,
  token,
  descriptor,
  allow,
  deny,
  allow_actions,
  deny_actions,
  inherit_permissions
from
  azuredevops_permission
where
  namespace_id = '2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87'
limit
  50;
```

```sql+sqlite
select
  namespace_name,
  token,
  descriptor,
  allow,
  deny,
  allow_actions,
  deny_actions,
  inherit_permissions
from
  azuredevops_permission
where
  namespace_id = '2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87'
limit
  50;
```

### Find entries that deny an action
List access control entries whose deny bit set is non-zero.

```sql+postgres
select
  namespace_name,
  token,
  descriptor,
  deny,
  deny_actions
from
  azuredevops_permission
where
  namespace_id = '2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87'
  and deny > 0;
```

```sql+sqlite
select
  namespace_name,
  token,
  descriptor,
  deny,
  deny_actions
from
  azuredevops_permission
where
  namespace_id = '2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87'
  and deny > 0;
```

### Read one token
Pass `token` to return the entries for that security token only.

```sql+postgres
select
  descriptor,
  allow_actions,
  deny_actions,
  extended_info
from
  azuredevops_permission
where
  namespace_id = '2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87'
  and token = 'repoV2/bdcdf70b-7757-4253-b36c-33c08ca07dbb';
```

```sql+sqlite
select
  descriptor,
  allow_actions,
  deny_actions,
  extended_info
from
  azuredevops_permission
where
  namespace_id = '2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87'
  and token = 'repoV2/bdcdf70b-7757-4253-b36c-33c08ca07dbb';
```

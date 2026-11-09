---
title: "Steampipe Table: azuredevopssh_identity - Query Azure DevOps Identities using SQL"
description: "Allows users to query Azure DevOps users and groups from the Identity service."
---

# Table: azuredevopssh_identity - Query Azure DevOps Identities using SQL

An identity is a user or a group known to the collection. The `azuredevopssh_identity` table reads them from the Identity service.

A list query needs `filter_value`, `descriptor`, or `id`. The Identity API rejects an empty search and `*`. `search_filter` chooses the match mode (`General`, `DisplayName`, `AccountName`, `MailAddress`, or `LocalGroupName`) and defaults to `General`.

On Azure DevOps Server 2022.1 the v7 location id for this API is not registered. The table calls `/_apis/identities` directly in that case. Direct membership is included on each row.

## Examples

### Find a collection group

```sql
select
  id,
  descriptor,
  subject_descriptor,
  provider_display_name,
  is_container
from
  azuredevopssh_identity
where
  search_filter = 'LocalGroupName'
  and filter_value = 'Project Collection Administrators';
```

### One identity by id

```sql
select
  id,
  provider_display_name,
  is_active,
  member_of
from
  azuredevopssh_identity
where
  id = '00000000-0000-0000-0000-000000000000';
```


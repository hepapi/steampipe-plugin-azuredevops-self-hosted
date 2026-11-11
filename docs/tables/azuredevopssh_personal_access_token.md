---
title: "Steampipe Table: azuredevopssh_personal_access_token - Query Azure DevOps Personal Access Tokens using SQL"
description: "Allows users to query Azure DevOps personal access token metadata, including expiration, without the secret value."
---

# Table: azuredevopssh_personal_access_token - Query Azure DevOps Personal Access Tokens using SQL

A personal access token is a credential a user creates to call Azure DevOps. This table lists those tokens for every user the caller is allowed to administer. Each row has the token name, scopes, owner, and expiration. The secret value is not returned.

The calling token needs permission to administer tokens. A 401 or 403 fails the query. Pass `subject_descriptor` to read one user. SSH keys are not included.

An unfiltered query walks Valid Users a few identities at a time and returns tokens as each user is found. It still has to check every user before it finishes, so a large server takes a while. `limit` stops early.

Tokens are deployment-wide. A server-root connection reads them once. A single-collection URL calls that collection.

## Examples

### Tokens and expiration

```sql
select
  display_name,
  user_display_name,
  scope,
  valid_from,
  valid_to,
  is_valid
from
  azuredevopssh_personal_access_token;
```

### Tokens that expire within 30 days

```sql
select
  display_name,
  user_display_name,
  valid_to
from
  azuredevopssh_personal_access_token
where
  valid_to < now() + interval '30 days'
order by
  valid_to;
```

### One user

```sql
select
  display_name,
  scope,
  valid_to
from
  azuredevopssh_personal_access_token
where
  subject_descriptor = 'aad.example';
```

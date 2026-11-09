---
title: "Steampipe Table: azuredevopssh_user_entitlement - Query Azure DevOps User Entitlements using SQL"
description: "Allows users to query Azure DevOps user licenses, extensions, and project entitlements."
---

# Table: azuredevopssh_user_entitlement - Query Azure DevOps User Entitlements using SQL

A user entitlement is the license, extensions, and project access assigned to a user. The `azuredevopssh_user_entitlement` table reads the member entitlement management API.

Azure DevOps Server 2022.1 does not register this API. On that server a query returns a location-not-registered error. The table is for organizations that expose user entitlements, including Azure DevOps Services.

The list returns the first page. `id` fetches one user. `filter` is passed through as the API `$filter`, for example `name eq 'ada'`.

## Examples

### List user licenses

```sql
select
  id,
  display_name,
  principal_name,
  account_license_type,
  license_display_name,
  license_status
from
  azuredevopssh_user_entitlement
limit
  20;
```

### One user by id

```sql
select
  id,
  display_name,
  license_display_name,
  extensions,
  project_entitlements
from
  azuredevopssh_user_entitlement
where
  id = '00000000-0000-0000-0000-000000000000';
```


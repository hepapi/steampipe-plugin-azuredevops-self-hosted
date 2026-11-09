---
title: "Steampipe Table: azuredevopssh_member_entitlement - Query Azure DevOps Member Entitlements using SQL"
description: "Allows users to query Azure DevOps member licenses and project entitlements."
---

# Table: azuredevopssh_member_entitlement - Query Azure DevOps Member Entitlements using SQL

A member entitlement is the license and project access assigned to a user, group, or service principal. The `azuredevopssh_member_entitlement` table reads the member entitlement management API.

Azure DevOps Server 2022.1 does not register this API. On that server a query returns a location-not-registered error. The table is for organizations that expose member entitlements, including Azure DevOps Services.

The generated client returns the first page. `filter` is passed through as the API `$filter`, for example `name eq 'ada'`.

## Examples

### List member licenses

```sql
select
  id,
  display_name,
  account_license_type,
  license_display_name,
  license_status
from
  azuredevopssh_member_entitlement
limit
  20;
```

### One member by id

```sql
select
  id,
  display_name,
  license_display_name,
  project_entitlements
from
  azuredevopssh_member_entitlement
where
  id = '00000000-0000-0000-0000-000000000000';
```


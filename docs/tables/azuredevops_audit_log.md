---
title: "Steampipe Table: azuredevops_audit_log - Query Azure DevOps Audit Logs using SQL"
description: "Allows users to query Azure DevOps audit log entries."
---

# Table: azuredevops_audit_log - Query Azure DevOps Audit Logs using SQL

An audit log entry records an action in the collection, such as creating a repository or changing a permission. Azure DevOps Server 2022.1 does not expose this API on every collection. When the route is missing, a query returns a not-found or location error.

## Examples

### Recent audit events

```sql+postgres
select
  id,
  action_id,
  actor_display_name,
  project_name,
  timestamp
from
  azuredevops_audit_log
order by
  timestamp desc
limit
  20;
```

```sql+sqlite
select
  id,
  action_id,
  actor_display_name,
  project_name,
  timestamp
from
  azuredevops_audit_log
order by
  timestamp desc
limit
  20;
```

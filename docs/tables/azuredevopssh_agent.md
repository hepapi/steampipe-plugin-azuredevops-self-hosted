---
title: "Steampipe Table: azuredevopssh_agent - Query Azure DevOps Agents using SQL"
description: "Allows users to query agents in an Azure DevOps agent pool."
---

# Table: azuredevopssh_agent - Query Azure DevOps Agents using SQL

An agent runs pipeline jobs. Agents belong to a deployment pool, so a server-root connection reads them once from the server rather than once per collection. Set `pool_id` to read one pool. A query without `pool_id` visits pools until the SQL `limit` is reached. `id` and `pool_id` together fetch one agent.

## Examples

### Agents in a pool

```sql
select
  id,
  name,
  version,
  status,
  enabled,
  os_description
from
  azuredevopssh_agent
where
  pool_id = 1
limit
  20;
```


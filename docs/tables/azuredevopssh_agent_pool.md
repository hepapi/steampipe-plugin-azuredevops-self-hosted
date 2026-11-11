---
title: "Steampipe Table: azuredevopssh_agent_pool - Query Azure DevOps Agent Pools using SQL"
description: "Allows users to query Azure DevOps agent pools."
---

# Table: azuredevopssh_agent_pool - Query Azure DevOps Agent Pools using SQL

An agent pool is a set of build and release agents. Pools are registered on the Azure DevOps Server deployment, so a server-root connection reads them once and does not repeat them per collection. A single-collection URL keeps calling that collection. `id` fetches one pool. `name` filters by pool name.

## Examples

### Agent pools

```sql
select
  id,
  name,
  pool_type,
  size,
  is_hosted
from
  azuredevopssh_agent_pool
limit
  20;
```


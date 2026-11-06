---
title: "Steampipe Table: azuredevops_agent_pool - Query Azure DevOps Agent Pools using SQL"
description: "Allows users to query Azure DevOps agent pools."
---

# Table: azuredevops_agent_pool - Query Azure DevOps Agent Pools using SQL

An agent pool is a set of build and release agents. The `azuredevops_agent_pool` table lists pools in the collection. `id` fetches one pool. `name` filters by pool name.

## Examples

### Agent pools

```sql+postgres
select
  id,
  name,
  pool_type,
  size,
  is_hosted
from
  azuredevops_agent_pool
limit
  20;
```

```sql+sqlite
select
  id,
  name,
  pool_type,
  size,
  is_hosted
from
  azuredevops_agent_pool
limit
  20;
```

---
title: "Steampipe Table: azuredevopssh_package - Query Azure Artifacts Packages using SQL"
description: "Allows users to query packages stored in an Azure Artifacts feed."
---

# Table: azuredevopssh_package - Query Azure Artifacts Packages using SQL

A package is one artifact name in a feed, such as a NuGet, npm, Maven, PyPI, or Universal package. `feed_id` is required. `id` and `feed_id` together fetch one package and its versions. `name_query` is a substring search, separate from the package `name`. `protocol_type` limits the list to one protocol.

## Examples

### Packages in a feed

```sql
select
  id,
  name,
  protocol_type,
  url
from
  azuredevopssh_package
where
  feed_id = '00000000-0000-0000-0000-000000000000'
limit
  20;
```

### One package

```sql
select
  id,
  name,
  protocol_type,
  versions
from
  azuredevopssh_package
where
  feed_id = '00000000-0000-0000-0000-000000000000'
  and id = '00000000-0000-0000-0000-000000000000';
```


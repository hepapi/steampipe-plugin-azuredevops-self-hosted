---
title: "Steampipe Table: azuredevops_feed - Query Azure Artifacts Feeds using SQL"
description: "Allows users to query Azure Artifacts feeds in a collection or a project."
---

# Table: azuredevops_feed - Query Azure Artifacts Feeds using SQL

A feed stores packages for one protocol or several. The `azuredevops_feed` table lists collection feeds. Set `project_id` to list feeds owned by that project. `id` fetches one feed.

## Examples

### Collection feeds

```sql+postgres
select
  id,
  name,
  project_id,
  upstream_enabled,
  hide_deleted_package_versions
from
  azuredevops_feed
limit
  20;
```

```sql+sqlite
select
  id,
  name,
  project_id,
  upstream_enabled,
  hide_deleted_package_versions
from
  azuredevops_feed
limit
  20;
```

### Feeds in one project

```sql+postgres
select
  id,
  name,
  description
from
  azuredevops_feed
where
  project_id = '00000000-0000-0000-0000-000000000000';
```

```sql+sqlite
select
  id,
  name,
  description
from
  azuredevops_feed
where
  project_id = '00000000-0000-0000-0000-000000000000';
```

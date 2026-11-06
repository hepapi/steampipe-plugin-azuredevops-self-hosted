---
title: "Steampipe Table: azuredevops_package_version - Query Azure Artifacts Package Versions using SQL"
description: "Allows users to query versions of a package in an Azure Artifacts feed."
---

# Table: azuredevops_package_version - Query Azure Artifacts Package Versions using SQL

A package version is one published release of a package. `feed_id` and `package_id` are required. `id`, `feed_id`, and `package_id` together fetch one version. `is_deleted` limits the list to deleted or current versions.

## Examples

### Versions of a package

```sql+postgres
select
  id,
  version,
  is_latest,
  is_deleted,
  publish_date,
  author
from
  azuredevops_package_version
where
  feed_id = '00000000-0000-0000-0000-000000000000'
  and package_id = '00000000-0000-0000-0000-000000000000'
order by
  publish_date desc;
```

```sql+sqlite
select
  id,
  version,
  is_latest,
  is_deleted,
  publish_date,
  author
from
  azuredevops_package_version
where
  feed_id = '00000000-0000-0000-0000-000000000000'
  and package_id = '00000000-0000-0000-0000-000000000000'
order by
  publish_date desc;
```

### One version

```sql+postgres
select
  id,
  version,
  description,
  dependencies
from
  azuredevops_package_version
where
  feed_id = '00000000-0000-0000-0000-000000000000'
  and package_id = '00000000-0000-0000-0000-000000000000'
  and id = '00000000-0000-0000-0000-000000000000';
```

```sql+sqlite
select
  id,
  version,
  description,
  dependencies
from
  azuredevops_package_version
where
  feed_id = '00000000-0000-0000-0000-000000000000'
  and package_id = '00000000-0000-0000-0000-000000000000'
  and id = '00000000-0000-0000-0000-000000000000';
```

---
title: "Steampipe Table: azuredevopssh_test_plan - Query Azure DevOps Test Plans using SQL"
description: "Allows users to query Azure DevOps test plans, including state, area, iteration, and owner."
---

# Table: azuredevopssh_test_plan - Query Azure DevOps Test Plans using SQL

A test plan groups suites and cases for a project. The `azuredevopssh_test_plan` table lists plans across projects. A query without `project_id` visits every project. `id` and `project_id` together fetch one plan.

Test Plans can require a separate license on Azure DevOps Server. An unlicensed collection returns an API error rather than an empty list.

## Examples

### Plans in a project

```sql+postgres
select
  id,
  name,
  state,
  area_path,
  iteration,
  updated_date
from
  azuredevopssh_test_plan
where
  project_id = '00000000-0000-0000-0000-000000000000'
order by
  name;
```

```sql+sqlite
select
  id,
  name,
  state,
  area_path,
  iteration,
  updated_date
from
  azuredevopssh_test_plan
where
  project_id = '00000000-0000-0000-0000-000000000000'
order by
  name;
```

### One plan

```sql+postgres
select
  id,
  name,
  state,
  root_suite
from
  azuredevopssh_test_plan
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and id = 12;
```

```sql+sqlite
select
  id,
  name,
  state,
  root_suite
from
  azuredevopssh_test_plan
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and id = 12;
```

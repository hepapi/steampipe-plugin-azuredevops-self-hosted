---
title: "Steampipe Table: azuredevops_test_run - Query Azure DevOps Test Runs using SQL"
description: "Allows users to query Azure DevOps test runs, including state, pass counts, and the linked plan or build."
---

# Table: azuredevops_test_run - Query Azure DevOps Test Runs using SQL

A test run is one execution of manual or automated tests. The `azuredevops_test_run` table lists runs per project. A query without `project_id` visits every project. Add a SQL `limit` when you only need the latest page. `id` and `project_id` together fetch one run.

Runs are read with the test REST API (`GetTestRuns`), which records pipeline and manual executions. Test Plans licensing is not required for this table.

## Examples

### Recent runs in a project

```sql+postgres
select
  id,
  name,
  state,
  is_automated,
  total_tests,
  passed_tests,
  completed_date
from
  azuredevops_test_run
where
  project_id = '00000000-0000-0000-0000-000000000000'
order by
  completed_date desc
limit
  20;
```

```sql+sqlite
select
  id,
  name,
  state,
  is_automated,
  total_tests,
  passed_tests,
  completed_date
from
  azuredevops_test_run
where
  project_id = '00000000-0000-0000-0000-000000000000'
order by
  completed_date desc
limit
  20;
```

### Automated runs

```sql+postgres
select
  id,
  name,
  state,
  passed_tests,
  total_tests
from
  azuredevops_test_run
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and is_automated = true
limit
  20;
```

```sql+sqlite
select
  id,
  name,
  state,
  passed_tests,
  total_tests
from
  azuredevops_test_run
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and is_automated = 1
limit
  20;
```

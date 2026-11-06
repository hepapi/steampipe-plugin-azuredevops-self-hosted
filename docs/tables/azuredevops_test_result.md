---
title: "Steampipe Table: azuredevops_test_result - Query Azure DevOps Test Results using SQL"
description: "Allows users to query results of one Azure DevOps test run."
---

# Table: azuredevops_test_result - Query Azure DevOps Test Results using SQL

A test result is one test inside a test run. `project_id` and `run_id` are required. `outcome` is sent to the API. `id`, `run_id`, and `project_id` together fetch one result.

## Examples

### Results of one run

```sql+postgres
select
  id,
  test_case_title,
  outcome,
  state,
  duration_in_ms,
  error_message
from
  azuredevops_test_result
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and run_id = 50
order by
  id;
```

```sql+sqlite
select
  id,
  test_case_title,
  outcome,
  state,
  duration_in_ms,
  error_message
from
  azuredevops_test_result
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and run_id = 50
order by
  id;
```

### Failed results

```sql+postgres
select
  id,
  test_case_title,
  failure_type,
  error_message
from
  azuredevops_test_result
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and run_id = 50
  and outcome = 'Failed';
```

```sql+sqlite
select
  id,
  test_case_title,
  failure_type,
  error_message
from
  azuredevops_test_result
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and run_id = 50
  and outcome = 'Failed';
```

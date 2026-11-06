---
title: "Steampipe Table: azuredevops_test_point - Query Azure DevOps Test Points using SQL"
description: "Allows users to query test points in one Azure DevOps test suite."
---

# Table: azuredevops_test_point - Query Azure DevOps Test Points using SQL

A test point is one test case paired with a configuration inside a suite. `project_id`, `plan_id`, and `suite_id` are required. `id`, `suite_id`, `plan_id`, and `project_id` together fetch one point.

## Examples

### Points in a suite

```sql+postgres
select
  id,
  test_case_id,
  test_case_name,
  configuration_name,
  outcome,
  is_active
from
  azuredevops_test_point
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and suite_id = 34;
```

```sql+sqlite
select
  id,
  test_case_id,
  test_case_name,
  configuration_name,
  outcome,
  is_active
from
  azuredevops_test_point
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and suite_id = 34;
```

### Failed points

```sql+postgres
select
  id,
  test_case_name,
  configuration_name,
  tester
from
  azuredevops_test_point
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and suite_id = 34
  and outcome = 'failed';
```

```sql+sqlite
select
  id,
  test_case_name,
  configuration_name,
  tester
from
  azuredevops_test_point
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and plan_id = 12
  and suite_id = 34
  and outcome = 'failed';
```

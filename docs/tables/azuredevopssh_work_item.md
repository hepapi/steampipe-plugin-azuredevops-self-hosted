---
title: "Steampipe Table: azuredevopssh_work_item - Query Azure DevOps Work Items using SQL"
description: "Allows users to query Azure DevOps work items, including type, state, assignment, area, iteration, and the full field map."
---

# Table: azuredevopssh_work_item - Query Azure DevOps Work Items using SQL

Work items track bugs, tasks, user stories, and other process types in Azure DevOps. The `azuredevopssh_work_item` table lists work items for each project. It resolves ids with a WIQL query, then loads each page with the batch work item API.

A query without `project_id` visits every project. Set `project_id` when you know the project. `id`, `work_item_type`, and `state` are sent to WIQL. `id` and `project_id` together fetch one work item. The `fields` column keeps the full field map, including process-specific fields such as story points. Relations are not expanded.

## Table Usage Guide

Use this table to inspect work item state, ownership, area and iteration paths, and the raw fields returned by Azure DevOps. Add `project_id` and a `limit` for large projects.

## Examples

### Open bugs in a project

```sql
select
  id,
  title,
  state,
  assigned_to_name,
  priority,
  changed_date
from
  azuredevopssh_work_item
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and work_item_type = 'Bug'
  and state = 'Active'
order by
  changed_date desc;
```

### One work item

```sql
select
  id,
  rev,
  title,
  work_item_type,
  state,
  url
from
  azuredevopssh_work_item
where
  project_id = '00000000-0000-0000-0000-000000000000'
  and id = 42;
```

### Count work items by type

```sql
select
  work_item_type,
  count(*) as work_items
from
  azuredevopssh_work_item
where
  project_id = '00000000-0000-0000-0000-000000000000'
group by
  work_item_type
order by
  work_items desc;
```


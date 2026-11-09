---
title: "Steampipe Table: azuredevopssh_release_definition - Query Azure DevOps Release Definitions using SQL"
description: "Allows users to query Azure DevOps classic release definitions, including path, revision, artifacts, and environments."
---

# Table: azuredevopssh_release_definition - Query Azure DevOps Release Definitions using SQL

A release definition is a classic release pipeline: artifacts, environments, variables, and triggers. The `azuredevopssh_release_definition` table lists those definitions per project.

An unscoped query walks every project. Pass `project_id` to read one project. `id` and `project_id` together fetch one definition.

## Table Usage Guide

Use this table to inventory classic release pipelines, their folder path, and the environments configured on them. Individual releases are in `azuredevopssh_release`.

## Examples

### Basic info

```sql
select
  project_id,
  id,
  name,
  path,
  revision,
  is_deleted
from
  azuredevopssh_release_definition
limit 20;
```

### Definitions in one project

```sql
select
  id,
  name,
  path,
  modified_on
from
  azuredevopssh_release_definition
where
  project_id = '00000000-0000-0000-0000-000000000000';
```


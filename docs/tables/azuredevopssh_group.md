---
title: "Steampipe Table: azuredevopssh_group - Query Azure DevOps Groups using SQL"
description: "Allows users to query Azure DevOps Groups, providing detailed insights into group membership, permissions, and associated metadata."
---

# Table: azuredevopssh_group - Query Azure DevOps Groups using SQL

The `azuredevopssh_group` table lists groups in the collection. It reads the Graph API when that location is registered. Azure DevOps Server 2022.1 often does not register the v7 Graph groups location (`ebbe6af8-0b91-4c13-8cf1-777c14858188`). The table then calls `/_apis/graph/groups` directly. When that route is also missing, it lists groups by walking the Valid Users group through the Identity service. On that path, `membership_state` is the identity active flag and `memberships` are the group's direct parent groups.

## Table Usage Guide

The `azuredevopssh_group` table provides insights into groups within Azure DevOps. As a DevOps engineer, explore group-specific details through this table, including group membership, permissions, and associated metadata. Utilize it to manage user permissions across multiple resources effectively and ensure the right level of access for each user.

## Examples

### Basic info
Explore which Azure DevOps groups are active within your domain. This is useful for understanding user access and permissions within your organization.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin,
  description
from
  azuredevopssh_group;
```

### List inactive groups
Identify groups within your Azure DevOps that are currently inactive. This could be useful for optimizing resource use, ensuring security, or maintaining an organized workspace.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin,
  description
from
  azuredevopssh_group
where
  not membership_state;
```

### List empty groups
Discover the segments that are classified as empty groups in the Azure DevOps platform. This can be useful to identify and clean up unused or redundant groups, enhancing the efficiency of your resource management.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin,
  description
from
  azuredevopssh_group
where
  memberships = '[]';
```

### Get parent group detail of a particular group
Determine the parent group details associated with a specific group in Azure DevOps. This can be useful for understanding group hierarchies and membership states, especially when managing user access and permissions.

```sql
select
  principal_name,
  display_name,
  membership_state,
  domain,
  origin,
  description
from
  azuredevopssh_group
where
  descriptor in
  (
    select
      m ->> 'containerDescriptor'
    from
      azuredevopssh_group,
      jsonb_array_elements(memberships) as m
    where
      display_name = 'Build Administrators'
  );
```

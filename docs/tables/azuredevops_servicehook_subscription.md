---
title: "Steampipe Table: azuredevops_servicehook_subscription - Query Azure DevOps Service Hook Subscriptions using SQL"
description: "Allows users to query Azure DevOps service hook subscriptions, including the event, consumer, status, and publisher or consumer inputs."
---

# Table: azuredevops_servicehook_subscription - Query Azure DevOps Service Hook Subscriptions using SQL

Azure DevOps service hooks send events, such as a pull request being created, to a consumer such as a web hook. The `azuredevops_servicehook_subscription` table lists those subscriptions for the organization or collection. `publisher_inputs` holds filters such as project or repository. `consumer_inputs` holds consumer settings such as the target URL.

## Table Usage Guide

The `azuredevops_servicehook_subscription` table provides service hook subscriptions in Azure DevOps. As a DevOps engineer, use it to audit which events leave the collection, which consumers receive them, and which subscriptions are disabled or on probation.

## Examples

### Basic info
List subscriptions with the event they listen for and the consumer that receives it.

```sql+postgres
select
  id,
  publisher_id,
  event_type,
  consumer_id,
  consumer_action_id,
  status,
  created_date
from
  azuredevops_servicehook_subscription;
```

```sql+sqlite
select
  id,
  publisher_id,
  event_type,
  consumer_id,
  consumer_action_id,
  status,
  created_date
from
  azuredevops_servicehook_subscription;
```

### List enabled web hook subscriptions
Find subscriptions that are enabled and delivered to a web hook.

```sql+postgres
select
  id,
  event_type,
  action_description,
  publisher_inputs,
  consumer_inputs
from
  azuredevops_servicehook_subscription
where
  status = 'enabled'
  and consumer_id = 'webHooks';
```

```sql+sqlite
select
  id,
  event_type,
  action_description,
  publisher_inputs,
  consumer_inputs
from
  azuredevops_servicehook_subscription
where
  status = 'enabled'
  and consumer_id = 'webHooks';
```

### List subscriptions for one event
Limit the list to pull request created events.

```sql+postgres
select
  id,
  status,
  consumer_id,
  consumer_action_id,
  created_by
from
  azuredevops_servicehook_subscription
where
  publisher_id = 'tfs'
  and event_type = 'git.pullrequest.created';
```

```sql+sqlite
select
  id,
  status,
  consumer_id,
  consumer_action_id,
  created_by
from
  azuredevops_servicehook_subscription
where
  publisher_id = 'tfs'
  and event_type = 'git.pullrequest.created';
```

create or replace view azuredevopssh_v_branch_policy_scope as
select
  p.id,
  p.project_id,
  p.policy_type_id,
  p.policy_type_display_name,
  p.is_enabled,
  p.is_blocking,
  p.is_deleted,
  nullif(s ->> 'repositoryId', '')             as repository_id,
  s ->> 'refName'                               as ref_name,
  replace(s ->> 'refName', 'refs/heads/', '')   as branch,
  s ->> 'matchKind'                             as match_kind
from
  azuredevopssh_git_repository_branch_policy p
  left join lateral jsonb_array_elements(
    case
      when jsonb_typeof(p.scope::jsonb) = 'array'
       and jsonb_array_length(p.scope::jsonb) > 0
        then p.scope::jsonb
      else '[{}]'::jsonb
    end
  ) s on true;

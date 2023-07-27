## 0.10.0 (May 9, 2023)

BREAKING CHANGES:

* `resources/pipes_organization`: Remove `avatar_url` argument

FEATURES:

* **New Resource:** `pipes_workspace_aggregator`

## 0.9.0 (February 23, 2023)

FEATURES:

* **New Resource:** `pipes_workspace_pipeline`
* **New Data Source:** `pipes_process`
* `resources/pipes_connection`: Add `plugin_version` attribute

## 0.8.0 (December 27, 2022)

FEATURES:

* `resources/pipes_workspace_snapshot`: Add `expires_at` attribute. 

## 0.7.0 (December 6, 2022)

FEATURES:

* **New Resource:** `pipes_user_preferences`

## 0.6.0 (August 22, 2022)

BREAKING CHANGES:

* `datasource/pipes_user`: Remove `email` attribute
* `resources/pipes_connection`: Resource to use `/` as a separator for its ID instead of `:`, e.g., `<org-handle>/<connection-handle>`
* `resources/pipes_organization_member`: Resource to use `/` as a separator for its ID instead of `:`, e.g., `<org-handle>/<user-handle>`
* `resources/pipes_organization_workspace_member`: Remove `email` attribute.
* `resources/pipes_organization_workspace_member`: Resource to use `/` as a separator for its ID instead of `:`, e.g., `<org-handle>/<workspace-handle>/<user-handle>`
* `resources/pipes_workspace`: Resource to use `/` as a separator for its ID instead of `:`, e.g., `<org-handle>/<workspace-handle>`
* `resources/pipes_workspace_connection`: Resource to use `/` as a separator for its ID instead of `:`, e.g., `<org-handle>/<workspace-handle>/<connection-handle>`
* `resources/pipes_workspace_mod`: Resource to use `/` as a separator for its ID instead of `:`, e.g., `<org-handle>/<workspace-handle>/<mod-alias>`
* `resources/pipes_workspace_mod_variable`: Resource to use `/` as a separator for its ID instead of `:`, e.g., `<org-handle>/<workspace-handle>/<mod-alias>/<variable-name>`

FEATURES:

* **New Resource:** `pipes_workspace_snapshot`

ENHANCEMENTS:

* `resources/pipes_organization_member`: Remove redundant call to get orgMember. 

## 0.5.0 (July 20, 2022)

FEATURES:

* **New Resource:** `pipes_organization_workspace_member`
* `resources/pipes_connection`: Add `created_at`, `updated_at`, `created_by`, `updated_by`, and `version_id` attributes
* `resources/pipes_organization`: Add `created_by`, and `updated_by` attributes
* `resources/pipes_organization_member`: Add `created_by`, `updated_by`, and `scope` attributes
* `resources/pipes_organization_member`: Modify the way organization members are listed, i.e. use the `List` call instead of `Invited` and `Accepted` calls that were used previously
* `resources/pipes_workspace`: Add `created_by`, and `updated_by` attributes
* `resources/pipes_workspace_connection`: Add `created_at`, `updated_at`, `created_by`, `updated_by`, `version_id`, and `identity_id` attributes
* `resources/pipes_workspace_mod`: Add `created_by`, `updated_by`, and `version_id` attributes

## 0.4.0 (March 31, 2022)

FEATURES:

* **New Resource:** `pipes_workspace_mod`
* **New Resource:** `pipes_workspace_mod_variable`

## 0.3.0 (March 4, 2022)

ENHANCEMENTS:

* `resources/pipes_connection`: Plugin connections are now defined in a `config` property and specific schemas are not required for new connection types. ([#33](https://github.com/turbot/terraform-provider-steampipecloud/issues/33))

BUG FIXES:

* `resources/pipes_connection`: Fix import for connections in an organization. ([#32](https://github.com/turbot/terraform-provider-steampipecloud/issues/32))
* `resources/pipes_workspace`: Fix import for workspaces in an organization. ([#32](https://github.com/turbot/terraform-provider-steampipecloud/issues/32))
* `resources/pipes_workspace_connection`: Fix import for workspace connections in an organization. ([#32](https://github.com/turbot/terraform-provider-steampipecloud/issues/32))

## 0.2.0 (December 17, 2021)

ENHANCEMENTS:

* `resources/pipes_connection`: Add support for `turbot` plugin. ([#26](https://github.com/turbot/terraform-provider-steampipecloud/issues/26))

BUG FIXES:

* `resources/pipes_workspace_connection`: Fix resource ID format when creating and deleting resources. ([#24](https://github.com/turbot/terraform-provider-steampipecloud/issues/24))

DOCUMENTATION:

* Update example usage in index doc to initialize plugin from `turbot/steampipecloud` instead of `hashicorp/steampipecloud`. ([#29](https://github.com/turbot/terraform-provider-steampipecloud/issues/29))

## 0.1.0 (December 16, 2021)

FEATURES:

* **New Resource:** `pipes_connection`
* **New Resource:** `pipes_organization`
* **New Resource:** `pipes_organization_member`
* **New Resource:** `pipes_workspace`
* **New Resource:** `pipes_workspace_connection`
* **New Data Source:** `pipes_organization`
* **New Data Source:** `pipes_user`

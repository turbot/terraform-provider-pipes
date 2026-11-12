## 0.17.2 (November 12, 2026)

BUG FIXES:

* Resources no longer fail `terraform plan` or `terraform apply` when the resource was already deleted in Turbot Pipes. Read removes it from state so the next plan recreates it, and Delete treats a `404` as already deleted. Before, `pipes_organization_connection_permission` and 16 other resources returned an error until the resource was removed from state by hand. (#51)

## 0.17.1 (November 11, 2026)

BREAKING CHANGES:

* A root module `output` that refers to `config` on one of the resources or data sources below must now set `sensitive = true`. This also applies to a value taken from `config`, such as `jsondecode(data.pipes_organization_integration.aws.config).role_name`. Without it, Terraform 0.15 and later and OpenTofu stop with `Output refers to sensitive values`. Outputs in child modules are not affected. You can add `sensitive = true` before you upgrade.
* A `for_each` that uses a value taken from `config` now stops with `Invalid for_each argument`. Wrap the value in `nonsensitive()`, for example `for_each = nonsensitive(toset(jsondecode(data.pipes_organization_integration.aws.config).regions))`. Make this change when you upgrade, not before: Terraform 1.5 refuses `nonsensitive()` on a value that is not sensitive.

SECURITY:

* `data/pipes_organization_integration`, `data/pipes_tenant_integration`, `data/pipes_user_integration`: `config` is now marked sensitive, the same as on the `pipes_organization_integration` and `pipes_tenant_integration` resources. Before, `terraform plan` showed it in clear text when an output or another resource referred to it. An integration's config can hold credentials for the system it connects to. Reported by [kta1kri](https://github.com/kta1kri). ([GHSA-f5px-m78q-8xxj](https://github.com/turbot/terraform-provider-pipes/security/advisories/GHSA-f5px-m78q-8xxj))
* `resource/pipes_connection`, `resource/pipes_organization_connection`, `resource/pipes_tenant_connection`, `resource/pipes_workspace_connection`, `resource/pipes_user_integration`: `config` is now marked sensitive. Before, a secret set in `config`, such as an AWS `secret_key`, appeared in clear text in plan output. Pipes does not return secrets, so the plan showed the secret as a change on every run. Use `config_wo` for secrets. ([GHSA-f5px-m78q-8xxj](https://github.com/turbot/terraform-provider-pipes/security/advisories/GHSA-f5px-m78q-8xxj))
* `provider`: `token` is now marked sensitive. Terraform does not show provider arguments in plan output, so nothing visible changes.
* A unit test now fails the build if an attribute named `config`, or an attribute whose name contains `secret`, `password`, `private_key` or `token`, is not marked sensitive. It also fails if a data source attribute is not sensitive where a resource attribute with the same name is.

Upgrading adds no plan changes and makes no API writes. Terraform still writes `config` to state in clear text, so protect the state file. The JSON plan (`terraform show -json`) also still holds the values; it now marks them in `after_sensitive`.

## 0.17.0 (October 17, 2025)

FEATURES:

* **New Resource:** `pipes_tenant_service_account` — Manage tenant-level service accounts (title, description).
* **New Resource:** `pipes_organization_service_account` — Manage organization-level service accounts (title, description).

ENHANCEMENTS:

* `pipes_workspace_datatank_table`:
  - Update no longer accepts changes to `part_per` and `source_table` (these are immutable after creation and are ignored on update).
  - Added write-only argument `skip_initial_refresh` on create/update to skip the initial refresh. This flag is not exported in state or read.

## 0.16.0 (September 4, 2025)

ENHANCEMENTS:
* Updated `hashicorp/terraform-plugin-sdk` to `v2.37.0`
* Updated `go-kit` to `v1.3.0`
* `resources/pipes_connection`: Added attributes `config_wo`, `config_wo_version` to support write-only configuration (i.e., not stored in state).
* `resources/pipes_organization_connection`: Added attributes `config_wo`, `config_wo_version` to support write-only configuration (i.e., not stored in state).
* `resources/pipes_tenant_connection`: Added attributes `config_wo`, `config_wo_version` to support write-only configuration (i.e., not stored in state).
* `resources/pipes_workspace_connection`: Added attributes `config_wo`, `config_wo_version` to support write-only configuration (i.e., not stored in state).
* `resources/pipes_organization_integration`: Added attributes `config_wo`, `config_wo_version` to support write-only configuration (i.e., not stored in state).
* `resources/pipes_tenant_integration`: Added attributes `config_wo`, `config_wo_version` to support write-only configuration (i.e., not stored in state).
* `resources/pipes_user_integration`: Added attributes `config_wo`, `config_wo_version` to support write-only configuration (i.e., not stored in state).

## 0.15.3 (August 13, 2025)

ENHANCEMENTS:

* Updated `pipes-sdk-go` to `v0.15.0`.
* **New Resource:** `pipes_tenant_settings`
* `resources/pipes_connection`: Added attributes `status`, `last_error_at`, `last_error_process_id`, `last_successful_update_at`, `last_successful_update_process_id`, `last_update_attempt_at`, `last_update_attempt_process_id`.
* `resources/pipes_organization_connection`: Added attributes `status`, `last_error_at`, `last_error_process_id`, `last_successful_update_at`, `last_successful_update_process_id`, `last_update_attempt_at`, `last_update_attempt_process_id`.
* `resources/pipes_tenant`: Added attribute `token_min_issued_at`.
* `resources/pipes_tenant_connection`: Added attributes `status`, `last_error_at`, `last_error_process_id`, `last_successful_update_at`, `last_successful_update_process_id`, `last_update_attempt_at`, `last_update_attempt_process_id`.
* `resources/pipes_workspace_connection`: Added attributes `status`, `last_error_at`, `last_error_process_id`, `last_successful_update_at`, `last_successful_update_process_id`, `last_update_attempt_at`, `last_update_attempt_process_id`.

## 0.15.2 (July 16, 2025)

ENHANCEMENTS:

* Updated `pipes-sdk-go` to `v0.14.0`.
* Support for creation of `db1.medium` workspace instance types.

## 0.15.1 (June 30, 2025)

BUG FIXES:

* Resource `pipes_tenant_connection`: Attribute `config` should store data from the response object.
* Resource `pipes_organization_connection`: Attribute `config` should store data from the response object.
* Resource `pipes_workspace_connection`: Attribute `config` should store data from the response object.

## 0.15.0 (February 5, 2025)

BREAKING CHANGES:

* Resource `pipes_organization_integration`: Attribute `github_installation_id` is now expected type `int` instead of `string`.
* Resource `pipes_tenant_integration`: Attribute `github_installation_id` is now expected type `int` instead of `string`.
* Resource `pipes_user_integration`: Attribute `github_installation_id` is now expected type `int` instead of `string`.

FEATURES:

* **New Resource:** `pipes_tenant_notifier`
* **New Resource:** `pipes_organization_notifier`
* **New Resource:** `pipes_user_notifier`
* **New Resource:** `pipes_workspace_notifier`
* **New Resource:** `pipes_workspace_flowpipe_mod`
* **New Resource:** `pipes_workspace_flowpipe_mod_variable`
* **New Resource:** `pipes_workspace_flowpipe_trigger`
* **New Data Source:** `pipes_organization_integration`
* **New Data Source:** `pipes_tenant_integration`
* **New Data Source:** `pipes_user_integration`
* **New Data Source:** `pipes_workspace`
* **New Data Source:** `pipes_workspace_flowpipe_pipeline`

ENHANCEMENTS:

* Updated some `string` based fields to be typed enums for better validation.
* `resources/pipes_tenant_member`: Add attribute `last_activity_at`.
* `resources/pipes_organization_member`: Add attribute `last_activity_at`.
* `resources/pipes_organization_workspace_member`: Add attribute `last_activity_at`.

## 0.14.0 (August 2, 2024)

BREAKING CHANGES:

* `resources/pipes_workspace_connection`: Resource functionality moved to manage connections at the workspace level. Previously, the resource used to manage `attachment` of connections to the workspace defined at the respective identity level. Please follow the [migration guide](https://github.com/turbot/terraform-provider-pipes/workspace_connection/docs/migrating.md) for migrating your existing configuration into the new model.
* `resources/pipes_connection`: Removed functionality to manage user level connections in line with changes in Pipes.

FEATURES:

* **New Resource:** `pipes_organization_connection`
* **New Resource:** `pipes_organization_connection_folder`
* **New Resource:** `pipes_organization_connection_folder_permission`
* **New Resource:** `pipes_organization_connection_permission`
* **New Resource:** `pipes_organization_integration`
* **New Resource:** `pipes_tenant_connection`
* **New Resource:** `pipes_tenant_connection_folder`
* **New Resource:** `pipes_tenant_connection_folder_permission`
* **New Resource:** `pipes_tenant_connection_permission`
* **New Resource:** `pipes_tenant_integration`
* **New Resource:** `pipes_user_integration`
* **New Resource:** `pipes_workspace_connection_folder`
* **New Resource:** `pipes_workspace_schema`

ENHANCEMENTS:

* `resources/pipes_workspace_mod`: Add attribute `state_reason`.

## 0.13.2 (March 21, 2024)

BUG FIXES: 

* `pipes_workspace_datatank_table`: Set `PartPer` setting for datatank table to be `nil` if nothing is passed in configuration while updating a datatank table. (#23)

ENHANCEMENTS:

* `resources/pipes_workspace`: Add support for passing `desired_state`, `db_volume_size_bytes` attribute when creating or updating a workspace. Add missing attribute `state_reason`.
* `resources/pipes_workspace_pipeline`: Add support for passing `desired_state` attribute when creating or updating a pipeline. Add attributes `state` and `state_reason`.
* `resources/pipes_workspace_datatank`: Add support for passing `desired_state` attribute when creating a datatank.
* `resources/pipes_workspace_datatank_table`: Add support for passing `desired_state` attribute when creating a datatank_table.

## 0.13.1 (March 7, 2024)

BUG FIXES: 

* `pipes_workspace_pipeline`: Format and pass value for pipeline `Tags` field only if a valid config is present. (#18)

## 0.13.0 (January 24, 2024)

FEATURES:

* **New Resource:** `pipes_tenant_member`
* **New Data Source:** `pipes_tenant`

ENHANCEMENTS:

* `resources/pipes_organization_member`: Add support for user to be automatically added to an organization in a custom tenant skipping the invite process.

## 0.12.1 (January 3, 2024)

BUG FIXES: 

* `pipes_workspace_datatank_table`: Set `PartPer` setting for datatank table to be `nil` if nothing is passed in configuration. (#14)

## 0.12.0 (November 3, 2023)

FEATURES:

* **New Resource:** `pipes_workspace_datatank`
* **New Resource:** `pipes_workspace_datatank_table`

ENHANCEMENTS:

* `resources/pipes_workspace`: Add support for setting `instance_type` for a workspace.

## 0.11.0 (July 27, 2023)

* The `Terraform Provider Steampipe Cloud` has been now been rebranded to use `Terraform Provider Turbot Pipes` instead:

FEATURES:

* **New Resource:** `pipes_connection`
* **New Resource:** `pipes_connection_test`
* **New Resource:** `pipes_organization`
* **New Resource:** `pipes_organization_member`
* **New Resource:** `pipes_organization_member_test`
* **New Resource:** `pipes_organization_test`
* **New Resource:** `pipes_organization_workspace_member`
* **New Resource:** `pipes_organization_workspace_member_test`
* **New Resource:** `pipes_user_preferences`
* **New Resource:** `pipes_user_preferences_test`
* **New Resource:** `pipes_workspace`
* **New Resource:** `pipes_workspace_aggregator`
* **New Resource:** `pipes_workspace_aggregator_test`
* **New Resource:** `pipes_workspace_connection`
* **New Resource:** `pipes_workspace_connection_test`
* **New Resource:** `pipes_workspace_mod`
* **New Resource:** `pipes_workspace_mod_test`
* **New Resource:** `pipes_workspace_mod_variable`
* **New Resource:** `pipes_workspace_mod_variable_test`
* **New Resource:** `pipes_workspace_pipeline`
* **New Resource:** `pipes_workspace_pipeline_test`
* **New Resource:** `pipes_workspace_snapshot`
* **New Resource:** `pipes_workspace_snapshot_test`
* **New Resource:** `pipes_workspace_test`
* **New Data Source:** `pipes_organization`
* **New Data Source:** `pipes_organization_test`
* **New Data Source:** `pipes_process`
* **New Data Source:** `pipes_user`
* **New Data Source:** `pipes_user_test`

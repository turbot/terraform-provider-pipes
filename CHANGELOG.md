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

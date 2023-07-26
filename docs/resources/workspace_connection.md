---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pipes_workspace_connection Resource - terraform-provider-pipes"
subcategory: ""
description: |-
  The `Turbot Pipes Workspace Connection` represents a set of connections that are currently attached to the workspace. This resource can be used multiple times with the same connection for non-overlapping workspaces.
---

# Resource: pipes_workspace_connection

Manages a workspace connection association.

## Example Usage

**Create a user workspace connection association**

```hcl
resource "pipes_workspace" "dev_workspace" {
  handle = "dev"
}

resource "pipes_connection" "dev_conn" {
  handle = "devconn"
  plugin = "bitbucket"
}

resource "pipes_workspace_connection" "test" {
  workspace_handle  = pipes_workspace.dev_workspace.handle
  connection_handle = pipes_connection.dev_conn.handle
}
```

**Create an organization workspace connection association**

```hcl
resource "pipes_workspace" "org_dev_workspace" {
  organization = "testorg"
  handle       = "dev"
}

resource "pipes_connection" "org_dev_conn" {
  organization = "testorg"
  handle       = "devconn"
  plugin       = "bitbucket"
}

resource "pipes_workspace_connection" "org_test" {
  organization      = "testorg"
  workspace_handle  = pipes_workspace.org_dev_workspace.handle
  connection_handle = pipes_connection.org_dev_conn.handle
}
```

## Argument Reference

The following arguments are supported:

- `connection_handle` - (Required) The handle of the connection to add to workspace.
- `workspace_handle` - (Required) The handle of the workspace to add the connection to.
- `organization` - (Optional) The organization ID or handle to create the connection association in.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `association_id` - An unique identifier of the workspace connection association.
- `connection_config` - A map of connection configuration.
- `connection_created_at` - The creation time of the connection.
- `connection_id` - An unique identifier of the connection.
- `connection_identity_id` - An unique identifier of the entity, where the connection is created.
- `connection_plugin` - The name of the plugin.
- `connection_type` - The type of the resource.
- `connection_updated_at` - The time when the connection was last updated.
- `connection_version_id` - The version of the connection.
- `created_at` - The time when the connection was associated to the workspace.
- `created_by` - The handle of the user who created the association.
- `identity_id` - The id of the user/organization to which the connection belongs to.
- `updated_at` - The time when the association was last updated.
- `updated_by` - The handle of the user who last updated the association.
- `version_id` - The association version.
- `workspace_created_at` - The creation time of the workspace.
- `workspace_database_name` - The name of the Steampipe workspace database.
- `workspace_hive` - The Steampipe workspace hive.
- `workspace_host` - The workspace hostname.
- `workspace_id` - An unique identifier of the workspace.
- `workspace_identity_id` - An unique identifier of the entity, where the workspace is created.
- `workspace_public_key` - The workspace public key.
- `workspace_state` - The current state of the workspace.
- `workspace_updated_at` - The time when the workspace was last updated.
- `workspace_version_id` - The workspace version.

## Import

### Import User Workspace Connection

User workspace connections can be imported an ID made up of `workspace_handle/connection_handle`, e.g.,

```sh
terraform import pipes_workspace_connection.example myworkspace/myconn
```

### Import Organization Workspace Connection

Organization workspace connections can be imported using an ID made up of `organization_handle/workspace_handle/connection_handle`, e.g.,

```sh
terraform import pipes_workspace_connection.example myorg/myworkspace/myconn
```
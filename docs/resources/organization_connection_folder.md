---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pipes_organization_connection_folder Resource - terraform-provider-pipes"
subcategory: ""
description: |-
  Provides a Pipes Connection Folder resource on an organization. The `Turbot Pipes Connection Folder` represents a grouping of a set of connections, which makes it easier to share them across workspaces in your tenant or organization.

  Connections Folders can be defined at the tenant, organization or workspace level.
---

# Resource: pipes_organization_connection_folder

Manages a connection folder, which is defined on an organization.

## Example Usage

**Create a connection folder in the `acme` organization**

```hcl
resource "pipes_organization_connection_folder" "devops_folder" {
  organization = "acme"
  title = "DevOps"
}
```

**Create a connection folder within another connection folder**

```hcl
resource "pipes_organization_connection_folder" "acme_folder" {
  organization = "acme"
  title = "Acme"
}

resource "pipes_organization_connection_folder" "devops_folder" {
  organization = "acme"
  title = "DevOps"
  parent_id = pipes_organization_connection_folder.acme_folder.id
}
```

**Create a connection within a connection folder**

```hcl
resource "pipes_organization_connection_folder" "acme_folder" {
  organization = "acme"
  title = "Acme"
}

resource "pipes_organization_connection_folder" "devops_folder" {
  organization = "acme"
  title = "DevOps"
  parent_id = pipes_organization_connection_folder.acme_folder.id
}

resource "pipes_organization_connection" "aws_aaa" {
  organization = "acme"
  plugin = "aws"
  handle = "aws_aaa"
  parent_id = pipes_organization_connection_folder.devops_folder.id
  config = jsonencode({
    access_key = "redacted"
    secret_key = "redacted"
    regions    = ["us-east-1"]
  })
}
```

## Argument Reference

The following arguments are supported:

- `organization` - (Required) The handle of the organization where the connection folder will be managed.
- `title` - (Required) A friendly title for your connection folder.
- `parent_id` - (Optional) Identifier of the connection folder in which this connection folder will be created. If nothing is passed the connection folder is created at the root level of the tenant.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `connection_folder_id` - Unique identifier of the connection folder.
- `created_at` - The time when the connection was created.
- `created_by` - The handle of the user who created the connection.
- `integration_resource_identifier` - Unique identifier of the resource if the connection folder is managed by an integration.
- `integration_resource_name` - Name of the resource if the connection folder is managed by an integration.
- `integration_resource_path` - Path of the resource in the hierarchy if the connection folder is managed by an integration.
- `integration_resource_type` - Type of the resource if the connection folder is managed by an integration. For example, if the connection folder is managed by an AWS integration, the value will be `aws_organization_unit`.
- `managed_by_id` - Unique identifier of the integration that is managing the connection, if any.
- `organization_id` - Unique identifier of the organization where the connection folder exists.
- `tenant_id` - Unique identifier of the tenant where the connection folder exists.
- `trunk` - An array of items that represent the path of the connection in the hierarchy, ordered from ancestor to parent.
- `type` - Denotes whether the connection folder is discovered or manually created. Possible values are `connection-folder` or `connection-folder-discovered`.
- `updated_at` - The time when the connection was last updated.
- `updated_by` - The handle of the user who last updated the connection.
- `version_id` - The connection version.

## Import

Organization connection folders can be imported using an ID made up of `organization_handle/connection_folder_id`, e.g.,

```sh
terraform import pipes_organization_connection_folder.example finance/c_cqlp0647sic7l5q2n5d0
```
---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pipes_organization Resource - terraform-provider-pipes"
subcategory: ""
description: |-
  The `Turbot Pipes Organization` includes multiple users and is intended for organizations to collaborate and share workspaces and connections.
---

# Resource: pipes_organization

Manages an organization.

## Example Usage

```hcl
resource "pipes_organization" "example" {
  handle       = "testorg"
  display_name = "Test Org"
}
```

## Argument Reference

The following arguments are supported:

- `handle` - (Required) A friendly identifier for your workspace, and must be unique across your workspaces.
- `display_name` - (Optional) A friendly name for your organization.
- `url` - (Optional) A publicly accessible URL for the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `created_at` - The creation time of the organization.
- `created_by` - The handle of the user who created the organization.
- `organization_id` - An unique identifier of the organization.
- `updated_at` - The time when the organization was last updated.
- `updated_by` - The handle of the user who last updated the organization.
- `version_id` - The organization version.

## Import

Workspaces can be imported using the `handle`, e.g.,

```sh
terraform import pipes_organization.example testorg
```
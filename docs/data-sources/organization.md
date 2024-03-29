---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pipes_organization Data Source - terraform-provider-pipes"
subcategory: ""
description: |-
  Use this data source to retrieve information about an existing organization.
---

# Data Source: pipes_organization

Use this data source to retrieve information about an existing organization.

## Example Usage

```terraform
data "pipes_organization" "org_aaa" {
  handle = "org_aaa"
}
```

```terraform
data "pipes_organization" "org_aaa" {
  organization_id = "o_c6rlv4gbb6gutha5gabc"
}
```

## Attributes Reference

The following attributes are exported.

- `handle` - Handle of the organization in Pipes.
- `organization_id` - ID of the organization in Pipes.
- `display_name` - Display name of the organization.
- `avatar_url` - URL of the Avatar for organization profile.
- `url` - Url in the organization profile.
- `created_at` - Created at timestamp of the organization.

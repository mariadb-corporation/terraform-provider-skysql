---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "skysql_credentials Data Source - terraform-provider-skysql"
subcategory: ""
description: |-
  Default credentials for connecting to a MariaDB service deployed by SkySQL
---

# skysql_credentials (Data Source)

Default credentials for connecting to a MariaDB service deployed by SkySQL

## Example Usage

```terraform
data "skysql_credentials" "wat" {
  id = "db00008965"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **id** (String) The ID of this resource.

### Read-Only

- **password** (String, Sensitive)
- **username** (String)



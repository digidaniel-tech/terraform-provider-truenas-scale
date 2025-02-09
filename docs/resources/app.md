---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "truenas-scale_app Resource - truenas-scale"
subcategory: ""
description: |-
  App resource
---

# truenas-scale_app (Resource)

App resource



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_name` (String) Application name to be set for installed application

### Optional

- `catalog_app` (String) Catalog app to use when installing application
- `custom_app` (Boolean) Catalog or custom app
- `custom_compose_config` (Object) Custom app configuration as an object (see [below for nested schema](#nestedatt--custom_compose_config))
- `custom_compose_config_string` (String) Custom app configuration as yaml
- `train` (String) Train to use when download application, ex. stable, test, community.
- `values` (Object) Application settings, ex. volumes, environment variables. (see [below for nested schema](#nestedatt--values))
- `version` (String) Version of application to use

<a id="nestedatt--custom_compose_config"></a>
### Nested Schema for `custom_compose_config`

Optional:



<a id="nestedatt--values"></a>
### Nested Schema for `values`

Optional:

# Blueprint

Blueprint is a command for terraform scaffolding generation for resources and data sources for Yandex tf provider.

## Installation

`make install`


## Command usage

### Iam member resource generation

 * Make sure that blueprint is installed
 * Return to the project root directory
 * Run command `blueprint generate resource --service-name=name_of_service --name=resource_name --template=iam_member`
 * Flags description:
 * `--service-name` - set service with resource is related (example: `--service-name=compute`)
 * `--name` - set a name of the resource with iam_member resource should be generated. (example: `--name=disk`, `--name=instance`)
 * `--template=iam_member` - the name of template that command use for generate resources/data sources. Now command has only one template: `iam_member
 * `--force` - use this if you want to overwrite exiting file.
 * Make sure that you've read and removed all `TIP:` in the generated file.

--- 

#### Command examples:
* Create an iam_member resource for compute.disk: ``` blueprint generate resource --service-name=compute --name=disk --template=iam_member```
* Create an iam_member resource for compute.instance: ```  blueprint generate resource --service-name=compute --name=instance --template=iam_member```

Two files will have been created by the end of command execution: 
* `yandex-framework/services/compute/disk/resource_iam_member.go` 
* `yandex-framework/services/compute/instance/resource_iam_member.go`


## Code generation

After change in `specification.json` run
```shell
~/go/bin/tfplugingen-framework generate resources --package spark_cluster --input specification.json  --output .
~/go/bin/tfplugingen-framework generate data-sources --package spark_cluster --input specification.json  --output .
```
Remove duplicated definitions in cluster_data_source_gen.go

For more information on code generation see [terraform documentation](https://developer.hashicorp.com/terraform/plugin/code-generation).

package mdb_redis_cluster_v2

type dataSourceCluster struct {
	clusterModel

	Config *dataSourceConfig `tfsdk:"config"`
}

func (c *dataSourceCluster) commonCluster() *clusterModel {
	return &c.clusterModel
}

func (c *dataSourceCluster) commonConfig() *configModel {
	if c.Config == nil {
		c.Config = &dataSourceConfig{}
	}
	return &c.Config.configModel
}

type dataSourceConfig struct {
	configModel
}

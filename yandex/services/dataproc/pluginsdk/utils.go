package pluginsdk

func convertStringMap(v map[string]interface{}) map[string]string {
	m := make(map[string]string)
	if v == nil {
		return m
	}
	for k, val := range v {
		m[k] = val.(string)
	}
	return m
}

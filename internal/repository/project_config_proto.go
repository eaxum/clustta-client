package repository

import "clustta/internal/repository/repositorypb"

func ToPbProjectConfigs(configs []ProjectConfig) []*repositorypb.ProjectConfig {
	result := make([]*repositorypb.ProjectConfig, 0, len(configs))
	for _, config := range configs {
		result = append(result, &repositorypb.ProjectConfig{
			Name: config.Name, Value: config.Value, Mtime: int64(config.Mtime),
		})
	}
	return result
}

func FromPbProjectConfigs(configs []*repositorypb.ProjectConfig) []ProjectConfig {
	result := make([]ProjectConfig, 0, len(configs))
	for _, config := range configs {
		result = append(result, ProjectConfig{
			Name: config.Name, Value: config.Value, Mtime: int(config.Mtime), Synced: true,
		})
	}
	return result
}

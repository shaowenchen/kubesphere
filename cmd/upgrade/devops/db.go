package devops

import (
	"kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/devops"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
)

func QueryDevops()([]*v1alpha2.DevOpsProject, error){
	dbconn, err := cs.ClientSets().MySQL()
	if err != nil {
		if _, ok := err.(cs.ClientSetNotEnabledError); ok {
			return nil, err
		}
		return nil, err
	}

	query := dbconn.Select(devops.GetColumnsFromStructWithPrefix(devops.DevOpsProjectTableName, v1alpha2.DevOpsProject{})...).
		From(devops.DevOpsProjectTableName)
	projects := make([]*v1alpha2.DevOpsProject, 0)
	_, err = query.Load(&projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func QueryPipeline(project string)([]*v1alpha2)
package devops

type DevOpsIM struct {
	workspace string
	devops    string
	username  string
	role      string
}

func GetDevOpsIm() []*DevOpsIM {

	// query devops
	projects, err := QueryDevops()
	if err != nil {
		DevOpsLogger().Println("query devops: %v", err)
	}
	result := make([]*DevOpsIM, 0)
	for _, project := range projects {
		//query memeber upder project
		membershipList, err := QueryProjectMemberShip(project.ProjectId)
		if err != nil {
			continue
		}
		for _, memeber := range membershipList {
			result = append(result, &DevOpsIM{workspace: project.Workspace, devops: project.ProjectId, username: memeber.Username, role: memeber.Role})
		}
	}

	return result

}

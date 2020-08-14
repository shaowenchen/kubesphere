package devops

import (
	"encoding/json"
	"fmt"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/simple/client"
)

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
			result = append(result, &DevOpsIM{workspace: project.Workspace, devops: GetVaildName(project.ProjectId), username: memeber.Username, role: memeber.Role})
		}
	}

	return result
}

func migrateDevOpsRoleBinding(oldRoleBindings []*DevOpsIM) error  {
	k8sClient :=  client.ClientSets().K8s().Kubernetes()
	roleBindings := make([]rbacv1.RoleBinding, 0)
	migrateMapping := map[string]string{
		"owner":      "admin",
		"maintainer": "operator",
		"developer":  "operator",
		"reporter":   "viewer",
	}

	for _, oldRoleBinding := range oldRoleBindings {
		role := migrateMapping[oldRoleBinding.role]
		if role == "" {
			continue
		}
		roleBinding := rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", oldRoleBinding.username, role),
				Namespace: oldRoleBinding.devops,
				Labels: map[string]string{
					"iam.kubesphere.io/user-ref": oldRoleBinding.username,
				},
			},
			Subjects: []rbacv1.Subject{{Name: oldRoleBinding.username, Kind: "User", APIGroup: "rbac.authorization.k8s.io"}},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     role,
			},
		}
		roleBindings = append(roleBindings, roleBinding)
	}

	for _, roleBinding := range roleBindings {
		outputData, _ := json.Marshal(roleBinding)
		DevOpsLogger().Infof("migrate devopsRoleBinding: namespace:%s, %s: %s", roleBinding.Namespace, roleBinding.Name, string(outputData))

		err := k8sClient.RbacV1().RoleBindings(roleBinding.Namespace).Delete(roleBinding.Name, metav1.NewDeleteOptions(0))
		if err != nil && !errors.IsNotFound(err) {
			DevOpsLogger().Error(err)
			return err
		}

		_, err = k8sClient.RbacV1().RoleBindings(roleBinding.Namespace).Create(&roleBinding)
		if err != nil && !errors.IsAlreadyExists(err) {
			DevOpsLogger().Error(err)
			return err
		}
	}
	return nil
}

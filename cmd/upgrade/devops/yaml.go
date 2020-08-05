package devops

import (
	"fmt"
	"github.com/lithammer/dedent"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}



func CreateDir(path string) error {
	if IsExist(path) == false {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateDevOpsProjectYaml(filename string, variables map[string]interface{}) error {

	var tmpl = template.Must(template.New(filename).Parse(
		dedent.Dedent(`---
apiVersion: v1
kind: Namespace
metadata:
  finalizers:
  - finalizers.kubesphere.io/namespaces
  generateName: svn
  labels:
    kubesphere.io/devopsproject: svn
    kubesphere.io/namespace: svnfvr6r
  name: svnfvr6r
  ownerReferences:
  - apiVersion: devops.kubesphere.io/v1alpha3
    blockOwnerDeletion: true
    controller: true
    kind: DevOpsProject
    name: svn
    uid: 8489354c-86b1-45c5-975c-28f6aac1e2fb
  resourceVersion: "4654830"
  selfLink: /api/v1/namespaces/svnfvr6r
  uid: 89c55b39-9aaf-4c12-89de-959d4d57d7e2
spec:
  finalizers:
  - kubernetes
status:
  phase: Active

---
apiVersion: devops.kubesphere.io/v1alpha3
kind: DevOpsProject
metadata:
  annotations:
    kubesphere.io/creator: admin
    kubesphere.io/workspace: w2
  creationTimestamp: "2020-07-27T08:42:36Z"
  finalizers:
  - devopsproject.finalizers.kubesphere.io
  generation: 2
  labels:
    kubesphere.io/workspace: w2
  name: svn
  ownerReferences:
  - apiVersion: tenant.kubesphere.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Workspace
    name: w2
    uid: 04e1dbcd-63af-45de-998e-ec39042193cc
  resourceVersion: "5261623"
  selfLink: /apis/devops.kubesphere.io/v1alpha3/devopsprojects/svn
  uid: 8489354c-86b1-45c5-975c-28f6aac1e2fb
spec: {}
status:
  adminNamespace: svnfvr6r
    `)))
	var buf strings.Builder
	if err := tmpl.Execute(&buf, variables); err != nil {
		return err
	}
	var path = "./" + DevOpsDir + "/" + filename + "/"
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	err := ioutil.WriteFile(fmt.Sprintf( path +  filename + ".yaml"), []byte(buf.String()), 0644)
	if err != nil {
		return err
	}
	return nil
}

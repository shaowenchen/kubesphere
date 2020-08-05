package devops

import (
	"fmt"
	"github.com/lithammer/dedent"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/models/devops"
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

func GeneratePipelineYaml(project string, filename string, pipeline devops.ProjectPipeline) error {

	var pipelineTmpl = template.Must(template.New(filename).Parse(
		dedent.Dedent(`---
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  annotations:
    kubesphere.io/creator: admin
  creationTimestamp: "2020-07-27T09:25:22Z"
  finalizers:
  - pipeline.finalizers.kubesphere.io
  generation: 1
  name: emptysvn
  namespace: a11tk9ph
  resourceVersion: "4665083"
  selfLink: /apis/devops.kubesphere.io/v1alpha3/namespaces/a11tk9ph/pipelines/emptysvn
  uid: d68e8975-c3af-4713-a395-6d9ce693406d
spec:
  multi_branch_pipeline:
    discarder:
      days_to_keep: "-1"
      num_to_keep: "-1"
    name: emptysvn
    script_path: Jenkinsfile
    source_type: svn
    svn_source:
      credential_id: svn
      includes: trunk,branches/*,tags/*,sandbox/*
      remote: svn://svnbucket.com/shaowenchen/empty/
    timer_trigger:
      interval: "600000"
  type: multi-branch-pipeline
status: {}
    `)))

	var multiBranchpipelineTmpl = template.Must(template.New(filename).Parse(
		dedent.Dedent(`---
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  annotations:
    kubesphere.io/creator: admin
  creationTimestamp: "2020-07-27T09:25:22Z"
  finalizers:
  - pipeline.finalizers.kubesphere.io
  generation: 1
  name: emptysvn
  namespace: a11tk9ph
  resourceVersion: "4665083"
  selfLink: /apis/devops.kubesphere.io/v1alpha3/namespaces/a11tk9ph/pipelines/emptysvn
  uid: d68e8975-c3af-4713-a395-6d9ce693406d
spec:
  multi_branch_pipeline:
    discarder:
      days_to_keep: "-1"
      num_to_keep: "-1"
    name: emptysvn
    script_path: Jenkinsfile
    source_type: svn
    svn_source:
      credential_id: svn
      includes: trunk,branches/*,tags/*,sandbox/*
      remote: svn://svnbucket.com/shaowenchen/empty/
    timer_trigger:
      interval: "600000"
  type: pipeline
status: {}
    `)))
	var buf strings.Builder
	if pipeline.Type == "multi-branch-pipeline" {
		if err := multiBranchpipelineTmpl.Execute(&buf, pipeline); err != nil {
			return err
		}
	}else if pipeline.Type == "pipeline" {
		if err := pipelineTmpl.Execute(&buf, pipeline); err != nil {
			return err
		}
	}else {
		return nil
	}
	var path = "./" + DevOpsDir + "/" + project + "/pipeline/"
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	err := ioutil.WriteFile(fmt.Sprintf( path +  filename + ".yaml"), []byte(buf.String()), 0644)
	if err != nil {
		return err
	}
	return nil
}

func GenerateSecretYaml(project string, filename string, secret *devops.JenkinsCredential) error {

	var basic_auth = template.Must(template.New(filename).Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  password: MGt5MDIwMzA1
  username: c2hhb3dlbmNoZW4=
kind: Secret
metadata:
  annotations:
    kubesphere.io/creator: admin
  finalizers:
  - finalizers.kubesphere.io/credential
  labels:
    app: svn
  name: svn
  namespace: a11tk9ph
  resourceVersion: "4660310"
  selfLink: /api/v1/namespaces/a11tk9ph/secrets/svn
  uid: 10774690-ff7b-43b7-b043-c4c805dc37b9
type: credential.devops.kubesphere.io/basic-auth
    `)))

	var ssh_auth = template.Must(template.New(filename).Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  password: MGt5MDIwMzA1
  username: c2hhb3dlbmNoZW4=
kind: Secret
metadata:
  annotations:
    kubesphere.io/creator: admin
  finalizers:
  - finalizers.kubesphere.io/credential
  labels:
    app: svn
  name: svn
  namespace: a11tk9ph
  resourceVersion: "4660310"
  selfLink: /api/v1/namespaces/a11tk9ph/secrets/svn
  uid: 10774690-ff7b-43b7-b043-c4c805dc37b9
type: credential.devops.kubesphere.io/ssh_auth
    `)))

	var secret_text = template.Must(template.New(filename).Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  password: MGt5MDIwMzA1
  username: c2hhb3dlbmNoZW4=
kind: Secret
metadata:
  annotations:
    kubesphere.io/creator: admin
  finalizers:
  - finalizers.kubesphere.io/credential
  labels:
    app: svn
  name: svn
  namespace: a11tk9ph
  resourceVersion: "4660310"
  selfLink: /api/v1/namespaces/a11tk9ph/secrets/svn
  uid: 10774690-ff7b-43b7-b043-c4c805dc37b9
type: credential.devops.kubesphere.io/secret_text
    `)))
	var kubeconfig = template.Must(template.New(filename).Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  password: MGt5MDIwMzA1
  username: c2hhb3dlbmNoZW4=
kind: Secret
metadata:
  annotations:
    kubesphere.io/creator: admin
  finalizers:
  - finalizers.kubesphere.io/credential
  labels:
    app: svn
  name: svn
  namespace: a11tk9ph
  resourceVersion: "4660310"
  selfLink: /api/v1/namespaces/a11tk9ph/secrets/svn
  uid: 10774690-ff7b-43b7-b043-c4c805dc37b9
type: credential.devops.kubesphere.io/kubeconfig
    `)))
	var buf strings.Builder
	if secret.Type == devops.CredentialTypeUsernamePassword{
		if err := basic_auth.Execute(&buf, secret); err != nil {
			return err
		}
	}else if secret.Type == devops.CredentialTypeSsh{
		if err := ssh_auth.Execute(&buf, secret); err != nil {
			return err
		}
	}else if secret.Type == devops.CredentialTypeSecretText{
		if err := secret_text.Execute(&buf, secret); err != nil {
			return err
		}
	}else if secret.Type == devops.CredentialTypeKubeConfig{
		if err := kubeconfig.Execute(&buf, secret); err != nil {
			return err
		}
	}else {
		return nil
	}
	var path = "./" + DevOpsDir + "/" + project + "/credential/"
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	err := ioutil.WriteFile(fmt.Sprintf( path +  filename + ".yaml"), []byte(buf.String()), 0644)
	if err != nil {
		return err
	}
	return nil
}




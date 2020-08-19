package devops

import (
	"fmt"
	"github.com/lithammer/dedent"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"os"
	"reflect"
	"regexp"
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

func GenerateDevOpsProjectYaml(filename string, creator string, showName string, desc string, workspaceName string) error {

	workspace, err := informers.KsSharedInformerFactory().Tenant().V1alpha1().Workspaces().Lister().Get(workspaceName)
	if err != nil {
		return err
	}

	var tmpl = template.Must(template.New(filename).Funcs(tf).Parse(
		dedent.Dedent(`---
apiVersion: devops.kubesphere.io/v1alpha3
kind: DevOpsProject
metadata:
  annotations:
    kubesphere.io/creator: "{{ .creator }}"
    kubesphere.io/workspace: "{{ .workspace }}"
    kubesphere.io/description: "{{ .desc }}"
  finalizers:
  - devopsproject.finalizers.kubesphere.io
  labels:
    kubesphere.io/workspace: "{{ .workspace }}"
  name: "{{ getValidName .filename }}"
  generateName: "{{ .name }}"
    `)))
	var buf strings.Builder
	variables := map[string]string{
		"workspace":  workspace.Name,
		"apiVersion": workspace.APIVersion,
		"uid":        string(workspace.UID),
		"filename":   filename,
		"creator":    creator,
		"name":       showName,
		"desc":       desc,
	}
	if err := tmpl.Execute(&buf, variables); err != nil {
		return err
	}
	var path = "./" + DevOpsDir + "/" + filename + "/"
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf(path+filename+".yaml"), []byte(buf.String()), 0644)
	if err != nil {
		return err
	}
	return nil
}

func GeneratePipelineYaml(project string, filename string, pipeline devops.ProjectPipeline, creator string) error {
	var pipelineTmpl = template.Must(template.New(filename).Funcs(tf).Parse(
		dedent.Dedent(`---
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  annotations:
    kubesphere.io/creator: "{{ .creator }}"
  finalizers:
  - pipeline.finalizers.kubesphere.io
  name: "{{ getValidName .Pipeline.Name}}"
  namespace: "{{ getValidName .Namespace}}"
spec:
  type: pipeline
  pipeline:
    {{ getYaml .Pipeline }}
status: {}
    `)))

	var multiBranchpipelineTmpl = template.Must(template.New(filename).Funcs(tf).Parse(
		dedent.Dedent(`---
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  annotations:
    kubesphere.io/creator: "{{ .creator }}"
  finalizers:
  - pipeline.finalizers.kubesphere.io
  name: "{{ getValidName .Pipeline.Name}}"
  namespace: "{{ getValidName .Namespace}}"
spec:
  type: multi-branch-pipeline
  multi_branch_pipeline:
    {{ getYaml .Pipeline }}
status: {}
    `)))
	var buf strings.Builder
	if pipeline.Type == "multi-branch-pipeline" && pipeline.MultiBranchPipeline != nil {
		data := map[string]interface{}{
			"Pipeline":  *pipeline.MultiBranchPipeline,
			"Namespace": project,
			"creator": creator,
		}
		if err := multiBranchpipelineTmpl.Execute(&buf, data); err != nil {
			DevOpsLogger().Println("Pipeline: %s, Exception: %v", pipeline, err)
			return err
		}
	} else if pipeline.Type == "pipeline" && pipeline.Pipeline != nil {
		data := map[string]interface{}{
			"Pipeline":  *pipeline.Pipeline,
			"Namespace": project,
			"creator": creator,
		}
		if err := pipelineTmpl.Execute(&buf, data); err != nil {
			DevOpsLogger().Println("Pipeline: %s, Exception: %v", pipeline, err)
			return err
		}
	} else {
		return nil
	}
	var path = "./" + DevOpsDir + "/" + project + "/pipeline/"
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	err := ioutil.WriteFile(fmt.Sprintf(path+filename+".yaml"), []byte(buf.String()), 0644)
	if err != nil {
		return err
	}
	return nil
}

func GenerateSecretYaml(project string, filename string, secret *devops.JenkinsCredential) error {
	var basic_auth = template.Must(template.New(filename).Funcs(tf).Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  password: ""
  username: ""
kind: Secret
metadata:
  annotations:
    kubesphere.io/creator: "{{.Creator}}"
    kubesphere.io/description: "{{.Description}}"
  finalizers:
  - finalizers.kubesphere.io/credential
  labels:
    app: "{{getValidName .Id}}"
  name: "{{getValidName .Id}}"
  namespace: "{{getValidName .Namespace}}"
type: credential.devops.kubesphere.io/basic-auth
    `)))

	var ssh_auth = template.Must(template.New(filename).Funcs(tf).Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  passphrase: ""
  private_key: ""
  username: ""
kind: Secret
metadata:
  annotations:
    kubesphere.io/creator: "{{.Creator}}"
    kubesphere.io/description: "{{.Description}}"
  finalizers:
  - finalizers.kubesphere.io/credential
  labels:
    app: "{{ getValidName .Id}}"
  name: "{{ getValidName .Id}}"
  namespace: "{{ getValidName .Namespace}}"
type: credential.devops.kubesphere.io/ssh-auth
    `)))

	var secret_text = template.Must(template.New(filename).Funcs(tf).Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  secret: ""
kind: Secret
metadata:
  annotations:
    kubesphere.io/creator: "{{.Creator}}"
    kubesphere.io/description: "{{.Description}}"
  finalizers:
  - finalizers.kubesphere.io/credential
  labels:
    app: "{{ getValidName .Id }}"
  name: "{{ getValidName .Id }}"
  namespace: "{{ getValidName .Namespace }}"
type: credential.devops.kubesphere.io/secret-text
    `)))
	var kubeconfig = template.Must(template.New(filename).Funcs(tf).Parse(
		dedent.Dedent(`---
apiVersion: v1
data:
  content: ""
kind: Secret
metadata:
  annotations:
    kubesphere.io/creator: "{{.Creator}}"
    kubesphere.io/description: "{{.Description}}"
  finalizers:
  - finalizers.kubesphere.io/credential
  labels:
    app: "{{ getValidName .Id }}"
  name: "{{ getValidName .Id }}"
  namespace: "{{ getValidName .Namespace }}"
type: credential.devops.kubesphere.io/kubeconfig
    `)))
	var buf strings.Builder
	data := map[string]string{
		"Id":          secret.Id,
		"Description": secret.Description,
		"Creator":     secret.Creator,
		"Namespace":   project,
	}
	if secret.Type == devops.CredentialTypeUsernamePassword {
		if err := basic_auth.Execute(&buf, data); err != nil {
			return err
		}
	} else if secret.Type == devops.CredentialTypeSsh {
		if err := ssh_auth.Execute(&buf, data); err != nil {
			return err
		}
	} else if secret.Type == devops.CredentialTypeSecretText {
		if err := secret_text.Execute(&buf, data); err != nil {
			return err
		}
	} else if secret.Type == devops.CredentialTypeKubeConfig {
		if err := kubeconfig.Execute(&buf, data); err != nil {
			return err
		}
	} else {
		return nil
	}
	var path = "./" + DevOpsDir + "/" + project + "/credential/"
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	err := ioutil.WriteFile(fmt.Sprintf(path+filename+".yaml"), []byte(buf.String()), 0644)
	if err != nil {
		return err
	}
	return nil
}

var tf = template.FuncMap{
	"isInt": func(str string) bool {
		pattern := "\\d+"
		result, _ := regexp.MatchString(pattern, str)
		return result
	},
	"isProperty": func(obj interface{}, property string) bool {
		t := reflect.TypeOf(obj)
		if t.String() == "Ptr" {
			t = reflect.TypeOf(obj)
		}
		if _, ok := t.FieldByName(property); ok {
			return true
		} else {
			return false
		}
	},
	"getYaml": func(obj interface{}) string {
		var pyaml, err = yaml.Marshal(obj)
		if err != nil {
			DevOpsLogger().Println("%v", err)
			return ""
		} else {
			return replaceKey(string(pyaml))
		}
	},
	"getValidName": func(old string) string {
		//return old
		return GetVaildName(old)
	},
}

func replaceKey(old string) string {
	var new = old
	var replaceList = map[string]string{
		"\n": "\n    ",
		// "null": "{}",
		"disableConcurrent":           "disable_concurrent",
		"timertrigger":                "timer_trigger",
		"remotetrigger":               "remote_trigger",
		"gitsource":                   "git_source",
		"githubsource":                "github_source",
		"svnsource":                   "svn_source",
		"singlesvnsource":             "single_svn_source",
		"singlesvn_source":             "single_svn_source",
		"bitbucketserversource":       "bitbucket_server_source",
		"scriptpath":                  "script_path",
		"multibranchjobtrigger":       "multibranch_job_trigger",
		"scmid":                       "scm_id",
		"credentialid":                "credential_id",
		"discoverbranches":            "discover_branches",
		"gitcloneoption":              "git_clone_option",
		"regexfilter":                 "regex_filter",
		"apiuri":                      "api_uri",
		"discoverprfromorigin":        "discover_pr_from_origin",
		"discoverprfromforks":         "discover_pr_from_forks",
		"gitclone_option":             "git_clone_option",
		"createactionjobtotrigger": "create_action_job_to_trigger",
		"deleteactionjobtotrigger": "delete_action_job_to_trigger",
		"daystokeep":                  "days_to_keep",
		"numtokeep":                   "num_to_keep",
		"defaultvalue":                "default_value",
		"sourcetype":                  "source_type",
		"cloneoption":                 "git_clone_option",
	}
	for key, value := range replaceList {
		new = strings.Replace(new, key, value, -1)
	}
	string_slice := strings.Split(new, "\n")
	string_target := []string{}
	for _, item := range string_slice {
		if !strings.Contains(item, "null") {
			string_target = append(string_target, item)
		}
	}
	return strings.Join(string_target, "\n")
}

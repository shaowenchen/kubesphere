package devops

import (
	"encoding/base64"
	"fmt"
	"kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/devops"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
	"net/url"
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

func QuerySecret(project string, doamin string)([]*devops.JenkinsCredential, error){
	return devops.GetProjectCredentials(project, doamin)
}

func AddBasicRequest(req *http.Request) *http.Request{
	var optoion = cs.ClientSets().GetOption()
	var devopsOption = optoion.GetDevopsOptions()

	creds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", devopsOption.Username, devopsOption.Password)))

	var header = req.Header
	if header == nil{
		header = make(map[string][]string)
	}
	header.Set("Authorization", fmt.Sprintf("Basic %s", creds))
	req.Header = header
	return req
}

func QueryPipelineList(project string)([]byte, error){
	var req http.Request
	var url = url.URL{RawQuery:"q=type:pipeline;organization:jenkins;pipeline:" + project + "%2F*;excludedFromFlattening:jenkins.branch.MultiBranchProject,hudson.matrix.MatrixProject&filter=no-folders&start=0&limit=9999"}
    req.URL = &url
	AddBasicRequest(&req)
	return devops.SearchPipelines(&req)
}
package devops

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/devops"
	apiserverconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"os"
	"path/filepath"
)

const (
	LogLevel = 1
	DevOpsDir = "devops_data"
)

func NewDevOpsCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "DevOps",
		Short: "Upgrade DevOps",
		Run: func(cmd *cobra.Command, args []string) {
			_ = apiserverconfig.Load()
			csop := &client.ClientSetOptions{}
			conf := apiserverconfig.Get()
			csop.SetDevopsOptions(conf.DevopsOptions).
				SetSonarQubeOptions(conf.SonarQubeOptions).
				SetKubernetesOptions(conf.KubernetesOptions).
				SetMySQLOptions(conf.MySQLOptions).
				SetLdapOptions(conf.LdapOptions).
				SetS3Options(conf.S3Options).
				SetOpenPitrixOptions(conf.OpenPitrixOptions).
				SetPrometheusOptions(conf.MonitoringOptions).
				SetKubeSphereOptions(conf.KubeSphereOptions).
				SetElasticSearchOptions(conf.LoggingOptions)

			client.NewClientSetFactory(csop, nil)
			klog.V(LogLevel).Info("start upgrade devops")
			upgradeDevOps()
			klog.V(LogLevel).Info("end upgrade devops")
		},
	}

	return cmd
}

func upgradeDevOps()  {

	// query devops
	projects, err := QueryDevops()
	if err != nil {
		klog.V(LogLevel).Info(fmt.Println("start upgrade devops, %s", err))
	}

	// create data dir

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		DevOpsLogger().Println("create dir error")
		return
	}
	dataDir := fmt.Sprintf("%s/%s", currentDir, DevOpsDir)
	CreateDir(dataDir)

	// query devops
    for _, project := range projects {
		GenerateDevOpsProjectYaml(project.ProjectId, nil)
		DevOpsLogger().Println(project.ProjectId)
    	pipelinesByte, err := QueryPipelineList(project.ProjectId)
    	if err != nil{
    		continue
		}
		type Pipelines struct {
			Items []devops.Pipeline `json:"items"`
			Total int               `json:"total_count"`
		}
		var pipelineList Pipelines
		err = json.Unmarshal(pipelinesByte, &pipelineList)
        if err != nil{
        	DevOpsLogger().Println(err)
        	continue
		}
		// query pipeline
		for _, pipeline := range pipelineList.Items{
			pipelineObj, err := devops.GetProjectPipeline(project.ProjectId, pipeline.Name)
			if err == nil{
				GeneratePipelineYaml(project.ProjectId, pipeline.Name, *pipelineObj)
			}
		}
		// query secret
		secretList, err := QuerySecret(project.ProjectId, "_")
		for _, secret := range secretList{
			GenerateSecretYaml(project.ProjectId, secret.DisplayName, secret)
		}

	}
}

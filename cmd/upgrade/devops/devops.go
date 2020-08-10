package devops

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/cmd/ks-apiserver/app"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/devops"
	apiserverconfig "kubesphere.io/kubesphere/pkg/server/config"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/signals"
	"os/exec"
	"sync"
	"time"
)

const (
	LogLevel  = 1
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
			err := app.WaitForResourceSync(signals.SetupSignalHandler())
			if err != nil {
			} else {
				upgradeDevOps()
			}
		},
	}

	return cmd
}

func upgradeDevOps() {

	// query devops
	projects, err := QueryDevops()
	if err != nil {
		klog.V(LogLevel).Info(fmt.Println("start upgrade devops, %s", err))
	}

	// create data dir

	//currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	//if err != nil {
	//	DevOpsLogger().Println("create dir error")
	//	return
	//}
	//dataDir := fmt.Sprintf("%s/%s", currentDir, DevOpsDir)
	//CreateDir(dataDir)

	// query devops
	DevOpsLogger().Println("Start Query Old Data from DB and Jenkins")
	for _, project := range projects {
		GenerateDevOpsProjectYaml(project.ProjectId, project.Workspace)
		DevOpsLogger().Println("Current DevOps Project: ", project.ProjectId)
		pipelinesByte, err := QueryPipelineList(project.ProjectId)
		if err != nil {
			continue
		}
		type Pipelines struct {
			Items []devops.Pipeline `json:"items"`
			Total int               `json:"total_count"`
		}
		var pipelineList Pipelines
		err = json.Unmarshal(pipelinesByte, &pipelineList)
		if err != nil {
			DevOpsLogger().Println(err)
			continue
		}
		// query pipeline
		for _, pipeline := range pipelineList.Items {
			DevOpsLogger().Println("Current Pipeline: ", pipeline.Name)
			pipelineObj, err := devops.GetProjectPipeline(project.ProjectId, pipeline.Name)
			if err == nil {
				GeneratePipelineYaml(project.ProjectId, pipeline.Name, *pipelineObj)
			}
		}
		// query secret
		secretList, err := QuerySecret(project.ProjectId, "_")
		for _, secret := range secretList {
			DevOpsLogger().Println("Current Secret: ", secret.Id)
			GenerateSecretYaml(project.ProjectId, secret.Id, secret)
		}

	}
	DevOpsLogger().Println("End Query Old Data from DB and Jenkins")


	// backup data
	DevOpsLogger().Println("Start Upload to S3")
	uploadDir(fmt.Sprintf("./%s", DevOpsDir))
	DevOpsLogger().Println("End Upload to S3")


	// upgrade
	DevOpsLogger().Println("Start Upgrade to 3.0")
	projectItems, err := GetDevOps(DevOpsDir)
	if err != nil{
		DevOpsLogger().Println(err)
		return
	}

	// init sync wait
	var wg sync.WaitGroup
	wg.Add(len(projectItems))

	for _, project := range projectItems{
		CreateDevOpsAndWaitNamespaces(project)
		go func() {
			defer wg.Done()
			pipelines, err := GetSubDirFiles(project.ProjectDir, "pipeline")

			if err != nil{
				for _, pipeline := range pipelines{
					CreatePipeline(pipeline)
				}
			}else{
				DevOpsLogger().Println(err)
			}

			credentials, err := GetSubDirFiles(project.ProjectDir, "credential")

			if err != nil{
				for _, credential := range credentials{
					CreateSecret(credential)
				}
			}else{
				DevOpsLogger().Println(err)
			}
		}()
	}
	DevOpsLogger().Println("End upgrade 3.0")

	// upgrade iam
	for _, item := range GetDevOpsIm() {
		DevOpsLogger().Println(*item)
	}
}

func CreateDevOpsAndWaitNamespaces(proj ProjectItem)  {
	DevOpsLogger().Println("Apply DevOps: ", proj.ProjectPath)
	KubectlApply(proj.ProjectPath)
	for{
		_, err := informers.SharedInformerFactory().Core().V1().Namespaces().Lister().Get(GetVaildName(proj.NameSpace))
		if err == nil{
			DevOpsLogger().Println("Success Namespace Create:", proj.NameSpace)
			break
		}else{
			time.Sleep(2 *time.Second)
			DevOpsLogger().Println("Wait Namespace Create:", proj.NameSpace)
		}
	}
}

func CreatePipeline(file string) {
	DevOpsLogger().Println("Apply Pipeline: ", file)
	KubectlApply(file)
}

func CreateSecret(file string) {
	DevOpsLogger().Println("Apply Secret: ", file)
	KubectlApply(file)
}

func KubectlApply(file string)error{
	cmd := exec.Command("/bin/sh", "-c", "kubectl apply -f " + file)
	stdout, err := cmd.Output()

	if err != nil {
		DevOpsLogger().Println(err)
		return err
	}

	DevOpsLogger().Println(string(stdout))
	return nil
}

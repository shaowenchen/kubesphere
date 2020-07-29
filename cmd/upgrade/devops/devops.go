package devops

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/klog"
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

	// query pipeline and save
    for _, project := range projects {
    	klog.V(LogLevel).Info(project.ProjectId)
		DevOpsLogger().Println(project.ProjectId)
    	pipelines := fmt.Sprintf()
	}

}

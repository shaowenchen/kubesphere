package devops

import (
	"bytes"
	"code.cloudfoundry.org/bytefmt"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"kubesphere.io/kubesphere/pkg/simple/client"
)

func uploadDir(path string) error {
	s3Client, err := client.ClientSets().S3()
	s3session := s3Client.Session()
	if s3session == nil {
		return err
	}
	timeNow := time.Now()
	timeString := timeNow.Format("2006-01-02 15:04:05")

	//files, dirs, _ := GetFilesAndDirs(path)
	//
	//for _, dir := range dirs {
	//	fmt.Printf("获取的文件夹为[%s]\n", dir)
	//}
	//
	//for _, table := range dirs {
	//	temp, _, _ := GetFilesAndDirs(table)
	//	for _, temp1 := range temp {
	//		files = append(files, temp1)
	//	}
	//}
	//
	//for _, table1 := range files {
	//	fmt.Printf("获取的文件为[%s]\n", table1)
	//}

	files, _ := GetAllFiles(path)
	for _, file := range files {
		fmt.Printf("Find [%s]\n", file)
	}
	for _, filepath := range files {
		// Open the file for use
		file, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Get file size and read the file content into a buffer
		fileInfo, _ := file.Stat()
		var size int64 = fileInfo.Size()
		buffer := make([]byte, size)
		file.Read(buffer)

		// Config settings: this is where you choose the bucket, filename, content-type etc.
		// of the file you're uploading.
		uploader := s3manager.NewUploader(s3session, func(uploader *s3manager.Uploader) {
			uploader.PartSize = 5 * bytefmt.MEGABYTE
			uploader.LeavePartsOnError = true
		})
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:             s3Client.Bucket(),
			Key:                aws.String(fmt.Sprintf("%s/%s", timeString, filepath)),
			Body:               bytes.NewReader(buffer),
			ContentDisposition: aws.String("attachment"),
		})
	}
	return err
}

func GetFilesAndDirs(dirPth string) (files []string, dirs []string, err error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, nil, err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			GetFilesAndDirs(dirPth + PthSep + fi.Name())
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), ".yaml")
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	return files, dirs, nil
}

func GetAllFiles(dirPth string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			GetAllFiles(dirPth + PthSep + fi.Name())
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), ".yaml")
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp, _ := GetAllFiles(table)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, nil
}

type ProjectItem struct {
	ProjectPath string
	ProjectDir  string
	NameSpace   string
}

// ./devops_data
func GetDevOps(dirPth string) ([]ProjectItem, error) {
	var project []ProjectItem
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)

	for _, fi := range dir {
		if fi.IsDir() {
			project = append(project, ProjectItem{
				ProjectPath: dirPth + PthSep + fi.Name() + PthSep + fi.Name() + ".yaml",
				ProjectDir:  dirPth + PthSep + fi.Name(),
				NameSpace:   fi.Name(),
			})
		}
	}
	return project, nil
}

// ./devops_data/project-xxxxxxx
func GetSubDirFiles(disPth string, sub string) ([]string, error) {
	PthSep := string(os.PathSeparator)

	if _, err := os.Stat(disPth + PthSep + sub); os.IsNotExist(err) {
		return nil, err
	}

	var files []string
	dir, err := ioutil.ReadDir(disPth + PthSep + sub)
	if err != nil {
		return nil, err
	}

	for _, fi := range dir {
		if fi.IsDir() {
			continue
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), ".yaml")
			if ok {
				files = append(files, disPth+PthSep+sub+PthSep+fi.Name())
			}
		}
	}

	return files, nil
}

func GetVaildName(old string) string {
	//return strings.ToLower(strings.Replace(old, "project-", "", -1))
	return strings.ToLower(old)
}

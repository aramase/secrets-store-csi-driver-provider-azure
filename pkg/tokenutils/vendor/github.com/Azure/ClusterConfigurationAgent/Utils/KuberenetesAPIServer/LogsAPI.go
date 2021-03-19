package KuberenetesAPIServer

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"time"
)

func (client *KubeAPIServerQueries)GetLogsForPartnerPods(selector string, containerName string, sinceTime time.Time) string {

	// 64 KBytes in every call
	limitBytes := int64(64 * 1024)
	logOptions := &v1.PodLogOptions{
		LimitBytes: &limitBytes,
	}

	logOptions.SinceTime = &metav1.Time{
		Time: sinceTime,
	}

	if containerName != "" {
		logOptions.Container = containerName
	}

	logs, err := client.GetLogs(selector, logOptions)
	if err != nil{
		log.Printf("Error while collecting logs %v", err)
		return ""
	}

	return logs
}
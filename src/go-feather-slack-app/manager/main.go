/**
* File              : main.go
* Author            : Alexandre Saison <alexandre.saison@inarix.com>
* Date              : 09.12.2020
* Last Modified Date: 22.01.2021
* Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package podManager

import (
	"errors"
	"fmt"
	"log"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func New(inCluster bool) *PodManager {

	var config *rest.Config
	var err error

	if inCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			log.Panicln(err.Error())
		}

	}

	// creates the clientset
	clientset := kubernetes.NewForConfigOrDie(config)
	return &PodManager{client: clientset}
}

func (self *PodManager) CreateConfigRefSpec(configMapRefsNames []string) []v1.ConfigMapEnvSource {
	configMapRefs := make([]v1.ConfigMapEnvSource, len(configMapRefsNames))
	for index, configMapName := range configMapRefsNames {
		isOptionnal := true
		tmpConfigMapRef := &v1.ConfigMapEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: configMapName}, Optional: &isOptionnal}
		configMapRefs[index] = *tmpConfigMapRef
	}
	return configMapRefs
}

func (self *PodManager) CreateEnvsRefSpec(envMapRefsValues map[string]string) []v1.EnvVar {
	envMapRefs := make([]v1.EnvVar, len(envMapRefsValues))
	indexMap := 0
	for key, value := range envMapRefsValues {
		tmpEnvVariableRef := &v1.EnvVar{Name: key, Value: value}
		envMapRefs[indexMap] = *tmpEnvVariableRef
		indexMap++
	}
	return envMapRefs
}

func (self *PodManager) fetchPodNameFromJobName(namespace string, jobName string) (*v1.Pod, error) {
	labelSelector := "job-name=" + jobName
	pods, err := self.client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 1 {
		return &pods.Items[0], nil
	}

	errorMessage := fmt.Sprintf("No pod found with job-name=%s on namespace %s", jobName, namespace)
	return nil, errors.New(errorMessage)
}

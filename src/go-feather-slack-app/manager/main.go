/**
* File              : main.go
* Author            : Alexandre Saison <alexandre.saison@inarix.com>
* Date              : 09.12.2020
* Last Modified Date: 28.12.2020
* Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package podManager

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type PodManager struct {
	client *kubernetes.Clientset
}

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
		isOptionnal := false
		tmpConfigMapRef := &v1.ConfigMapEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: configMapName}, Optional: &isOptionnal}
		configMapRefs[index] = *tmpConfigMapRef
	}
	return configMapRefs
}

func (self *PodManager) GetPods(namespace string) (*v1.PodList, error) {
	return self.client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

func (self *PodManager) GetPod(namespace string, podName string) (*v1.Pod, error) {
	return self.client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
}

func (self *PodManager) CreateJob(namespace string, prefixName string, jobSpec batchv1.JobSpec) error {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefixName,
			Namespace:    namespace,
		},
		Spec: jobSpec,
	}
	job, err := self.client.BatchV1().Jobs(namespace).Create(job)

	if err != nil {
		return err
	}

	log.Printf("Job %s has been created successfuly", job.GetName())
	pod, err := self.fetchPodNameFromJobName(namespace, job.GetName())
	if err != nil {
		return err
	}
	logs, err := self.GetPodLogs(namespace, pod.GetName(), false)
	if err != nil {
		return err
	}
	log.Println("found logs", logs)
	return nil
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

func (self *PodManager) CreateJobSpec(jobNamePrefix string, containerName string, containerImage string, envs []v1.EnvVar, configMapRefs []v1.ConfigMapEnvSource) *batchv1.JobSpec {
	jobSpec := &batchv1.JobSpec{
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: jobNamePrefix,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  containerName,
						Image: containerImage,
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
			},
		},
	}
	if envs != nil || len(envs) > 0 {
		log.Printf("Adding %d environment variable to the container %s", len(envs), containerName)
		jobSpec.Template.Spec.Containers[0].Env = envs
	}

	if configMapRefs != nil || len(configMapRefs) > 0 {
		log.Printf("Adding %d configMapRefs to the container %s", len(configMapRefs), containerName)
		envFrom := make([]v1.EnvFromSource, len(configMapRefs))
		for index := range configMapRefs {
			envSource := v1.EnvFromSource{ConfigMapRef: &configMapRefs[index]}
			envFrom[index] = envSource
		}
		jobSpec.Template.Spec.Containers[0].EnvFrom = envFrom
	}

	return jobSpec
}

func (self *PodManager) GetPodLogs(namespace string, podName string, async bool) (string, error) {
	podLogOpts := v1.PodLogOptions{Follow: true}
	req := self.client.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	reader, err := req.Stream()
	if err != nil {
		panic(err.Error())
	}
	defer reader.Close()
	body, err := ioutil.ReadAll(reader)

	if err != nil {
		log.Println("POD_READING_FAILURE : ", err.Error())
		return "", errors.New("An error occured during reading pod logs, watch over server pod logs for more informations")
	}
	log.Println("Logs are : ", string(body))
	return string(body), nil
}

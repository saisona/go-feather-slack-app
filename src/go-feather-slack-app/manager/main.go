/**
 * File              : main.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 09.12.2020
 * Last Modified Date: 16.12.2020
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package podManager

import (
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

	if inCluster {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		return &PodManager{client: clientset}
	} else {
		config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			log.Panicln(err.Error())
		}

		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Panicln(err.Error())
		}
		return &PodManager{client: clientset}
	}
}

func (self *PodManager) GetPods(namespace string) (*v1.PodList, error) {
	return self.client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

func (self *PodManager) GetPod(namespace string, podName string) (*v1.Pod, error) {
	return self.client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
}

func (self *PodManager) CreateJob(namespace string, prefixName string, jobSpec batchv1.JobSpec) (bool, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefixName,
			Namespace:    namespace,
		},
		Spec: jobSpec,
	}
	job, err := self.client.BatchV1().Jobs(namespace).Create(job)

	if err != nil {
		return false, err
	}
	log.Printf("Job %s has been created successfuly", job.GetName())
	return true, nil
}

func (self *PodManager) CreateJobSpecWithEnvVariable(jobNamePrefix string, containerName string, containerImage string, envs []v1.EnvVar) *batchv1.JobSpec {
	return &batchv1.JobSpec{
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: jobNamePrefix,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  containerName,
						Image: containerImage,
						Env:   envs,
					},
				},

				RestartPolicy: v1.RestartPolicyNever,
			},
		},
	}
}

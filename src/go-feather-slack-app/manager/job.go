/**
 * File              : job.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 29.12.2020
 * Last Modified Date: 23.01.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package podManager

import (
	"log"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *PodManager) DeleteJob(namespace string, jobName string) error {
	log.Printf("Deleteing job %s on namespace %s", jobName, namespace)
	if err := self.client.BatchV1().Jobs(namespace).Delete(jobName, &metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
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

func (self *PodManager) CreateJob(namespace string, prefixName string, jobSpec batchv1.JobSpec) (*v1.Pod, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefixName,
			Namespace:    namespace,
		},
		Spec: jobSpec,
	}
	job, err := self.client.BatchV1().Jobs(namespace).Create(job)
	time.Sleep(250 * time.Millisecond)

	if err != nil {
		return nil, err
	}
	log.Printf("Job %s has been created successfuly", job.GetName())
	pod, err := self.fetchPodNameFromJobName(namespace, job.GetName())
	if err != nil {
		return nil, err
	}
	return pod, nil
}

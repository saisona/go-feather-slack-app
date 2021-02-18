/**
 * File              : pod.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 29.12.2020
 * Last Modified Date: 18.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package podManager

import (
	"errors"
	"io/ioutil"
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (self *PodManager) DeletePod(namespace string, podName string) error {
	if err := self.client.CoreV1().Pods(namespace).Delete(podName, metav1.NewDeleteOptions(5)); err != nil {
		return err
	}
	return nil
}

func (self *PodManager) GetPods(namespace string) (*v1.PodList, error) {
	return self.client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

func (self *PodManager) GetPod(namespace string, podName string) (*v1.Pod, error) {
	return self.client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
}

// GetPodLogs: use namespace and podName args to fetch logs of an ended pod.
// Most of the time, it is used for Jobs since waits for pod to be completed.
//@args namespace: Namespace of the pod to watch for logs.
//@args podName: Name of the pod's logs to fetch on previously specified namespace.
//@returns (string, string, error):
// string -> returns the logs of the ended pod.
// string -> returns last post status (Completed/Error/Oom ...)
// error -> any error from kubernetes api.
func (self *PodManager) GetPodLogs(namespace string, podName string) (string, string, error) {
	podLogOpts := v1.PodLogOptions{}
	log.Printf("Getting logs from %s in namespace %s", podName, namespace)
	pod, err := self.GetPod(namespace, podName)

	if err != nil {
		return "", "", err
	}

	podPhase, err := self.WaitForPodReady(namespace, pod)
	if err != nil {
		log.Printf("Error while waiting for pod readiness : %s", err.Error())
		return "", "", err
	}

	req := self.client.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	reader, err := req.Stream()

	if err != nil {
		return "", pod.Status.Reason, err
	}

	defer reader.Close()
	body, err := ioutil.ReadAll(reader)

	if err != nil {
		log.Println("POD_READING_FAILURE : ", err.Error())
		return "", pod.Status.Reason, errors.New("An error occured during reading pod logs, watch over server pod logs for more informations")
	}

	return string(body), podPhase, nil
}

func (self *PodManager) WaitForPodReady(namespace string, pod *v1.Pod) (string, error) {
	watcher, err := self.client.CoreV1().Pods(namespace).Watch(metav1.SingleObject(metav1.ObjectMeta{Namespace: namespace, Name: pod.GetName()}))
	if err != nil {
		return "", err
	}

	podPhase, err := DefaultHandlerWaitingFunc(watcher, pod)
	if err != nil {
		return "", err
	}

	return podPhase, nil
}

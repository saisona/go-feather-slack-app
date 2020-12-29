/**
 * File              : pod.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 29.12.2020
 * Last Modified Date: 29.12.2020
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

func (self *PodManager) GetPodLogs(namespace string, pod *v1.Pod) (string, error) {
	podLogOpts := v1.PodLogOptions{}
	podName := pod.GetName()
	log.Printf("Getting logs from %s in namespace %s", podName, namespace)
	if err := self.WaitForPodReady(namespace, pod); err != nil {
		log.Printf("Error while waiting for pod readiness : %s", err.Error())
	}
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
	return string(body), nil
}

func (self *PodManager) WaitForPodReady(namespace string, pod *v1.Pod) error {
	watcher, err := self.client.CoreV1().Pods(namespace).Watch(metav1.SingleObject(metav1.ObjectMeta{Namespace: namespace, Name: pod.GetName()}))
	if err != nil {
		return err
	}

	if err := DefaultHandlerWaitingFunc(watcher, pod); err != nil {
		return err
	}

	return nil
}

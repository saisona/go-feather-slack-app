/**
 * File              : helpers.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 28.12.2020
 * Last Modified Date: 28.12.2020
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package podManager

import (
	"errors"
	"log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type PodManager struct {
	client *kubernetes.Clientset
}

type HandlerWaitingFunc func(watcher watch.Interface, pod *v1.Pod) error

func DefaultHandlerWaitingFunc(watcher watch.Interface, pod *v1.Pod) error {
	for event := range watcher.ResultChan() {
		p, ok := event.Object.(*v1.Pod)
		if !ok {
			return errors.New("Unexpected type for *v1.Pod whithin watcher event loop")
		}
		log.Printf("Pod %s is in state %s", p.GetName(), string(p.Status.Phase))
		phase := string(p.Status.Phase)
		if phase == "Succeeded" || phase == "Error" {
			watcher.Stop()
		}
	}
	return nil
}

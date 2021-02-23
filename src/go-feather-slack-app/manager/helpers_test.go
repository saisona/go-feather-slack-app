/**
 * File              : helpers_test.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 23.02.2021
 * Last Modified Date: 23.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package podManager

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TestDefaultHandlerWaitingFuncNoEvents(t *testing.T) {
	fakeWatcher := watch.NewFake()
	fakePod := &v1.Pod{}
	fakeWatcher.Stop()
	podStatus, err := DefaultHandlerWaitingFunc(fakeWatcher, fakePod)
	if err != nil || podStatus != "" {
		t.Errorf("Failed as not supposed to receive any events before running")
	}
	t.Logf("%s Succeeded", t.Name())
}

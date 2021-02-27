/**
 * File              : job_test.go
 * Author            : Alexandre Saison <alexandre.saison@inarix.com>
 * Date              : 23.02.2021
 * Last Modified Date: 26.02.2021
 * Last Modified By  : Alexandre Saison <alexandre.saison@inarix.com>
 */
package podManager

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	fakebatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	k8stesting "k8s.io/client-go/testing"
)

func NewFakePodManager() *PodManager {
	return &PodManager{client: fake.NewSimpleClientset()}
}

func TestDeleteJobNotExists(t *testing.T) {
	podManager := NewFakePodManager()
	podManager.client.BatchV1().(*fakebatchv1.FakeBatchV1).Fake.PrependReactor("delete", "batch/v1", func(a k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return false, nil, nil
	})
	err := podManager.DeleteJob("toto", "dummy-job")
	if err != nil {
		t.Logf("Succeed, wanted err=%s got=%s", "jobs.batch \"dummy-job\" not found", err.Error())
	} else {
		t.Errorf("Failed, wanted err=%s got=%s", "jobs.batch \"dummy-job\" not found", "nil")
	}
}

func TestDeleteJobSuccess(t *testing.T) {
	podManager := NewFakePodManager()
	podManager.client.BatchV1().(*fakebatchv1.FakeBatchV1).Fake.PrependReactor("delete", "batch/v1", func(a k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})
	err := podManager.DeleteJob("toto", "dummy-job")
	if err != nil {
		t.Errorf("Failed, wanted err=%s got=%s", "nil", err.Error())
	} else {
		t.Logf("Succeed, wanted err=%s got=%s", "nil", "nil")
	}
}

func TestCreateCreateJobSpecDefault(t *testing.T) {
	podManager := NewFakePodManager()
	createdJobSpec := podManager.CreateJobSpec("dummy", "dummy-job", "eu.gcr.io/toto/dummy-feather", nil, nil)
	if createdJobSpec == nil {
		t.Error("Failed !")
	}
}

func TestCreateCreateJobSpecWithEnv(t *testing.T) {
	podManager := NewFakePodManager()
	envMap := map[string]string{"DUMMY": "dummy_value"}
	v1EnvVars := podManager.CreateEnvsRefSpec(envMap)
	createdJobSpec := podManager.CreateJobSpec("dummy", "dummy-job", "eu.gcr.io/toto/dummy-feather", v1EnvVars, nil)
	specContainer := createdJobSpec.Template.Spec.Containers[0]
	if specContainer.Env[0].Name != "DUMMY" || specContainer.Env[0].Value != "dummy_value" {
		t.Errorf("Failed, wanted=%+v, got=%+v", v1EnvVars, specContainer.Env[0])
	}
}

func TestCreateCreateJobSpecWithConfigMap(t *testing.T) {
	podManager := NewFakePodManager()
	configMapNames := []string{"dummy-config"}
	v1ConfigMaps := podManager.CreateConfigRefSpec(configMapNames)
	createdJobSpec := podManager.CreateJobSpec("dummy", "dummy-job", "eu.gcr.io/toto/dummy-feather", nil, v1ConfigMaps)
	specContainer := createdJobSpec.Template.Spec.Containers[0]

	if specContainer.EnvFrom[0].ConfigMapRef.Name != v1ConfigMaps[0].Name {
		t.Errorf("Failed, wanted=%+v, got=%+v", v1ConfigMaps, specContainer.EnvFrom)
	}

}

func TestCreateCreateJobSpecWithEverything(t *testing.T) {
	podManager := NewFakePodManager()

	configMapNames := []string{"dummy-config"}
	v1ConfigMaps := podManager.CreateConfigRefSpec(configMapNames)

	envMap := map[string]string{"DUMMY": "dummy_value"}
	v1EnvVars := podManager.CreateEnvsRefSpec(envMap)

	createdJobSpec := podManager.CreateJobSpec("dummy", "dummy-job", "eu.gcr.io/toto/dummy-feather", v1EnvVars, v1ConfigMaps)
	specContainer := createdJobSpec.Template.Spec.Containers[0]

	if specContainer.EnvFrom[0].ConfigMapRef.Name != v1ConfigMaps[0].Name {
		t.Errorf("Failed, wanted=%+v, got=%+v", v1ConfigMaps, specContainer.EnvFrom)
	}

	if specContainer.Env[0].Name != "DUMMY" || specContainer.Env[0].Value != "dummy_value" {
		t.Errorf("Failed, wanted=%+v, got=%+v", v1EnvVars, specContainer.Env[0])
	}
}

func TestCreateCreateJobFailed(t *testing.T) {
	podManager := NewFakePodManager()
	fakeError := errors.New("dummy_error")
	podManager.client.CoreV1().(*fakecorev1.FakeCoreV1).Fake.PrependReactor("create", "pods", func(a k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		t.Log("!!!! ENTER CREATE POD!!!")
		t.Logf("action=%+v", a)
		return true, nil, nil
	})
	podManager.client.BatchV1().(*fakebatchv1.FakeBatchV1).Fake.PrependReactor("create", "jobs", func(a k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		t.Log("!!!! ENTER CREATE JOB!!!")
		t.Logf("action=%+v", a)
		podManager.client.CoreV1().(*fakecorev1.FakeCoreV1).Fake.PrependReactor("create", "pods", func(a k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			t.Log("!!!! ENTER CREATE POD inside !!!")
			t.Logf("action=%+v", a)
			return true, nil, nil
		})
		return false, nil, fakeError
	})

	jobSpec := podManager.CreateJobSpec("prefix", "job", "image", nil, nil)
	t.Log(jobSpec.Template.Name)
	_, err := podManager.CreateJob("toto", "dummy", *jobSpec)
	if err == nil || !errors.Is(err, fakeError) {
		t.Errorf("Failed, Should have failed with wanted=%s but got=%s", fakeError.Error(), err.Error())
	}
}

func TestCreateCreateJobSucceed(t *testing.T) {
	podManager := NewFakePodManager()
	podManager.client.BatchV1().(*fakebatchv1.FakeBatchV1).Fake.PrependReactor("create", "batch/v1", func(a k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})
	jobSpec := podManager.CreateJobSpec("prefix", "job", "image", nil, nil)
	_, err := podManager.CreateJob("toto", "dummy", *jobSpec)
	if err != nil {
		t.Errorf("Failed, wanted=%s, got=%s", "nil", err.Error())
	}
}

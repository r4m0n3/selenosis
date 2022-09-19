package platform

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	testcore "k8s.io/client-go/testing"

	"github.com/alcounit/selenosis/selenium"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
)

func TestErrorsOnServiceCreate(t *testing.T) {
	tests := map[string]struct {
		ns        string
		podName   string
		layout    ServiceSpec
		eventType watch.EventType
		podPhase  apiv1.PodPhase
		err       error
	}{
		"Verify platform error on pod startup phase PodSucceeded": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Added,
			podPhase:  apiv1.PodSucceeded,
			err:       errors.New("pod is not ready after creation: pod exited early with status Succeeded"),
		},
		"Verify platform error on pod startup phase PodFailed": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Added,
			podPhase:  apiv1.PodFailed,
			err:       errors.New("pod is not ready after creation: pod exited early with status Failed"),
		},
		"Verify platform error on pod startup phase PodUnknown": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Added,
			podPhase:  apiv1.PodUnknown,
			err:       errors.New("pod is not ready after creation: couldn't obtain pod state"),
		},
		"Verify platform error on pod startup phase Unknown": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Added,
			err:       errors.New("pod is not ready after creation: pod has unknown status"),
		},
		"Verify platform error on pod startup event Error": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Error,
			podPhase:  apiv1.PodUnknown,
			err:       errors.New("pod is not ready after creation: received error while watching pod: /, Kind="),
		},
		"Verify platform error on pod startup event Deleted": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Deleted,
			podPhase:  apiv1.PodUnknown,
			err:       errors.New("pod is not ready after creation: pod was deleted before becoming available"),
		},
		"Verify platform error on pod startup event Unknown": {
			ns:      "selenosis",
			podName: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
			eventType: watch.Bookmark,
			podPhase:  apiv1.PodSucceeded,
			err:       errors.New("pod is not ready after creation: received unknown event type BOOKMARK while watching pod"),
		},
	}

	for name, test := range tests {

		t.Logf("TC: %s", name)

		mock := fake.NewSimpleClientset()
		watcher := watch.NewFakeWithChanSize(1, false)
		mock.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
		watcher.Action(test.eventType, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: test.podName,
			},
			Status: apiv1.PodStatus{
				Phase: test.podPhase,
			},
		})

		client := &Client{
			ns:        test.ns,
			clientset: mock,
			service: &service{
				ns:        test.ns,
				clientset: mock,
			},
		}

		_, err := client.Service().Create(test.layout)

		assert.Equal(t, test.err.Error(), err.Error())
	}
}

func TestPodDelete(t *testing.T) {
	tests := map[string]struct {
		ns           string
		createPod    string
		deletePod    string
		browserImage string
		proxyImage   string
		err          error
	}{
		"Verify platform deletes running pod": {
			ns:           "selenosis",
			createPod:    "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			deletePod:    "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			browserImage: "selenoid/vnc:chrome_85.0",
			proxyImage:   "alcounit/seleniferous:latest",
		},
		"Verify platform delete return error": {
			ns:           "selenosis",
			createPod:    "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
			deletePod:    "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144912",
			browserImage: "selenoid/vnc:chrome_85.0",
			proxyImage:   "alcounit/seleniferous:latest",
			err:          errors.New(`pods "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144912" not found`),
		},
	}

	for name, test := range tests {

		t.Logf("TC: %s", name)

		mock := fake.NewSimpleClientset()

		client := &Client{
			ns:        test.ns,
			clientset: mock,
			service: &service{
				ns:        test.ns,
				clientset: mock,
			},
		}

		ctx := context.Background()
		_, err := mock.CoreV1().Pods(test.ns).Create(ctx, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: test.createPod,
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  "browser",
						Image: test.browserImage,
					},
					{
						Name:  "seleniferous",
						Image: test.proxyImage,
					},
				},
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create fake pod: %v", err)
		}

		err = client.Service().Delete(test.deletePod)

		if err != nil {
			assert.Equal(t, test.err.Error(), err.Error())
		} else {
			assert.Equal(t, test.err, err)
		}
	}

}

func TestListPods(t *testing.T) {
	tests := map[string]struct {
		ns       string
		svc      string
		podNames []string
		podData  []struct {
			podName   string
			podPhase  apiv1.PodPhase
			podStatus ServiceStatus
		}
		podPhase     []apiv1.PodPhase
		podStatus    []ServiceStatus
		labels       map[string]string
		browserImage string
		proxyImage   string
		err          error
	}{
		"Verify platform returns running pods that match selector": {
			ns:           "selenosis",
			svc:          "selenosis",
			podNames:     []string{"chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911", "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144912", "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144913"},
			podPhase:     []apiv1.PodPhase{apiv1.PodRunning, apiv1.PodPending, apiv1.PodFailed},
			podStatus:    []ServiceStatus{Running, Pending, Unknown},
			labels:       map[string]string{"selenosis.app.type": "browser"},
			browserImage: "selenoid/vnc:chrome_85.0",
			proxyImage:   "alcounit/seleniferous:latest",
		},
	}

	for name, test := range tests {

		t.Logf("TC: %s", name)

		mock := fake.NewSimpleClientset()
		client := &Client{
			ns:        test.ns,
			svc:       test.svc,
			svcPort:   intstr.FromString("4445"),
			clientset: mock,
		}

		for i, name := range test.podNames {
			ctx := context.Background()
			_, err := mock.CoreV1().Pods(test.ns).Create(ctx, &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:   name,
					Labels: test.labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "browser",
							Image: test.browserImage,
						},
						{
							Name:  "seleniferous",
							Image: test.proxyImage,
						},
					},
				},
				Status: apiv1.PodStatus{
					Phase: test.podPhase[i],
				},
			}, metav1.CreateOptions{})

			if err != nil {
				t.Logf("failed to create fake pod: %v", err)
			}
		}

		state, err := client.State()
		if err != nil {
			t.Fatalf("Failed to list pods %v", err)
		}

		for i, pod := range state.Services {
			assert.Equal(t, pod.SessionID, test.podNames[i])

			u := &url.URL{
				Scheme: "http",
				Host:   net.JoinHostPort(fmt.Sprintf("%s.%s", test.podNames[i], test.svc), "4445"),
			}

			assert.Equal(t, pod.URL.String(), u.String())

			assert.Equal(t, pod.Status, test.podStatus[i])
		}
	}
}

// TestSvcSuccessCreation verify creation of service session
func TestSvcSuccessCreation(t *testing.T) {
	tests := map[string]struct {
		ns      string
		svc     string
		layout  ServiceSpec
		podName string
		svcPort intstr.IntOrString
	}{
		"pod success creation": {
			podName: "chrome-100-0-screenResolution",
			layout: ServiceSpec{
				SessionID: "chrome-100-0-screenResolution",
			},
			ns:      "selenosis",
			svc:     "svc",
			svcPort: intstr.FromString("4444"),
		},
	}

	for name, test := range tests {
		t.Logf("TC: Verify %s", name)

		cliMock := fake.NewSimpleClientset()
		watcher := watch.NewFakeWithChanSize(1, false)
		cliMock.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
		watcher.Action(watch.Added, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: test.podName,
			},
			Status: apiv1.PodStatus{
				Phase: apiv1.PodRunning,
			},
		})
		client := &Client{
			ns:        test.ns,
			clientset: cliMock,
			service: &service{
				ns:             test.ns,
				svc:            test.svc,
				svcPort:        test.svcPort,
				clientset:      cliMock,
				waitForService: func(u url.URL, t time.Duration) error { return nil },
			},
		}

		svc, err := client.Service().Create(test.layout)
		assert.Nil(t, err)
		assert.Equal(t, svc.Status, Running)
		assert.Equal(t, svc.SessionID, test.podName)
		u := &url.URL{
			Scheme: "http",
			Host:   test.podName + "." + test.svc + ":" + browserPorts.selenium.StrVal,
		}
		assert.Equal(t, u, svc.URL)
	}
}

// TestSetEnvAndMeta verify setting environment variables and meta data from capabilities and previously set environment.
func TestSetEnvAndMeta(t *testing.T) {
	tests := map[string]struct {
		layout                  ServiceSpec
		labels                  map[string]string
		envs                    []apiv1.EnvVar
		labels_should_not_exist []string
		envs_should_not_exist   []string
	}{
		"screenResolution from caps": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					ScreenResolution: "800x600",
				},
			},
			labels: map[string]string{
				defaultsAnnotations.screenResolution: "800x600",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.screenResolution, Value: "800x600"},
			},
		},
		"screenResolution from caps and env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					ScreenResolution: "800x600",
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.screenResolution, Value: "1024x768"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.screenResolution: "800x600",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.screenResolution, Value: "800x600"},
			},
		},
		"screenResolution from env": {
			layout: ServiceSpec{
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.screenResolution, Value: "1024x768"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.screenResolution: "1024x768",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.screenResolution, Value: "1024x768"},
			},
		},
		"enableVNC from caps": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVNC: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVNC, Value: "true"},
			},
		},
		"enableVNC from caps and env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					VNC: true,
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.enableVNC, Value: "false"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVNC: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVNC, Value: "true"},
			},
		},
		"enableVNC from env": {
			layout: ServiceSpec{
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.enableVNC, Value: "true"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVNC: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVNC, Value: "true"},
			},
		},
		"timeZone from caps": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					TimeZone: "Europe/Moscow",
				},
			},
			labels: map[string]string{
				defaultsAnnotations.timeZone: "Europe/Moscow",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.timeZone, Value: "Europe/Moscow"},
			},
		},
		"timeZone from caps and env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					TimeZone: "Europe/Moscow",
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.timeZone, Value: "Europe/Amsterdam"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.timeZone: "Europe/Moscow",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.timeZone, Value: "Europe/Moscow"},
			},
		},
		"timeZone from env": {
			layout: ServiceSpec{
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.timeZone, Value: "Europe/Moscow"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.timeZone: "Europe/Moscow",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.timeZone, Value: "Europe/Moscow"},
			},
		},
		"only videoEnable from caps": {
			layout: ServiceSpec{
				SessionID: "test",
				RequestedCapabilities: selenium.Capabilities{
					Video: true,
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo: "true",
				defaultsAnnotations.videoName:   "test.mp4",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoName, Value: "test.mp4"},
			},
		},
		"only videoEnable from caps and env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video: true,
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.enableVideo, Value: "false"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
			},
		},
		"only videoEnable from env": {
			layout: ServiceSpec{
				SessionID: "chrome-with-video",
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.enableVideo, Value: "true"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
			},
		},
		"videoName from caps": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video:     true,
					VideoName: "abc.mp4",
				},
			},
			labels: map[string]string{
				defaultsAnnotations.videoName:   "abc.mp4",
				defaultsAnnotations.enableVideo: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.videoName, Value: "abc.mp4"},
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
			},
		},
		"videoName from caps and env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video:     true,
					VideoName: "abc.mp4",
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.videoName, Value: "efg.mp4"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.videoName:   "abc.mp4",
				defaultsAnnotations.enableVideo: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.videoName, Value: "abc.mp4"},
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
			},
		},
		"videoName from env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video: true,
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.videoName, Value: "abc.mp4"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.videoName:   "abc.mp4",
				defaultsAnnotations.enableVideo: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.videoName, Value: "abc.mp4"},
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
			},
		},
		"videoScreenSize from caps": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video:           true,
					VideoScreenSize: "800x600",
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo:     "true",
				defaultsAnnotations.videoScreenSize: "800x600",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoScreenSize, Value: "800x600"},
			},
		},
		"videoScreenSize from caps and env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video:           true,
					VideoScreenSize: "800x600",
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.videoScreenSize, Value: "300x200"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo:     "true",
				defaultsAnnotations.videoScreenSize: "800x600",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoScreenSize, Value: "800x600"},
			},
		},
		"videoScreenSize from env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video: true,
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.videoScreenSize, Value: "800x600"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoScreenSize, Value: "800x600"},
			},
		},
		"videFrameRate from caps": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video:          true,
					VideoFrameRate: 24,
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo:    "true",
				defaultsAnnotations.videoFrameRate: "24",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoFrameRate, Value: "24"},
			},
		},
		"videoFrameRate from caps and env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video:          true,
					VideoFrameRate: 24,
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.videoFrameRate, Value: "60"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo:    "true",
				defaultsAnnotations.videoFrameRate: "24",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoFrameRate, Value: "24"},
			},
		},
		"videoFrameRate from caps": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video: true,
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.videoFrameRate, Value: "24"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoFrameRate, Value: "24"},
			},
		},
		"videoCodec from caps": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video:      true,
					VideoCodec: "codec-name",
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo: "true",
				defaultsAnnotations.videoCodec:  "codec-name",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoCodec, Value: "codec-name"},
			},
		},
		"videoCodec from caps and env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video:      true,
					VideoCodec: "codec-name",
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.videoCodec, Value: "wrong-codec-name"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo: "true",
				defaultsAnnotations.videoCodec:  "codec-name",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoCodec, Value: "codec-name"},
			},
		},
		"videoCodec from env": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video: true,
				},
				Template: BrowserSpec{
					Spec: Spec{
						EnvVars: []apiv1.EnvVar{
							{Name: defaultsAnnotations.videoCodec, Value: "codec-name"},
						},
					},
				},
			},
			labels: map[string]string{
				defaultsAnnotations.enableVideo: "true",
			},
			envs: []apiv1.EnvVar{
				{Name: defaultsAnnotations.enableVideo, Value: "true"},
				{Name: defaultsAnnotations.videoCodec, Value: "codec-name"},
			},
		},
		"no video caps if videoEnable=false": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{
					Video: false,
				},
			},
			labels_should_not_exist: []string{
				defaultsAnnotations.videoCodec,
				defaultsAnnotations.videoScreenSize,
				defaultsAnnotations.videoFrameRate,
				defaultsAnnotations.videoName,
			},
			envs_should_not_exist: []string{
				defaultsAnnotations.enableVideo,
				defaultsAnnotations.videoCodec,
				defaultsAnnotations.videoScreenSize,
				defaultsAnnotations.videoFrameRate,
				defaultsAnnotations.videoName,
			},
		},
		"only default envs and labels exist": {
			layout: ServiceSpec{
				RequestedCapabilities: selenium.Capabilities{},
			},
			labels_should_not_exist: []string{
				defaultsAnnotations.screenResolution,
				defaultsAnnotations.enableVNC,
				defaultsAnnotations.timeZone,
				defaultsAnnotations.enableVideo,
				defaultsAnnotations.videoName,
				defaultsAnnotations.videoCodec,
				defaultsAnnotations.videoScreenSize,
				defaultsAnnotations.videoFrameRate,
			},
			envs_should_not_exist: []string{
				defaultsAnnotations.screenResolution,
				defaultsAnnotations.enableVNC,
				defaultsAnnotations.timeZone,
				defaultsAnnotations.enableVideo,
				defaultsAnnotations.videoCodec,
				defaultsAnnotations.videoScreenSize,
				defaultsAnnotations.videoFrameRate,
				defaultsAnnotations.videoName,
			},
		},
	}

	for name, test := range tests {
		t.Logf("TC: Vrify setting caps and lbls for %v", name)
		layout := setEnvAndMeta(test.layout)

		for _, v := range test.envs {
			assert.Contains(t, layout.Template.Spec.EnvVars, v)
		}

		for _, l := range test.labels {
			assert.Contains(t, layout.Template.Meta.Annotations["capabilities"], l)
		}

		for _, l := range test.labels_should_not_exist {
			assert.NotContains(t, layout.Template.Meta.Annotations["capabilities"], l)
		}

		var currentEnvVars []string
		for _, v := range layout.Template.Spec.EnvVars {
			currentEnvVars = append(currentEnvVars, v.Name)
		}
		for _, v := range test.envs_should_not_exist {
			assert.NotContains(t, currentEnvVars, v)
		}
	}
}

// TestPodBuild verify pod building func
func TestPodBuild(t *testing.T) {
	privileged := true
	tests := map[string]struct {
		svc           string
		layout        ServiceSpec
		numContainers int
		proxyImage    string
		videoImage    string
	}{
		"build simple pod": {
			proxyImage:    "seleniferous",
			svc:           "selenosis",
			numContainers: 2,
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
					Privileged:     &privileged,
					Spec: Spec{
						NodeSelector: map[string]string{
							"disktype": "ssd",
						},
						HostAliases: []apiv1.HostAlias{
							{Hostnames: []string{"localhost"}},
						},
						DNSConfig: apiv1.PodDNSConfig{
							Nameservers: []string{
								"8.8.8.8",
							},
						},
						Tolerations: []apiv1.Toleration{
							{
								Key:      "example-key",
								Operator: "Exist",
								Effect:   "NoSchedule",
							},
						},
						ServiceAccountName: "selenosis",
						PriorityClassName:  "selenosis",
						EnvVars: []apiv1.EnvVar{
							{Name: "FILE_NAME", Value: "TEST"},
						},
						Resources: apiv1.ResourceRequirements{
							Requests: apiv1.ResourceList{
								apiv1.ResourceCPU:    resource.MustParse("250m"),
								apiv1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		},
		"build pod with video": {
			svc:           "selenosis",
			numContainers: 3,
			videoImage:    "selenoid-video-recorder",
			layout: ServiceSpec{
				SessionID: "chrome-85-0-de44c3c4-1a35-412b-b526-f5da802144911",
				RequestedCapabilities: selenium.Capabilities{
					Video: true,
				},
				Template: BrowserSpec{
					BrowserName:    "chrome",
					BrowserVersion: "85.0",
					Image:          "selenoid/vnc:chrome_85.0",
					Path:           "/",
				},
			},
		},
	}

	for name, c := range tests {
		t.Logf("TC: Verify %s", name)

		mock := fake.NewSimpleClientset()
		service := &service{
			clientset:  mock,
			svc:        c.svc,
			videoImage: c.videoImage,
			proxyImage: c.proxyImage,
		}
		pod := service.buildPod(c.layout)

		assert.Equal(t, c.layout.SessionID, pod.ObjectMeta.Name)
		assert.Equal(t, c.layout.SessionID, pod.Spec.Hostname)
		assert.Equal(t, c.svc, pod.Spec.Subdomain)
		assert.Equal(t, c.layout.Template.Spec.NodeSelector, pod.Spec.NodeSelector)
		assert.Equal(t, c.layout.Template.Spec.HostAliases, pod.Spec.HostAliases)
		assert.Equal(t, &c.layout.Template.Spec.DNSConfig, pod.Spec.DNSConfig)
		assert.Equal(t, c.layout.Template.Spec.Tolerations, pod.Spec.Tolerations)
		assert.Equal(t, c.layout.Template.Spec.ServiceAccountName, pod.Spec.ServiceAccountName)
		assert.Equal(t, c.layout.Template.Spec.PriorityClassName, pod.Spec.PriorityClassName)

		assert.Equal(t, c.numContainers, len(pod.Spec.Containers))

		browser := pod.Spec.Containers[0]
		assert.Equal(t, "browser", browser.Name)
		assert.Equal(t, c.layout.Template.Image, browser.Image)
		assert.Equal(t, c.layout.Template.Privileged, browser.SecurityContext.Privileged)
		assert.Equal(t, c.layout.Template.Spec.EnvVars, browser.Env)
		assert.Equal(t, c.layout.Template.Spec.Resources, browser.Resources)

		proxy := pod.Spec.Containers[1]
		assert.Equal(t, "seleniferous", proxy.Name)
		assert.Equal(t, c.proxyImage, proxy.Image)

		var lifecycle *apiv1.Lifecycle = nil
		if c.layout.RequestedCapabilities.Video {
			lifecycle = &apiv1.Lifecycle{
				PreStop: &apiv1.Handler{
					Exec: &apiv1.ExecAction{
						Command: []string{"sh", "-c", "sleep 5"},
					},
				},
			}

		}
		assert.Equal(t, lifecycle, pod.Spec.Containers[0].Lifecycle)

		if !c.layout.RequestedCapabilities.Video {
			return
		}
		videoRecorder := pod.Spec.Containers[2]
		assert.Equal(t, "video-recorder", videoRecorder.Name)
		assert.Equal(t, c.videoImage, videoRecorder.Image)
		assert.Equal(t, c.layout.Template.Spec.EnvVars, videoRecorder.Env)
	}
}

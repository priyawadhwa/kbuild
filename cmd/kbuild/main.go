package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/priyawadhwa/kbuild/pkg/storage"
	"github.com/priyawadhwa/kbuild/pkg/util"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	batch "k8s.io/api/batch/v1"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig *string
	logrus.SetLevel(logrus.DebugLevel)
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	var dockerfilePath = flag.String("dockerfile", "/dockerfile/Dockerfile", "path to dockerfile")
	var context = flag.String("context", "", "source context")
	var name = flag.String("name", "gcr.io/priya-wadhwa/kbuild:finalimage", "name of image destination")
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// create source context
	bucket, _, err := storage.CreateStorageBucket()
	if err != nil {
		panic(err)
	}
	if err := storage.UploadContextToBucket(*context, bucket); err != nil {
		panic(err)
	}
	return

	env := v1.EnvVar{
		Name:  "KBUILD_DEST_IMAGE",
		Value: *name,
	}

	b, err := ioutil.ReadFile(*dockerfilePath)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "build-dockerfile",
		},
		Data: map[string]string{
			"Dockerfile": string(b),
		},
	}

	cfgmap, err := clientset.CoreV1().ConfigMaps("default").Create(cfg)
	if err != nil {
		log.Fatal(err)
	}

	j := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "build-job-",
		},
		Spec: batch.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "init-static",
							Image: "gcr.io/priya-wadhwa/kbuilder:latest",
							Command: []string{
								"/work-dir/main",
							},
							Args:         []string{},
							VolumeMounts: []v1.VolumeMount{v1.VolumeMount{Name: "dockerfile", MountPath: "/dockerfile"}},
							Env:          []v1.EnvVar{env},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Volumes: []v1.Volume{
						{
							Name: "dockerfile",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: cfgmap.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	job, err := clientset.BatchV1().Jobs("default").Create(j)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created build job: ", job.Name)

	stopCh := make(chan bool)

	for {
		j, err = clientset.BatchV1().Jobs("default").Get(job.Name, metav1.GetOptions{})
		if err != nil {
			// wait until the job exists
			continue
		}
		break
	}

	// Stream logs
	fmt.Println("Stream logs")
	for {
		opts := metav1.ListOptions{LabelSelector: labels.Set(j.Spec.Selector.MatchLabels).AsSelector().String()}
		jobPods, err := clientset.CoreV1().Pods("default").List(opts)
		if err != nil {
			continue
		}
		// Stream logs
		for _, p := range jobPods.Items {
			f := func() {
				streamLogs(clientset, "do-build", p.Name, "default")
			}
			go util.Until(f, stopCh)
		}
		break
	}
	fmt.Println("Something omsething")
	for {
		j, err := clientset.BatchV1().Jobs("default").Get(job.Name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		if j.Status.CompletionTime == nil {
			time.Sleep(2 * time.Second)
		} else {
			fmt.Println("Job finished.")
			stopCh <- true
			break
		}
	}
}

func streamLogs(clientset *kubernetes.Clientset, container, pod, namespace string) error {
	r, err := clientset.CoreV1().Pods(namespace).GetLogs(pod, &v1.PodLogOptions{Container: container, Follow: true}).Stream()
	if err != nil {
		return err
	}
	defer r.Close()
	if _, err := io.Copy(os.Stderr, r); err != nil {
		return err
	}
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

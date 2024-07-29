package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type containerRes struct {
	identifier string
	cpuReq     resource.Quantity
	cpuLim     resource.Quantity
	memReq     resource.Quantity
	memLim     resource.Quantity
}

func main() {

	var ns = flag.String("namespace", "", "namespace to scan")

	clientset := BuildClient()

	res := GetAllContainerResources(clientset, ns)

	totalCpuReq := resource.Quantity{}
	totalCpuLim := resource.Quantity{}
	totalMemReq := resource.Quantity{}
	totalMemLim := resource.Quantity{}

	fmt.Printf("container | CPU req | CPU lim | Memory req | Memory lim \n")

	for _, line := range res {

		fmt.Printf("%s | %s | %s | %s | %s\n",
			line.identifier, &line.cpuReq, &line.cpuLim, &line.memReq, &line.memLim)

		totalMemReq.Add(line.memReq)
		totalMemLim.Add(line.memLim)
		totalCpuReq.Add(line.cpuReq)
		totalCpuLim.Add(line.cpuLim)
	}

	fmt.Printf("Total | %s | %s | %s | %s\n", &totalCpuReq, &totalCpuLim, &totalMemReq, &totalMemLim)

}

func GetAllContainerResources(clientset *kubernetes.Clientset, ns *string) []containerRes {

	pods := GetNamespacePods(clientset, ns)

	res := make([]containerRes, 0)

	for _, pod := range pods.Items {

		for _, cont := range pod.Spec.Containers {

			res = append(res, containerRes{
				identifier: fmt.Sprintf("%s:%s", pod.Name, cont.Name),
				cpuReq:     *cont.Resources.Requests.Cpu(),
				cpuLim:     *cont.Resources.Limits.Cpu(),
				memReq:     *cont.Resources.Requests.Memory(),
				memLim:     *cont.Resources.Limits.Memory(),
			})
		}
	}

	return res
}

func GetNamespacePods(clientset *kubernetes.Clientset, ns *string) *v1.PodList {

	pods, err := clientset.CoreV1().Pods(*ns).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	return pods
}

func BuildClient() *kubernetes.Clientset {

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
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

	return clientset
}

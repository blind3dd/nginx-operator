package assets

import (
	"embed"
	appsv1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/pointer"
)

var (
	//go:embed manifests/*
	deployment embed.FS
	appsScheme = runtime.NewScheme()
	appsCodecs = serializer.NewCodecFactory(appsScheme)
)

func init() {
	if err := appsv1.AddToScheme(appsScheme); err != nil {
		panic(err)
	}
}

func GetDeploymentFromFile(manifestName string) *appsv1.Deployment {
	manifestName = "assets/manifests/nginx_deployment.yaml"
	deployObjectBytes, err := deployment.ReadFile(manifestName)
	if err != nil {
		panic(err)
	}
	deployObject, err := runtime.Decode(appsCodecs.UniversalDecoder(appsv1.SchemeGroupVersion),
		deployObjectBytes)
	if err != nil {
		panic(err)
	}

	deploy := deployObject.(*appsv1.Deployment)

	return deploy
}

func GetDeployment() *appsv1.Deployment {
	deploymentBytes, err := nginxDeployment().Marshal()
	if err != nil {
		panic(err)
	}
	deploymentObject, err := runtime.Decode(appsCodecs.UniversalDecoder(appsv1.SchemeGroupVersion),
		deploymentBytes)
	if err != nil {
		panic(err)
	}
	dep := deploymentObject.(*appsv1.Deployment)

	return dep
}

func nginxDeployment() *appsv1.Deployment {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: "nginx-operator-system",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"k8s-app": "nginx"},
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"k8s-app": "nginx"},
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{{
						Image:   "nginx:latest",
						Name:    "nginx",
						Command: []string{"nginx"},
						Ports: []v12.ContainerPort{{
							ContainerPort: 8080,
							Name:          "nginx",
						}},
					},
					},
				},
			},
		},
	}

	return dep
}

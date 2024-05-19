package secret

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type kubernetesSecret struct {
	secret    string
	key       string
	clientset *kubernetes.Clientset
	value     []byte
}

type KubernetesSecretOptions struct {
	Secret string
	Key    string
}

func NewKubernetesSecretValue(options KubernetesSecretOptions) (Secret, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	secret := kubernetesSecret{}
	if err := secret.Reload(); err != nil {
		return nil, err
	}

	return &kubernetesSecret{
		secret:    options.Secret,
		key:       options.Key,
		clientset: clientset,
	}, nil
}

func (s *kubernetesSecret) Get() ([]byte, error) {
	return s.value, nil
}

func (s *kubernetesSecret) Reload() error {
	secret, err := s.clientset.CoreV1().Secrets("").Get(context.TODO(), s.secret, v1.GetOptions{})
	if err != nil {
		return err
	}

	data := secret.Data[s.key]
	s.value = data
	return nil
}

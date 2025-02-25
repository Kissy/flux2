/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sync

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"

	"github.com/fluxcd/flux2/pkg/manifestgen"
)

func Generate(options Options) (*manifestgen.Manifest, error) {
	gvk := sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind)
	gitRepository := sourcev1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
		Spec: sourcev1.GitRepositorySpec{
			URL: options.URL,
			Interval: metav1.Duration{
				Duration: options.Interval,
			},
			Reference: &sourcev1.GitRepositoryRef{
				Branch: options.Branch,
			},
			SecretRef: &corev1.LocalObjectReference{
				Name: options.Name,
			},
		},
	}

	gitData, err := yaml.Marshal(gitRepository)
	if err != nil {
		return nil, err
	}

	gvk = kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind)
	kustomization := kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{
				Duration: 10 * time.Minute,
			},
			Path:  fmt.Sprintf("./%s", strings.TrimPrefix(options.TargetPath, "./")),
			Prune: true,
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: sourcev1.GitRepositoryKind,
				Name: options.Name,
			},
			Validation: "client",
		},
	}

	ksData, err := yaml.Marshal(kustomization)
	if err != nil {
		return nil, err
	}

	return &manifestgen.Manifest{
		Path:    path.Join(options.TargetPath, options.Namespace, options.ManifestFile),
		Content: fmt.Sprintf("---\n%s---\n%s", resourceToString(gitData), resourceToString(ksData)),
	}, nil
}

func resourceToString(data []byte) string {
	data = bytes.Replace(data, []byte("  creationTimestamp: null\n"), []byte(""), 1)
	data = bytes.Replace(data, []byte("status: {}\n"), []byte(""), 1)
	return string(data)
}

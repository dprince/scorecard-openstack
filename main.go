// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright 2023 Red Hat Inc.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	scapiv1alpha3 "github.com/operator-framework/api/pkg/apis/scorecard/v1alpha3"
	apimanifests "github.com/operator-framework/api/pkg/manifests"
	csvv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

const (
	PodBundleRoot            = "/bundle"
	CustomRelatedImagesCheck = "related-images-check"
	CustomAnnotationsCheck   = "annotations-check"
	CustomInstallModesCheck  = "install-modes-check"
)

func main() {
	entrypoint := os.Args[1:]
	if len(entrypoint) == 0 {
		log.Fatal("Test name argument is required")
	}

	// Read the pod's untar'd bundle from a well-known path.
	cfg, err := apimanifests.GetBundleFromDir(PodBundleRoot)
	if err != nil {
		log.Fatal(err.Error())
	}

	var result scapiv1alpha3.TestStatus

	switch entrypoint[0] {
	case CustomRelatedImagesCheck:
		result = RelatedImagesCheck(cfg)
	case CustomAnnotationsCheck:
		result = AnnotationsCheck(cfg)
	case CustomInstallModesCheck:
		result = InstallModesCheck(cfg)
	default:
		result = printValidTests()
	}

	// Convert scapiv1alpha3.TestResult to json.
	prettyJSON, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		log.Fatal("Failed to generate json", err)
	}
	fmt.Printf("%s\n", string(prettyJSON))

}

// printValidTests will print out full list of test names to give a hint to the end user on what the valid tests are.
func printValidTests() scapiv1alpha3.TestStatus {
	result := scapiv1alpha3.TestResult{}
	result.State = scapiv1alpha3.FailState
	result.Errors = make([]string, 0)
	result.Suggestions = make([]string, 0)

	str := fmt.Sprintf("Valid tests for this image include: %s %s %s",
		CustomRelatedImagesCheck,
		CustomAnnotationsCheck,
		CustomInstallModesCheck)
	result.Errors = append(result.Errors, str)
	return scapiv1alpha3.TestStatus{
		Results: []scapiv1alpha3.TestResult{result},
	}
}

func RelatedImagesCheck(bundle *apimanifests.Bundle) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = CustomRelatedImagesCheck
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)
	relatedImages := bundle.CSV.Spec.RelatedImages
	for _, image := range relatedImages {
		// verify that kube-rbac-proxy is a SHA256 form of gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1
		// we want these to stay in sync
		if image.Name == "kube-rbac-proxy" {
			if image.Image != "gcr.io/kubebuilder/kube-rbac-proxy@sha256:d4883d7c622683b3319b5e6b3a7edfbf2594c18060131a8bf64504805f875522" {
				r.State = scapiv1alpha3.FailState
				r.Errors = append(r.Errors, "kube-rbac-proxy does not match a SHA256 form of gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1")
			}
		}
	}
	return wrapResult(r)
}

func AnnotationsCheck(bundle *apimanifests.Bundle) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = CustomAnnotationsCheck
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)
	annotations := bundle.CSV.GetAnnotations()
	if annotations["operators.openshift.io/infrastructure-features"] != "[\"disconnected\"]" {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, "Missing annotation for disconnected/offline operator installation support")
	}
	if annotations["operators.operatorframework.io/operator-type"] != "non-standalone" {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, "Missing annotation for operator type: non-standalone")
	}
	// only the openstack-operator should have the suggested-namespace set
	if !strings.HasPrefix(bundle.CSV.Name, "openstack-operator") && annotations["operatorframework.io/suggested-namespace"] == "" {
		r.State = scapiv1alpha3.FailState
		r.Errors = append(r.Errors, "Missing annotation for operator type: non-standalone")
	}
	return wrapResult(r)
}

func InstallModesCheck(bundle *apimanifests.Bundle) scapiv1alpha3.TestStatus {
	r := scapiv1alpha3.TestResult{}
	r.Name = CustomInstallModesCheck
	r.State = scapiv1alpha3.PassState
	r.Errors = make([]string, 0)
	r.Suggestions = make([]string, 0)
	for _, mode := range bundle.CSV.Spec.InstallModes {
		if mode.Type == csvv1alpha1.InstallModeTypeOwnNamespace {
			if !mode.Supported {
				r.State = scapiv1alpha3.FailState
				r.Errors = append(r.Errors, "installMode type OwnNamespace should be true")
			}
		}
		if mode.Type == csvv1alpha1.InstallModeTypeSingleNamespace {
			if !mode.Supported {
				r.State = scapiv1alpha3.FailState
				r.Errors = append(r.Errors, "installMode type SingleNamespace should be true")
			}
		}
		if mode.Type == csvv1alpha1.InstallModeTypeMultiNamespace {
			if mode.Supported {
				r.State = scapiv1alpha3.FailState
				r.Errors = append(r.Errors, "installMode type MultiNamespace should be false")
			}
		}
		if mode.Type == csvv1alpha1.InstallModeTypeAllNamespaces {
			if !mode.Supported {
				r.State = scapiv1alpha3.FailState
				r.Errors = append(r.Errors, "installMode type AllNamespaces should be true")
			}
		}
	}
	return wrapResult(r)
}

func wrapResult(r scapiv1alpha3.TestResult) scapiv1alpha3.TestStatus {
	return scapiv1alpha3.TestStatus{
		Results: []scapiv1alpha3.TestResult{r},
	}
}

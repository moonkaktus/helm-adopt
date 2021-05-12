package discovery

import (
	"context"
	"fmt"
	"github.com/HamzaZo/helm-adopt/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MatchedResources struct {
	Gvr map[bool][]schema.GroupVersionResource
	WantRes map[string][]string
}

var (
	selectedGVR []schema.GroupVersionResource
	objects []string
	f *unstructured.Unstructured
)


//TODO Next task - 0
// TODO add helm plugin to fetch namespace auto : Done
// TODO create release and set helm as ownership of resources : Done

// TODO use releaseName validate release Name + use auto generate release
// TODO Later Task - 1
// TODO test with other resources cm,secrets,pv,pvc etc..
// TODO clear clusterIps []string
//TODO refactor this func it's too complicated :: using 4 for loop not a good idea

// TODO Later Task - 2
// TODO add templating
// TODO create a fresh repo
// TODO add Unit tests

func (m MatchedResources) Query(client *ApiClient, namespace string) (map[string][]byte, error) {
	var err error
	result := make(map[string][]byte)

	for res, object := range m.WantRes{
		for namespaced, g := range m.Gvr {
			for _, gvr := range g {
				if res != gvr.Resource {
					continue
				}
				j := 0
				for _, k := range object{
					if namespaced {
						f, err = client.DynClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), k, metav1.GetOptions{})
						if err != nil {
							return nil, err
						}
					} else {
						f, err = client.DynClient.Resource(gvr).Get(context.TODO(), k, metav1.GetOptions{})
						if err != nil {
							return nil, err
						}
					}
					err = deepCleaning(f)
					if err != nil {
						return nil, err
					}
					output, err := utils.GetPrettyYaml(f)
					if err != nil {
						return nil, err
					}
					result[fmt.Sprintf("%s-%v", res, j)] = output

					j++
				}
			}

		}
	}
	return result, nil

}

func deepCleaning(obj *unstructured.Unstructured) error {
	return func(f *unstructured.Unstructured) error {
		utils.CommonCleaning(f)
		status, found, err := unstructured.NestedFieldNoCopy(f.Object, "status")
		if err != nil {
			return err
		}
		if found {
			utils.CleanStatus(status.(map[string]interface{}))
		}
		return nil
	}(obj)

}

func FetchedFilteredResources(client *ApiClient, wantResources map[string][]string) (namespaceResource, clusterResource *MatchedResources, err error) {
	groupResources, err := getResources(client)
	if err != nil {
		return nil, nil, err
	}
	return filteringProcess(groupResources, true, wantResources), filteringProcess(groupResources, false, wantResources), nil
}

func filteringProcess(gvrs map[schema.GroupVersion][]metav1.APIResource, namespaced bool, wantResources map[string][]string) *MatchedResources {
	wR := make(map[string][]string)
	result := make(map[bool][]schema.GroupVersionResource)

	for gv, resources := range gvrs {
		for _, res := range resources {
			// Get either namespaced or non-namespaced resources.
			if namespaced != res.Namespaced {
				continue
			}

			if wantResources != nil {
				ok := false
				if objects, ok = utils.Contains(wantResources, res.Name); !ok {
					continue
				}
			}
			selectedGVR = append(selectedGVR, gv.WithResource(res.Name))

			result[namespaced] = selectedGVR
			wR[res.Name] = objects
		}
	}
	selectedGVR = nil
	return &MatchedResources{
		Gvr: result,
		WantRes: wR,
	}
}

func getResources(client *ApiClient) (map[schema.GroupVersion][]metav1.APIResource, error) {

	resourceLists, err := client.ClientSet.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	versionResource := map[schema.GroupVersion][]metav1.APIResource{}

	for _, apiResourceList := range resourceLists {
		version, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("unable to parse GroupVersion %v",err)
		}

		versionResource[version] = uniqResources(apiResourceList.APIResources)
	}

	return versionResource, nil
}

func uniqResources(resources []metav1.APIResource) []metav1.APIResource {
	seen := make(map[string]struct{}, len(resources))
	i := 0
	for _, k := range resources {
		if _, ok := seen[k.Name]; ok {
			continue
		}
		seen[k.Name] = struct{}{}
		resources[i] = k

		i++
	}
	return resources[:i]
}
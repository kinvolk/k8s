package k8s

import (
	"context"
	"errors"
	"fmt"

	"github.com/ericchiang/k8s/api/v1"
	"github.com/ericchiang/k8s/watch/versioned"
)

// ThirdPartyResources is a client used for interacting with user
// defined API groups. It uses JSON encoding instead of protobufs
// which are unsupported for these APIs.
//
// Users are expected to define their own third party resources.
//
//		const metricsResource = "metrics"
//
//		// First, define a third party resources with TypeMeta
//		// and ObjectMeta fields.
//		type Metric struct {
//			*unversioned.TypeMeta   `json:",inline"`
//			*v1.ObjectMeta          `json:"metadata,omitempty"`
//
//			Timestamp time.Time `json:"timestamp"`
//			Value     []byte    `json:"value"`
//		}
//
//		// Define a list wrapper.
//		type MetricsList struct {
//			*unversioned.TypeMeta `json:",inline"`
//			*unversioned.ListMeta `json:"metadata,omitempty"`
//
//			Items []Metric `json:"items"`
//		}
//
// Register the new resource by creating a ThirdPartyResource type.
//
//		// Create a ThirdPartyResource
//		tpr := &v1beta1.ThirdPartyResource{
//			Metadata: &v1.ObjectMeta{
//				Name: k8s.String("metric.metrics.example.com"),
//			},
//			Description: k8s.String("A custom third party resource"),
//			Versions:    []*v1beta1.APIVersions{
//				{Name: k8s.String("v1")},
//			},
//		}
//		_, err := client.ExtensionsV1Beta1().CreateThirdPartyResource(ctx, trp)
//		if err != nil {
//			// handle error
//		}
//
// After creating the resource type, create a ThirdPartyResources client then
// use interact with it like any other API group. For example to create a third
// party resource:
//
//		metricsClient := client.ThirdPartyResources("metrics.example.com", "v1")
//
//		metric := &Metric{
//			ObjectMeta: &v1.ObjectMeta{
//				Name: k8s.String("foo"),
//			},
//			Timestamp: time.Now(),
//			Value:     42,
//		}
//
//		err = metricsClient.Create(ctx, metricsResource, client.Namespace, metric, metric)
//		if err != nil {
//			// handle error
//		}
//
// List a set of third party resources:
//
//		var metrics MetricsList
//		metricsClient.List(ctx, metricsResource, &metrics)
//
// Or delete:
//
//		tprClient.Delete(ctx, metricsResource, client.Namespace, *metric.Name)
//
type ThirdPartyResources struct {
	c *Client

	apiGroup   string
	apiVersion string
}

// ThirdPartyResources returns a client for interacting with a ThirdPartyResource
// API group.
func (c *Client) ThirdPartyResources(apiGroup, apiVersion string) *ThirdPartyResources {
	return &ThirdPartyResources{c, apiGroup, apiVersion}
}

func checkResource(apiGroup, apiVersion, resource, namespace, name string) error {
	if apiGroup == "" {
		return errors.New("no api group provided")
	}
	if apiVersion == "" {
		return errors.New("no api version provided")
	}
	if resource == "" {
		return errors.New("no resource version provided")
	}
	if namespace == "" {
		return errors.New("no namespace provided")
	}
	if name == "" {
		return errors.New("no resource name provided")
	}
	return nil
}

type object interface {
	GetMetadata() *v1.ObjectMeta
}

func (t *ThirdPartyResources) Create(ctx context.Context, resource, namespace string, req, resp interface{}) error {
	if err := checkResource(t.apiGroup, t.apiVersion, resource, namespace, "not required"); err != nil {
		return err
	}
	url := t.c.urlFor(t.apiGroup, t.apiVersion, namespace, resource, "")
	return t.c.create(ctx, jsonCodec, "POST", url, req, resp)
}

func (t *ThirdPartyResources) Update(ctx context.Context, resource, namespace, name string, req, resp interface{}) error {
	if err := checkResource(t.apiGroup, t.apiVersion, resource, namespace, "not required"); err != nil {
		return err
	}
	url := t.c.urlFor(t.apiGroup, t.apiVersion, namespace, resource, name)
	return t.c.create(ctx, jsonCodec, "PUT", url, req, resp)
}

func (t *ThirdPartyResources) Get(ctx context.Context, resource, namespace, name string, resp interface{}) error {
	if err := checkResource(t.apiGroup, t.apiVersion, resource, namespace, name); err != nil {
		return err
	}
	url := t.c.urlFor(t.apiGroup, t.apiVersion, namespace, resource, name)
	return t.c.get(ctx, jsonCodec, url, resp)
}

func (t *ThirdPartyResources) Delete(ctx context.Context, resource, namespace, name string) error {
	if err := checkResource(t.apiGroup, t.apiVersion, resource, namespace, name); err != nil {
		return err
	}
	url := t.c.urlFor(t.apiGroup, t.apiVersion, namespace, resource, name)
	return t.c.delete(ctx, jsonCodec, url)
}

func (t *ThirdPartyResources) List(ctx context.Context, resource, namespace string, resp interface{}) error {
	if err := checkResource(t.apiGroup, t.apiVersion, resource, namespace, "name not required"); err != nil {
		return err
	}
	url := t.c.urlFor(t.apiGroup, t.apiVersion, namespace, resource, "")
	return t.c.get(ctx, jsonCodec, url, resp)
}

type ThirdPartyResourceWatcher struct {
	watcher *watcher
}

func (w *ThirdPartyResourceWatcher) Next() (*versioned.Event, interface{}, error) {
	event, unknown, err := w.watcher.next()
	if err != nil {
		return nil, nil, err
	}
	fmt.Println(unknown.Raw)
	var resp interface{}
	/* if err := proto.Unmarshal(unknown.Raw, resp); err != nil {
		return nil, nil, err
	} */
	return event, resp, nil
}

func (w *ThirdPartyResourceWatcher) Close() error {
	return w.watcher.Close()
}

func (t *ThirdPartyResources) Watch(ctx context.Context, resource, namespace string) (*ThirdPartyResourceWatcher, error) {
	if err := checkResource(t.apiGroup, t.apiVersion, resource, namespace, "name not required"); err != nil {
		return nil, err
	}
	url := t.c.urlFor(t.apiGroup, t.apiVersion, namespace, resource, "")
	watcher, err := t.c.watch(ctx, url)
	if err != nil {
		return nil, err
	}
	return &ThirdPartyResourceWatcher{watcher}, nil
}

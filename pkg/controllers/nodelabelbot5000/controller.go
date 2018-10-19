package nodelabelbot5000

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"os"

	"github.com/samsung-cnct/nodelabelbot5000/pkg/util/k8sutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const (
	KubernetesWorkerLabelName  = "node-role.kubernetes.io/worker"
	KubernetesWorkerLabelValue = "true"
	SDSWorkerLabelName         = "kubernetes.io/nodetype"
	SDSWorkerLabelValue        = "app"
	KubernetesCPLabelName      = "node-role.kubernetes.io/master"
)

type NodeLabelBot5000Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller

	client *kubernetes.Clientset
}

func NewNodeLabelBot5000Controller(config *rest.Config) (output *NodeLabelBot5000Controller, err error) {
	if config == nil {
		config = k8sutil.DefaultConfig
	}
	if config == nil {
		logger.Criticalf("Oh Nodes! I can't get a valid kubeconfig!  Yes that was an intentional mispell! Fix me!")
		os.Exit(1)
	}
	client := kubernetes.NewForConfigOrDie(config)

	// create sdsapplication list/watcher
	nodeListWatcher := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"nodes",
		"",
		fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	sharedInformer := cache.NewSharedIndexInformer(
		nodeListWatcher,
		&v1.Node{},
		60*time.Minute,
		cache.Indexers{},
	)

	sharedInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, 60*time.Minute)

	output = &NodeLabelBot5000Controller{
		informer: sharedInformer,
		indexer:  sharedInformer.GetIndexer(),
		queue:    queue,
		client:   client,
	}
	output.SetLogger()
	return
}

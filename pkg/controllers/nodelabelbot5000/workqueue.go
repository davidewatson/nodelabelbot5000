package nodelabelbot5000

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"time"
)

func (c *NodeLabelBot5000Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two SDSClusters with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.processItem(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// processItem is the business logic of the controller.
func (c *NodeLabelBot5000Controller) processItem(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)

	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a SDSPackageManager, so that we will see a delete for one SDSPackageManager
		fmt.Printf("Node -->%s<-- does not exist anymore\n", key)
	} else {
		node, ok := obj.(*v1.Node)
		if !ok {
			logger.Errorf("Could not cast item contents from key -->%s<-- into a Node! This is bad!", key)
			return fmt.Errorf("could not cast item contents from key -->%s<-- into a node", key)
		}

		_, cpNodeLabelPresent := node.Labels[KubernetesCPLabelName]
		workerNodeLabelValue, workerNodeLabelPresent := node.Labels[KubernetesWorkerLabelName]
		sdsWorkerNodeLabelValue, sdsWorkerNodeLabelPresent := node.Labels[SDSWorkerLabelName]
		if !cpNodeLabelPresent {
			// Not a master node, so let's see if the node needs labels
			if !workerNodeLabelPresent || !sdsWorkerNodeLabelPresent || workerNodeLabelValue != KubernetesWorkerLabelValue || sdsWorkerNodeLabelValue != sdsWorkerNodeLabelValue {
				node.Labels[KubernetesWorkerLabelName] = KubernetesWorkerLabelValue
				node.Labels[SDSWorkerLabelName] = SDSWorkerLabelValue
				_, err = c.client.CoreV1().Nodes().Update(node)
				if err != nil {
					logger.Errorf("Could not set labels on node -->%s<--, error was %s", node.Name, err)
					return fmt.Errorf("could not set labels on node -->%s<--, error was %s", node.Name, err)
				} else {
					logger.Infof("Set worker labels on node -->%s<--", node.Name)
				}
			} else {
				logger.Infof("worker node -->%s<-- does not need labeling", node.Name)
			}
		} else {
			logger.Infof("control plane node -->%s<-- does not need labeling", node.Name)
		}
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *NodeLabelBot5000Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		glog.Infof("Error syncing SDSApplication %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping krakenCluster %q out of the queue: %v", key, err)
}

func (c *NodeLabelBot5000Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	glog.Info("Starting NodeLabelBot5000 controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping SDSCluster controller")
}

func (c *NodeLabelBot5000Controller) runWorker() {
	for c.processNextItem() {
	}
}

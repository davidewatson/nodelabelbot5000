# nodelabelbot5000

`nodelabelbot5000` is deployed by [`cma-operator`][cma-operator] and adds
the following labels to non-control-plane nodes:

```
node-role.kubernetes.io/worker=true
kubernetes.io/nodetype=app
```

This mechanism is not configurable, i.e. all labels will be added to all
non-control-plane nodes. Therefore it has been deprecated. To set specific
labels for specific nodes please use the Kubernetes API directly.

[cma-operator]: https://github.com/samsung-cnct/cma-operator

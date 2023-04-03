# wildwest

Welcome to the Wild West!

## Run the program

### Build the app
The Helm charts use images hosted in GHCR containing the current version of the app.

If you'd like to build the container image yourself:
- Change the image names in `Makefile` and in `helm/values.yaml`
- Run `make`, which will build and push the images with Docker Buildx to your repository

### Install Helm chart
```
helm install wildwest helm/
```

### Check whether all pods are ready
```
kubectl get po --watch
```

Once the shootout is over, the pods will remain running.

### Check logs
```
kubectl logs --tail=-1 -l 'app in (cowboy, cowboy-controller)' --all-containers --ignore-errors | grep -v "DEBUG" | sort | less
```

### Uninstall Helm chart
```
helm uninstall wildwest
```

## Things learned
- If you don't know almost anything about distributed systems, and don't have enough time to read through
Martin Kleppmann's "Designing Data-Intensive Applications", at least watch some of his videos first and understand
that you will need some sort of consistent datastore, for example etcd, before trying to implement everything and
wasting almost the entire time trying to have only one cowboy remaining at the end!!! :)
- More about what Helm can be used for
- What "internal", "cmd", and "pkg" directories are actually used for
- How to write testable interfaces and mock tests
- gRPC in practice (never used it before, had only watched some tutorials before) - it's pretty cool!
- What etcd is, also that it does a small number of things, but does it very well
- It is worth investing early in things that take a long amount of time - installing a Kubernetes cluster on my laptop
instead of using the existing (slow) one on a Raspberry Pi, also, ensuring that the pods can start in parallel sped
everything up by a lot

## TODO
- Use secure gRPC connections
- Needs some more unit tests, also needs integration, end-to-end tests
- Needs a pipeline, of course
- Contexts could be used in more places
- Move out some configuration out of constants to a config file, environment variables, or a ConfigMap
- Add resource requests and limits to the containers to be able to utilize certain services for cost saving :)

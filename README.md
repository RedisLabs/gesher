# Gesher
K8s Admission control proxy.  
"Gesher" means bridge in Hebrew.

## Motivation
Kubernetes has [Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#what-are-admission-webhooks).  

These enable administrators of a Kubernetes cluster to include their own set of admission control over Kubernetes objects.  

The problem is that admission control as it works today, really just works at the cluster level.  Every webhook admission service will get admission requests for every object in every namespace by default.  In the world of namespaced operators this becomes a problem for 2 reasons.

1) as each operator might want to do admission control itself.  In fact, the operator-sdk encourages building the admission control webhook HTTPs servers as part of the operator binary.  This would result in many webhook servers getting sent many requests for objects not in their own namespace.  While, one can create webhook that select based on a namspace's labels, these can be fragile as they are dependent on the labels always being setup correctly

2) to create a webhook, one needs cluster level privileges, which would mean that a user in a namespace can use a validating webhook to intercept all any object in any namespace.  users within a namespace should be able to setup admission control juet for their namespace without any ability to impact other namespaces.

## Solution
Gesher is a cluster level admission proxy, that is the single point for the kubernetes api-server to issue admission requests.
In turn, Gesher proxies the request to the correct admission control https server in the correct namespace.

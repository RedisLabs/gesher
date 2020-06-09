# Gesher - A Webhook Proxy for K8s


# Motivation

Kubernetes has Admission Webhooks.  These enable administrators of a Kubernetes cluster to include their own set of admission control over Kubernetes objects.  The problem is that admission control as it works today, really just works at the cluster level.  I.e. every webhook admission service will get admission requests for every object in every namespace by default.  In the world of namespaced operators this becomes a problem, as each operator might want to do admission control itself.  In fact, the operator-sdk encourages building the admission control webhook HTTPs servers as part of the operator binary.  This would result in many webhook servers getting sent many requests for objects not in their own namespace.


# Possible Solutions



1. Just register each webhook service in each namespace and ignore (i.e. return allowed) requests for resources that are not in that namespace
    1. However, this can put a large load on kubernetes api-server as it will still have to send http requests to all namespaces.  
    2. If one namespace is broken and doesn’t return allowed on resources not in its namespace, admission will be denied erroneously.
2. Have a single cluster level admission control that knows how to directly interact with operators in individual namespaces.
    3. &lt;issue here>
3. Webhook’s have the ability to use a namespace label selector, i.e. they will only be used when a resource’s namespace’s labels match the webhook’s label selector
    4. This is fragile as it requires namespaces be setup correctly (i.e. with the proper set of labels)
4. Modify Kubernetes to be able to select on a namespace’s name, much like it can on labels.
    5. Can be problematic to get it included in kubernetes, and even if one can, doesn’t help with legacy kubernetes installations
5. Create a cluster level admission proxy that is the single point for the kubernetes api-server to issue admission requests which in turn knows how to proxy the request to the correct admission control https server in the correct namespace


# Design of the Proxy

The Proxy is a relatively simple design.  It’s configured to kubernetes as the webhoo validating service, the kubernetes api-server sends https requests to it, it determines if it should proxy the request to a namespaced admission server or just return the default answer (allowed or denied) if no namespaced admission server is configured 

It will have 2 custom resources, one a cluster level resource that is always needed to be configured by cluster administrator that define which resources it will proxy and a namespace resource that is defined by end users / deployers of an operator that instruct the proxy to proxy resources defined in it to the admission control service named within it. \
 \
The proxy will be stateful.  Namely it will keep an internal representation of these 2 resources within it and update its internal representation as the resources are created / modified / deleted.  The assumption is that on startup it will be asked to reconcile all existing resources if they already exist and hence reconstruct it’s internal representation.  If this assumption is false, will have to figure out another mechanism to keep its state between runs


## Custom Resources

As noted, the operator will make use of 2 Custom Resources (which can be extended to 4 if mutating admission controllers are added to the mix, but right now limited to just validating controllers).


### <span style="text-decoration:underline;">ValidatingWebhookProxyType</span>

This cluster level CustomResource defines the type of “rules” (resources/verbs) that can be proxied

The operator will aggregate all the existing custom resources on creation/deletion/update into a single ValidatingWebhookConfiguration resource and apply it to the cluster becoming the cluster’s endpoint for validating admission control for all the defined type.

Note: if a **NamespacedValidatingWebhook** resource is using defines a proxy for the resource/rule contained, it will be unable to be deleted until the **NamespacedValidatingWebhook** is deleted


### <span style="text-decoration:underline;">NamespacedValidatingWebhook</span>

This namespace level custom resource is an analogue to the current ValidatingWebhook resource and instructs the proxy to forward requests (that have to be already accepted by a ValidatingProxyType) to a service defined within it.

It’s primary difference from the normal ValidatingWebhook is that it doesn’t allow external http’s URLs to be used as an https server, but requires that it be a kubernetes service within the namespace.


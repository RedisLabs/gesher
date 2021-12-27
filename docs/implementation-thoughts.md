# Gesher/Proxy CRDs and lookups

This is an example of a Kubernetes Validating Webhook Configuration resource

```
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: redb-admission
webhooks:
- name: redb.admission.redislabs
  rules:
  - apiGroups:   ["app.redislabs.com"]
    apiVersions: ["v1alpha1"]
    operations:  ["*"]
    resources:   ["redisenterprisedatabases"]
  clientConfig:
    service: 
      namespace: automation-1
      name: admission
      path: /admission
    caBundle: &lt;cert>
```

We want to create 2 resources, one resource that is created by the cluster administrator that defines what “rules” are proxied and one resource that end users can configure to control validation within their own namespace.

**ProxyValidatingType**

It’s spec would be 

```
proxyTypes:
- name: string
  rules:
  - apiGroups:   []string
    apiVersions: []string
    operations:  []string
    resources:   []string
```

This would just be the apigroup/version/operation/resource above, and would provide the same list format.

It would be be stored in a multilevel map setup

I.e.

 `map[string=apiGroup]map[string=apiVersion]map[string=resource]map[string=operation]uid`

Where ‘*’ is a special case.

uuid is the uuid of the resource.  Enabling easy removal on updates / deletes

Every validation lookup would go like

```
func allowedToProxy(resourceObject) bool {
	apiVerMap, ok := map[resourceObject.apiGroup]
	If !ok {
		apiVerMap, ok = map["*"]
		if !ok {
			return false
		}
	}
	resourceMap, ok := apiVerMap[resourceObject.apiVersion]
	if !ok {
		resourceMap, ok = apiVerMap["*"]
		if !ok {
			return false
		}
	}
	opMap, ok := resourceMap[resourceObject.resource]
	if !ok {
		opMap, ok = resourceMap["*"]
		if !ok {
         return false
        }
	}


	return opMap[resourceObject.operation] != 0 || opMap["*"] != 0
}
```


It would create its own ValidatingWebhookConfiguration to have kubernetes api server point all these rules to it.

// need to figure out how to create function to emit the ValidatingWebhookConfiguration resource out of the data structure efficiently computationally and to kubernetes

**ProxyValidingWebhookConfiguration**

This would be essentially the same as the existing ValidatingWebhookConfiguration, but would only impact resources in the namespace that it belongs to.

It would be stored in the same mapping type structure as described above, but instead of ending in a uid, would have another mapping level for namespace which would point to a map of these resources by name (really a list, but a map enables easier updates) that 

So instead of just looking up a single map, at each level you would lookup both the value of the resourceObject and ‘*’ and stick them into a list (if existing) and at the next level iterator over every map in the list, doing the same thing until one reaches the namespace level which would return a map/list of all endpoints in the namespace for that resourceObject

Something like

```
func findAdmissionEndpoints(resourceObject) []Endpoint {
	verLst = []map[string]map[string]map[string]map[string][]Endpoint
	resList = []map[string]map[string]map[string][]Endpoint
	opList = []map[string]map[string][]Endpoint
	nsList = []map[string[]Endpoint
	endpointList = []Endpoint

	if apiVerMap, ok := map[resourceObject.apiGroup]; ok {
		verList = append(verList apiVerMap)
	}

	if apiVerMap, ok := map["*"]; ok {
		verList = append(verList apiVerMap)
	}

	for _, m := range verLst {
		if resMap, ok := m[resourceObject.apiVersion]; ok {
			resList = append(resList, resMap)
		}
		if resMap, ok := m["*"]; ok {
			resList = append(resList, resMap)
		}
	}

	for _, m := range resList {
		if opMap, ok := m[resourceObject.resource]; ok {
			opList = append(opList, opMap)
		}
		if opMap, ok := m["*"]; ok {
			opList = append(opList, opMap)
		}
	}

	for _, m := range opList { // perhaps combine with above level
		if nsMap, ok := m[resourceObject.operation]; ok {
			nsList = append(nsList, nsMap)
		}
	}

	for _, m := range nsList {
		endpoints = m[resourceObject.namespace]
		endpointList = append(endpointList, endpoints…)
	}

	return endpointList
}
```

# Crossplane Composition Function - Iterable

This Crossplane Composition Function is an experimental function to work around the lack of support for iterations
in standard Crossplane compositions.

The function was created to solve a very specific use case, but the end result is a generic solution that seems
applicable to numerous use cases - as it enables a "composition inside a composition", based on iterating a array-typed
field in the composite resource.

## Usage

The usage will be illustrated by an example - the original use case it was built to solve.

We wanted to create a composite resource representing a Kafka topic, containing the core topic configuration, but in
addition a configuration of the Kafka client principals allowed to produce/consume data to/from the topic.

Example XR:

```yaml
apiVersion: kafka.statnett.no/v1alpha1
kind: XTopic
metadata:
  name: my-topic
spec:
  topicName: my_topic
  producers:
    - principal: "User:producer-app"
  consumers:
    - principal: "User:foo-app"
    - principal: "User:bar-app"
```

With this function at hand, you can create a composition to create the managed resources (omitting irrelevant details
in this context):

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: topic.kafka.statnett.no
spec:
  compositeTypeRef:
    apiVersion: kafka.statnett.no/v1alpha1
    kind: XTopic
  resources:
    - name: topic
      base:
        apiVersion: topic.kafka.crossplane.io/v1alpha1
        kind: Topic
      patches:
        - fromFieldPath: spec.topicName
          toFieldPath: metadata.annotations[crossplane.io/external-name]
        # Write actual topic name to status to ensure ACLs are referring correct topic 
        - type: ToCompositeFieldPath
          fromFieldPath: metadata.annotations[crossplane.io/external-name]
          toFieldPath: status.topicName
  functions:
    - name: topic-producers
      config:
        apiVersion: xfn-iterable.crossplane.io/v1alpha1
        kind: Config
        metadata:
          name: topic-producers
        spec:
            # The field path to iterate in the composite resource
            fromFieldPath: spec.producers
            # Use a patch set to avoid duplication of patches
            patchSets:
              - name: resource-name-principal
                patches:
                  - fromFieldPath: status.topicName
                    toFieldPath: spec.forProvider.resourceName
                  # This is a new patch type; selecting a field from the nested iterable
                  - type: FromIterableFieldPath
                    fromFieldPath: principal
                    toFieldPath: spec.forProvider.resourcePrincipal
            resources:
              - name: write-acl # allow producer to write data
                base:
                  apiVersion: acl.kafka.crossplane.io/v1alpha1
                  kind: AccessControlList
                  spec:
                    forProvider:
                      resourceType: "Topic"
                      resourceHost: "*"
                      resourceOperation: "Write"
                      resourcePermissionType: "Allow"
                      resourcePatternTypeFilter: "Literal"
                patches:
                  - type: PatchSet
                    patchSetName: resource-name-principal
      type: Container
      container:
        image: crossplane/xfn-iterable:latest
    - name: topic-consumers
      config:
        apiVersion: xfn-iterable.crossplane.io/v1alpha1
        kind: Config
        metadata:
          name: topic-consumers
        spec:
            # The field path to iterate in the composite resource
            fromFieldPath: spec.consumers
            resources:
              - name: read-acl # allow producer to read data
                base:
                  apiVersion: acl.kafka.crossplane.io/v1alpha1
                  kind: AccessControlList
                  spec:
                    forProvider:
                      resourceType: "Topic"
                      resourceHost: "*"
                      resourceOperation: "Read"
                      resourcePermissionType: "Allow"
                      resourcePatternTypeFilter: "Literal"
                patches:
                  - fromFieldPath: status.topicName
                    toFieldPath: spec.forProvider.resourceName
                  # This is a new patch type; selecting a field from the nested iterable
                  - type: FromIterableFieldPath
                    fromFieldPath: principal
                    toFieldPath: spec.forProvider.resourcePrincipal
      type: Container
      container:
        image: crossplane/xfn-iterable:latest
```

This composition should create the wanted managed resources:

- The Topic
- One Write ACL per topic producer principal
- One Read ACL per topic consumer principal

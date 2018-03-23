# Subject Access Delegation

Subject Access Delegation is a tool used to automate the life cycle of RBAC
permissions in Kubernetes clusters. A controller listens to user written rules,
which once are met, will execute some defined RBAC event. This is achieved
through a controller listening to new rules, stored as objects in the API server
of a `subjectaccessdelegation` custom resource type.

Rules are broken into 3 main groups; **Origin Subject**, **Destination
Subjects** and **Triggers**. Once all the triggers have been met within the
rule, destination subjects will then take on the permissions of whatever the
origin subject holds by means of replicating appropriate Role Bindings and
Cluster Role Bindings.

## Origin Subject
All Subject Access Delegations need one and only one Origin Subject. This will
be the origin for all Role Bindings created onto the Destination Subjects. An
Origin Subject can be one of a:

* **Role**: Destination Subjects are simply bound to this role in the specified
  namespace.
* **Cluster Role**: Destination Subjects are simply bound to this role cluster
  wide.
* **Service Account**: Role Bindings and Cluster Role Bindings bound to this
  Service Account are replicated onto the Destination Subjects.
* **User**: Role Bindings and Cluster Role Bindings bound to this
  User are replicated onto the Destination Subjects.
* **Group**: Role Bindings and Cluster Role Bindings bound to this
  Group are replicated onto the Destination Subjects.

## Destination Subjects
All Subject Access Delegations need one or more of a Destination Subject. These
subjects will have the corresponding Role Bindings applied to them. A
destination subject can be one of a:

* **Service Account**
* **User**
* **Group**

## Triggers
All Subject Access Delegations need one or more triggers within their rule. Once
these triggers have been satisfied they will trigger the permissions to take
place. Triggers come as two different kinds:

* **Time**: A short hand or full time stamp string. Simply a time till the
  trigger will be satisfied.
* **Event**: Some event that needs to take place within the cluster for the
  trigger to be satisfied. For example, a pod being created or terminated.

## Deletion Triggers
Like the activation triggers, for a rule's permission to be removed the
deletion triggers must be satisfied. They also come in the form of:

* **Time**
* **Event**

## Metadata and Spec
The Subject Access Delegation will also take several other attributes:

* **Name**: Name of the Subject Access Delegation. Must be unique to the
  namespace.
* **Namespace**: Namespace the delegation is active in.
* **Repeat**: How many times the delegation should be repeated. Default value of
  one.

An example of a rule is as follows. Here `Remote-Employee1` and
`Remote-Employee2` will take on the permissions of the user `Employee` at 6:00pm
for 14 hours every day for 365 days.

```yaml
apiVersion: authz.k8s.io/v1alpha1
kind: SubjectAccessDelegation
metadata:
name: my-subject-access-delegation
namespace: dev-namespace
spec:
repeat: 365
originSubject:
  kind: User
  name: Employee
destinationSubjects:
- kind: User
  name: Remote-Employee1
- kind: User
  name: Remote-Employee2
triggers:
- kind: Time
  value: 6:00pm
deletionTriggers:
- kind: Time
  value: 14h
```

## Notes
- The controller also supports an adjusted internal time clock through use of
passed NTP server URLs.
- Permissions on destination subjects are dynamic in accordance to changes to
  the origin subject's within active rules.
- Event trigger names accepts wildcards (\*) that are resolved during runtime.
- The controller will attempt failure recovery if for whatever reason the
  controller exits or crashes. This means a state re-sync with all rules in the
  API server.

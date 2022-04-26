package replicaset

type ReplicaSetController struct {
}

func NewReplicaSetController() *ReplicaSetController {
	return &ReplicaSetController{}
}

// Run begins watching and syncing.
func (rsc *ReplicaSetController) Run() {

}

func (rsc *ReplicaSetController) syncReplicaSet() {

}

/*
addRS 在有 ReplicaSet 被创建（Informer 发现从前未出现过的 ReplicaSet）时被调用。
*/
func (rsc *ReplicaSetController) addRS(obj any) {

}

/* updateRS
可能在几种情况下被调用：
1. 某 ReplicaSet 被使用 Update 或 Patch 方法更改，触发 Update 事件；
2. 在 Periodic Resync 的过程中，所有 ReplicaSet 都会被触发 Update 事件；
3. 某个旧 ReplicaSet 被删除（删除之前可能被进行若干次修改）,
后一个新的有同样 Namespaced Name 的 ReplicaSet 被创建出来，如果删除事件被 Informer 错失的话，
它是无法区分新旧 ReplicaSet 的，因此它认为发生了一次 Update；
*/
func (rsc *ReplicaSetController) updateRS() {

}

/* deleteRS
触发有两种情况：
即 API Server 告知 Informer 有 Object 被删除，
或 Informer 自行产生的 DeletedFinalStateUnknown 。
*/
func (rsc *ReplicaSetController) deleteRS() {

}

func (rsc *ReplicaSetController) manageReplicas() {

}

func (rsc *ReplicaSetController) getPods() {}

func (rsc *ReplicaSetController) addPod() {

}

func (rsc *ReplicaSetController) updatePod() {

}

func (rsc *ReplicaSetController) deletePod() {

}

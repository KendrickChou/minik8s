Workflow of informer:

1. Informer 在初始化时，Reflector 会先 List API 获得所有的 Pod
2. Reflect 拿到全部 Pod 后，会将全部 Pod 放到 Store 中
3. 如果有人调用 Lister 的 List/Get 方法获取 Pod， 那么 Lister 会直接从 Store 中拿数据
4. Informer 初始化完成之后，Reflector 开始 Watch Pod，监听 Pod 相关 的所有事件;如果此时 pod_1 被删除，那么 Reflector 会监听到这个事件
5. Reflector 将 pod_1 被删除 的这个事件发送到 DeltaFIFO
6. DeltaFIFO 首先会将这个事件存储在自己的数据结构中(实际上是一个 queue)，然后会直接操作 Store 中的数据，删除 Store 中的 pod_1
7. DeltaFIFO 再 Pop 这个事件到 Controller 中
8. Controller 收到这个事件，会触发 Processor 的回调函数
9. LocalStore 会周期性地把所有的 Pod 信息重新放到 DeltaFIFO 中


example response(pod):
type T struct {
    Kind       string `json:"Kind"`
    APIVersion string `json:"APIVersion"`
    Name       string `json:"Name"`
    Namespace  string `json:"Namespace"`
    UID        string `json:"UID"`
    Spec       struct {
        Containers []struct {
            Name            string   `json:"Name"`
            Namespace       string   `json:"Namespace"`
            ID              string   `json:"ID"`
            Image           string   `json:"Image"`
            ImagePullPolicy string   `json:"ImagePullPolicy"`
            Entrypoint      []string `json:"Entrypoint"`
            Mounts          []struct {
                MountType   string `json:"MountType"`
                MountSource string `json:"MountSource"`
                MountTarget string `json:"MountTarget"`
            } `json:"Mounts,omitempty"`
        } `json:"Containers"`
    } `json:"Spec"`
    Status struct {
    } `json:"Status"`
}
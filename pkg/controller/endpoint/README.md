endpoint是k8s集群中的一个资源对象，存储在etcd中，用来记录一个service对应的所有pod的访问地址。

service配置selector，endpoint controller才会自动创建对应的endpoint对象；否则，不会生成endpoint对象.


endpoint controller是k8s集群控制器的其中一个组件，其功能如下：

1. 负责生成和维护所有endpoint对象的控制器

2. 负责监听service和对应pod的变化

3. 监听到service被删除，则删除和该service同名的endpoint对象

4. 监听到新的service被创建，则根据新建service信息获取相关pod列表，然后创建对应endpoint对象

5. 监听到service被更新，则根据更新后的service信息获取相关pod列表，然后更新对应endpoint对象

6. 监听到pod事件，则更新对应的service的endpoint对象，将podIp记录到endpoint中
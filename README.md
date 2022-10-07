### day01

在写LRU算法时，需要使用双向链表(container/list)，越靠近尾节点tail越“旧”，越靠近头节点head越“旧”。

同时也需要封装一个Node类型，包括key和value两个字段。双向链表中的节点就是我们自定义的Node。

最后还需要一个map，来实现时间复杂度为O(1)的get和put操作。之所以要自定义一个Node，保存k-v，是因为我们在链表中删除队首后，map表中也要删除，这时候根据key删除map表中的键值对，更加快速和简单。

Cache的增删改查：

* 增加：将key、value存入map中，同时将新节点加入到队尾（PushFront）。
* 删除（缓存淘汰）：获取队首后（Back），在链表中删除（Delete），同时在map中删除，如果回调函数不为空，执行回调函数。
* 更新：更新这条记录，并将该节点移动到队尾（MoveToFront）。
* 查找：如果要查询的节点存在（使用map查询），将节点返回，并移动到队尾（MoveToFront）。不存在的话，直接返回。

list双向链表定义：

* 对于双向链表来说，**队首和队尾是相对的。这里使用Front为队尾，Back为队头。**

哨兵节点：

* GO标准库实现的双向链表，没有像Java那样使用head、tail来分别表示头部和尾部。而是使用了一个虚拟节点Root实现了head、tail的作用。
* ![哨兵节点](https://richarli.oss-cn-beijing.aliyuncs.com/images/list哨兵节点.png)

参考资料：

[什么是鸭子类型](https://cloud.tencent.com/developer/article/1849579)  
[Go 标准库-双向链表 (container/list) 源码解析](https://blog.csdn.net/eight_eyes/article/details/121068799)  
[理解Golang中的interface和interface{}](https://www.cnblogs.com/maji233/p/11178413.html)            

### day02

GoCache代码结构雏形：

~~~go
gocache/
	|--lru/
		|--lru.go		// lru缓存淘汰策略
	|--byteview.go 	// 缓存值的抽象与封装
	|--cache.go		// 并发控制
	|--gocache.go		// 负责与外部交互，控制缓存存储和获取的主流程
~~~

### day03
~~~go
gocache/
	|--lru/
		|--lru.go		// lru缓存淘汰策略
	|--byteview.go 	// 缓存值的抽象与封装
	|--cache.go		// 并发控制
	|--gocache.go		// 负责与外部交互，控制缓存存储和获取的主流程
	|--http.go			// 提供被其他节点访问的能力（基于http）
~~~
### day04

参考资料：

[Redis中的数据分布算法](https://blog.csdn.net/m0_53474063/article/details/113381122)

### day05

~~~go
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
~~~

~~~go
使用一致性哈希选择节点        是                                    是
    |-----> 是否是远程节点 -----> HTTP 客户端访问远程节点 --> 成功？-----> 服务端返回返回值
                    |  否                                    ↓  否
                    |----------------------------> 回退到本地节点处理。
~~~

### 测试流程

~~~go
模拟本地数据库数据：
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
~~~

1. 运行run.sh。

   默认将apiServer和端口为8003的节点绑定在一起，所以会先检查这个服务器中有没有请求key的数据，如果没有，再去其他节点获取。

   ![image-20221007181159001](E:\develop\code\GOProjects\GoCache\pic\run.png)

2. 获取key为"Tom"的数据。

   默认节点没有key为Tom的数据，通过一致性哈希算法，pick出了端口为8001的节点。

   此时8001节点还没有缓存该条数据，所以调用回调函数，查询本地数据库，并将结果加入本机mainCache中。然后返回。

   ~~~shell
   curl "http://localhost:9999/api?key=Tom"
   ~~~

   ![image-20221007181529027](E:\develop\code\GOProjects\GoCache\pic\curl1.png)

3. 再次获取key为"Tom"的数据

   此时默认节点pick出8001节点后，8001节点命中缓存，直接返回数据。

   ![image-20221007181910823](E:\develop\code\GOProjects\GoCache\pic\curl2.png)

4. 查询数据库中不存在数据"Jocker"，返回不存在的响应。

   ![image-20221007182238959](E:\develop\code\GOProjects\GoCache\pic\cur3.png)
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
* ![](E:\develop\code\GOProjects\GoCache\GOCache\list哨兵节点.png)

参考资料：

[什么是鸭子类型]: https://cloud.tencent.com/developer/article/1849579
[Go 标准库-双向链表 (container/list) 源码解析]: https://blog.csdn.net/eight_eyes/article/details/121068799
[理解Golang中的interface和interface{}]:https://www.cnblogs.com/maji233/p/11178413.html


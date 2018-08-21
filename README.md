# 在任何地方访问在家中连接着路由器的linux主机
## 实现原理
本质就是转发消息：两个tcp连接，将一个tcp连接中读到的消息不经过任何处理直接写到另一个tcp连接。
通过ssh访问:使用ssh访问，最终所有的消息都转发到家中的ssh端口上
## 现实细节
需要借助一台公网的linux主机，例如腾讯云。
* 1、在带有公网ip的腾讯云上运行server.go监听连个端口A, B
* 2、让家中的liunx上运行client.go和server端口A建立tcp连接（管这个连接叫ctrl_connection） 这个连接只用来告知家中的主机，有新的连接请求来了
* 3、然后通过ssh user@server_ip -p B 连接server的B端口建立一个叫user_connect的连接。现在server知道有新的连接到了并通过ctrl_connection告知client来了新的ssh连接
* 4、client收到有新连接需要建立的通知后会建立两个连接，一个是在和server的A端口建立一个叫home_connect的连接, 一个和本地ssh的tcp连接。这两个连接会交换自己收到的数据（即转发）
* 5、在server上的连个连接user_connect和home_connect也交换自己的数据（即转发）

重复3、4、5可以用来处理多个ssh连接请求

整个流程：用户<---user_connect---->server <=交换数据=> server <----home_connect-----> client <=交换数据=> client <---ssh---> ssh
## 使用
#### 编译
```cgo
# 如果没有安装go
yum install golang
# 编译
go build server.go 
go build client.go 
```
#### 运行
* 公网主机（如腾讯云）
```shell
./server -user_port=12345 -home_port=12346 
```
* 家中
```shell
./client -ip=xxx.xxx.xxx.xxx -port=12346
```
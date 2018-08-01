# cvChain
## 阿里云全球区块链大赛-初赛

在自有系统中搭建Hyperledger Fabric开发测试环境。

推荐软件版本：Hyperledger Fabric v1.1，go version go1.9，Docker version 1.13.1  

Hyperledger Fabric环境准备及智能合约开发可参考：
http://hyperledger-fabric.readthedocs.io/en/release-1.1/chaincode4ade.html

我们只需编写chaincode，测试运行环境可以使用hyperleger frabric示例中的区块链配置。我们使用的时官方文档开发手册中的chaincode-docker-devmode来进行测试（位置：$GOPATH/src/github.com/hyperledger/fabric-samples/chaincode-docker-devmode）

将chaincode代码放入$GOPATH/src/github.com/hyperledger/fabric-samples/chaincode/路径下。按照文档中的步骤进行即可。

可能出现的报错：
在CLI（第三个终端）中无法找到我们引入的包

```
Error: Error getting chaincode code chaincode: Error getting chaincode package bytes: Error obtaining dependencies for github.com/hyperledger/fabric/bccsp: <go, [list -f {{ join .Deps "\n"}} github.com/hyperledger/fabric/bccsp]>: failed with error: "exit status 1"
can't load package: package github.com/hyperledger/fabric/bccsp: cannot find package "github.com/hyperledger/fabric/bccsp" in any of:
	/opt/go/src/github.com/hyperledger/fabric/bccsp (from $GOROOT)
	/opt/gopath/src/github.com/hyperledger/fabric/bccsp (from $GOPATH)
```
因为在容器中也需要下载fabric文件夹下的源码。可以将vendor打包的依赖文件导入，或者执行以下命令来安装需要的包：

```
$ go get -u --tags nopkcs11 github.com/hyperledger/fabric
```
安装需要一段时间，出现：

```
package github.com/hyperledger/fabric: no Go files in /opt/gopath/src/github.com/hyperledger/fabric
```
也没关系，已经安装成功。可以在CLI中安装chaincode并实例化以及测试：

```
//安装chaincode
$ peer chaincode install -p chaincodedev/chaincode/cv -n mycc -v 0
//实例化chaincode，这里可以不传参数
$ peer chaincode instantiate -n mycc -v 0 -c '{"Args":["1001","10","col1","master"]}' -C myc
//写入账本
$ peer chaincode invoke -n mycc -c '{"Args":["addRecord","1001","10","col1","master"]}' -C myc
//查询数据
$ peer chaincode query -n mycc -c '{"Args":["getRecord","1001","10"]}' -C myc
//写入加密数据
$ peer chaincode invoke -n mycc -c '{"Args":["encRecord","1009","2006","corp2", "engineer"]}' --transient  "{\"ENCKEY\":\"$ENCKEY\",\"IV\":\"$IV\"}" -C myc
//查询加密数据
$ peer chaincode query -n mycc -c '{"Args":["decRecord","1009","2006"]}' --transient  "{\"DECKEY\":\"$DECKEY\"}" -C myc
```

## 打包

首先需要安装govendor：

```
$ go get -u github.com/kardianos/govendor
```
安装完成后，进入chaincode源码所在的文件夹。

进行打包：

```
$ govendor init
$ govendor add +e
```
govendor会自动将chaincode依赖的包导入vendor文件夹中，其中的vendor.json记录了依赖包的关系。

可以通过

```
$ govendor list
```
查看导入的依赖包




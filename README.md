# yager

### golang installation

* download package from [here](https://golang.google.cn/dl/), go version **v1.13.4**
* extract it into /usr/local, creating a Go tree in /usr/local/go. 
For example:
> tar -C /usr/local -xzf go1.13.4.linux-amd64.tar.gz
* Add /usr/local/go/bin to the PATH environment variable. You can do this by adding this line to your `~/.bashrc`
> export PATH=$PATH:/usr/local/go/bin
> export GOPATH=$HOME/go
* Test your installation
> go version



### Packages manager govendor

* install govendor
> go get -u github.com/kardianos/govendor
* quick start
```
# Setup your project.
cd "my project in GOPATH"
govendor init

# Add existing GOPATH files to vendor.
govendor add +external

# View your work.
govendor list

# Look at what is using a package
govendor list -v fmt

# Specify a specific version or revision to fetch
govendor fetch golang.org/x/net/context@a4bbce9fcae005b22ae5443f6af064d80a6f5a55
govendor fetch golang.org/x/net/context@v1   # Get latest v1.*.* tag or branch.
govendor fetch golang.org/x/net/context@=v1  # Get the tag or branch named "v1".

# Update a package to latest, given any prior version constraint
govendor fetch golang.org/x/net/context

```
* more reference https://github.com/kardianos/govendor



### Swagger user guide

* install swagger
```
   $ mkdir -p $GOPATH/src/github.com/swaggo
   $ cd $GOPATH/src/github.com/swaggo
   $ git clone https://github.com/swaggo/swag
   $ cd swag/cmd/swag/
   $ go install -v
```
* download `gin-swagger`
```
   $ cd $GOPATH/src/github.com/swaggo
   $ git clone https://github.com/swaggo/gin-swagger
```
* swagger init
```
   $ cd xxx/pathto/yager/
   $ swag init
```

* example
```
   package user
   
   import (
       ...
   )
   
   
   // @Summary Add new user to the database
   // @Description Add a new user
   // @Tags user
   // @Accept  json
   // @Produce  json
   // @Param user body user.CreateRequest true "Create a new user"
   // @Success 200 {object} user.CreateResponse "{"code":0,"message":"OK","data":{"username":"admin"}}"
   // @Router /user [post]
   func Create(c *gin.Context) {
       ...
   }
```
>
>  Summary：简单阐述 API 的功能  
>   Description：API 详细描述  
>   Tags：API 所属分类  
>   Accept：API 接收参数的格式  
>   Produce：输出的数据格式，这里是 JSON 格式  
>   Param：参数，分为 6 个字段，其中第 6 个字段是可选的，各字段含义为： 
>  
>     1. 参数名称  
>     2. 参数在 HTTP 请求中的位置（body、path、query）  
>     3. 参数类型（string、int、bool 等）  
>     4. 是否必须（true、false）  
>     5. 参数描述  
>     6. 选项，这里用的是 default() 用来指定默认值  
>   Success：成功返回数据格式，分为 4 个字段  
>  
>     1. HTTP 返回 Code  
>     2. 返回数据类型  
>     3. 返回数据模型  
>     4. 说明  
>   路由格式，分为 2 个字段： 
>   
>     1. API 路径  
>     2. HTTP 方法  
>



### How to start

* git clone https://github.com/CadenOf/yager.git
* govendor init
* govendor add +e 
* govendor sync
* govendor list
* swag init
* make
* ./admin.sh start



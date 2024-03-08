# lxc-api

Web API for LXC.

# How to use

Place the server certificate in the/data/lxc/directory and name the private key and certificate lxc api.crt, lxc api-key, and lxc api-ca.crt, respectively.

For the client, you need to provide a p12 format client authentication certificate issued by lxc-api-ca.crt for Android.

Note: Starting from version 1.2, API port 8000 forces the use of HTTP bidirectional authentication(Mutual TLS authentication).

From the root source tree :

```
PATH=/your/lxc/bin:$PATH ./lxc-api
```

After that, API listen on port 8000, WEB terminal listen on port 8001, SSH server listen on port 8022.

# Documentation

1.获取服务器版本信息：<br>
请求：

```
curl -X GET http://localhost:8000/version
```
响应：

```
{
  "version": "LXC version 5.0.0"
}
```

2.获取当前所有容器的列表：<br>
请求：

```
curl -X GET http://localhost:8000/containers
```
响应：

```
{
  "containers": [
    "container1",
    "container2",
    "container3"
  ]
}
```

3.获取特定容器的信息：<br>
请求：

```
curl -X GET http://localhost:8000/container/<container1>
```
响应：

```
{
    "containers": [
        {
            "name": "ubuntu",
            "state": "RUNNING",
            "pid": "20058",
            "ip": "10.0.3.58",
            "cpu_usage": "2.61 seconds",
            "blkio_usage": "36.46 MiB",
            "memory_use": "66.71 MiB",
            "kmem_use": "10.28 MiB",
            "link": "vethXv9EeG",
            "link_state": {
                "tx_bytes": "3.04 KiB",
                "rx_bytes": "3.82 KiB",
                "total_bytes": "6.86 KiB"
            }
        }
    ]
}
```

4.启动容器：<br>
请求：

```
curl -X POST http://localhost:8000/container/<container1>/start
```
响应：

```
{
  "status_code": 200,
  "message": "Container started successfully"
}
```

5.停止容器：<br>
请求：

```
curl -X POST http://localhost:8000/container/<container1>/stop
```
响应：

```
{
  "status_code": 200,
  "message": "Container stopped successfully"
}
```

6.冻结容器：<br>
请求：

```
curl -X POST http://localhost:8000/container/<container1>/freeze
```
响应：

```
{
  "status_code": 200,
  "message": "Container frozen successfully"
}
```

7.解冻容器：<br>
请求：

```
curl -X POST http://localhost:8000/container/<container1>/unfreeze
```
响应：

```
{
  "status_code": 200,
  "message": "Container unfrozen successfully"
}
```

8.创建容器：<br>
请求：

```
curl -m 180 -X POST -H "Content-Type: application/json" -d '{
  "template": "download",
  "container_name": "my-container",
  "image_source": "mirrors.tuna.tsinghua.edu.cn/lxc-images",
  "distribution": "ubuntu",
  "release": "lunar",
  "architecture": "arm64"
}' http://localhost:8000/add/container
```
响应：

```
{
  "status_code": 200,
  "message": "Container created successfully"
}
```

9.删除容器：<br>
请求：

```
curl -X POST -H "Content-Type: application/json" -d '{
  "del_container": "my-container"
}' http://localhost:8000/del/container
```
响应：

```

{
  "status_code": 200,
  "message": "Container destroyed successfully"
}
```

# Thanks

- [NaisuXu / web-terminal-demo-with-golang-and-xterm](https://github.com/NaisuXu/web-terminal-demo-with-golang-and-xterm) : Web-terminal 实现

## Limiter

限流器数据储存在redis中，对于使用ip做key的限流器，程序内部ip的获取方式有两种：
1. 获取环境变量中的HOST_IP，对于容器运行的服务，在使用docker run的时候需要指定宿主IP为 HOST_IP
2. 获取程序出口IP，适用于非容器运行的服务
# TKMall
复刻抖音电商

# Deployment
```
# 本地安装工具链：redis mysql etcd protobuf等
# 如下操作
cd tools/etcd/ && bash install.sh

# docker 安装 kafka & zookper
cd config/env && docker-compose -f kafka-docker-compose.yaml up -d


chmod +x make.py

# 生成proto桩代码, 生成路径在build/proto_gen下
./make.py genproto

# 编译服务
./make.py build gateway
# 运行服务
./make.py run gateway
# 编译+运行
./make.py install gateway

# 一键全部编译和启动
./make.py install

# 单元测试
./make.py unitest
```

# Doc
https://bytedance.larkoffice.com/docx/VMgWdZBFsoSJyFxD5ByclZghn3g

# Backend official demo
https://github.com/cloudwego/biz-demo

# Frontend official demo
https://github.com/cloudwego/biz-demo/blob/main/gomall/app/frontend

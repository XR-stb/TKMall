启动报错：kafka server: In the middle of a leadership election, there is currently no leader for this partition and hence it is unavailable for writes


解决方案：
# 停止并删除Kafka和ZooKeeper容器
cd ~/code/go/TKMall/config/env
docker-compose -f kafka-docker-compose.yaml down

# 确保容器已经完全停止
docker ps | grep -E 'kafka|zookeeper'

# 清理卷和网络（可选，但有助于彻底重置）
docker-compose -f kafka-docker-compose.yaml down -v

# 重新启动容器
docker-compose -f kafka-docker-compose.yaml up -d

# 查看容器状态
docker ps

# 查看Kafka日志，确认启动正常
docker logs env-kafka-1 -f
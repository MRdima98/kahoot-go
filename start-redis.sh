redis-server --daemonize yes && sleep 1 
redis-cli < /dev.redis
redis-cli shutdown 
redis-server

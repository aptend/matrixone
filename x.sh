for num in {1..10000} 
do
    time mysql -h127.0.0.1 -P6001 -udump -p111 -e"select sum(k) from sbtest.sbtest1 where id > 15 and id < $RANDOM";
    sleep 1;
done

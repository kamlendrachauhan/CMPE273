CMPE 273 Project
Scalable cache using Consistent hashing on top of Redis server

The consistent hasing is implemented using three cache servers and one main db. The administrator can add and retrieve key-values in the database. The key-vales get sharded in the available caches based on consistent hashing technique. In case of adding a key-value, client performs the hashing and place the value on corressponsing cache. In case of retrival, client prforms hashing on key and contacts corressponding cache. If there is a miss, the cache contacts main database to retrieve the value. The administrator can also add and delete cache node to achieve load balance.

We have added GetBulk.go file which continuously gets the values and generates graph using kibana to show response time of retrival. It shows high response time value when is is retrieved from main database and takes less time when cache is used. 

The architecture diagram is shown below:


![Architecture Diagram](https://github.com/kamlendrachauhan/CMPE273/blob/master/ConsistentHashing.png)

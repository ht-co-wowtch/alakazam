# Kafka JMX

以dockerfile做範例如何啟動kafka jmx採樣kafka metrics

```dockerfile
kafka:
    image: wurstmeister/kafka:2.12-2.3.0
    environment:
      KAFKA_JMX_OPTS: -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.rmi.port=9091 -Djava.rmi.server.hostname=127.0.0.1
      KAFKA_OPTS: "-javaagent:/opt/kafka/jmx/jmx_prometheus_javaagent-0.12.0.jar=3036:/opt/kafka/jmx/config.yaml"
      JMX_PORT: 9091
    ports:
      - "9091:9091"
      - "3036:3036"
```

1. KAFKA_JMX_OPTS

   kafka本身有提供[KAFKA_JMX_OPTS](https://github.com/apache/kafka/blob/trunk/bin/kafka-run-class.sh#L178)來做啟動參數設定，所以只需要設定好KAFKA_JMX_OPTS即可

   | 參數                                       | 說明                                                         |
   | ------------------------------------------ | ------------------------------------------------------------ |
   | Dcom.sun.management.jmxremote              | 要啟動jmx                                                    |
   | Dcom.sun.management.jmxremote.authenticate | 不需要帳密認證                                               |
   | Dcom.sun.management.jmxremote.ssl          | 不需要ssl                                                    |
   | Dcom.sun.management.jmxremote.port         | metrics port，可以利用[JMX_PORT](https://github.com/apache/kafka/blob/trunk/bin/kafka-run-class.sh#L182) env做設定 |
   | Dcom.sun.management.jmxremote.rmi.port     | 對外port                                                     |
   | Djava.rmi.server.hostname                  | kafka所在的ip                                                |

   更多參數請[參考](https://docs.oracle.com/javase/9/management/monitoring-and-management-using-jmx-technology.htm#JSMGM-GUID-096EA656-4D07-4B09-A493-9EDEF83ABF28)

2. KAFKA_OPTS

   由於使用`prometheus`做採樣，但是jmx本身資料格式並不符合`prometheus`規範且jmx需透過java做輸出，於是使用`prometheus`提供[jmx_exporter](https://github.com/prometheus/jmx_exporter)做代理，所以需要以`java -javaagent:/path/to/JavaAgent.jar=[host:]<port>:<yaml configuration file> <kafka>`方式啟動，所以需借助[KAFKA_OPTS](https://github.com/apache/kafka/blob/trunk/bin/kafka-run-class.sh#L211)

3. port

| Port | 說明                                                 |
| ---- | ---------------------------------------------------- |
| 9091 | jmx port                                             |
| 3036 | 經由9091 port資料做轉成prometheus可以接受的資料 port |

4. 驗證

   jmx可以透過[jconsole](https://docs.oracle.com/javase/7/docs/technotes/guides/management/jconsole.html)觀看metrics，host是 127.0.0.1:9091 

  `prometheus`啟動完成後​http://127.0.0.1:3036 則可以看到`prometheus` metrics

5. [完整範例](https://gitlab.com/jetfueltw/cpw/cpwbox/blob/master/docker-compose.yml#L446)

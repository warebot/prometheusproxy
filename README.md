## This project is no longer maintained 

### Prometheus Proxy

The prometheusproxy proxies scraping requests initialized by the Prometheus collector based on configurable service names.  
It allows us to optionally add labels to the collected metrics requested by the Prometheus collector.  

#### Building the service  
The service builds to a single binary.  
To build the service simply execute **make**:  

    make

If the build is successful, a new binary named ** prometheusproxy **  should now be available in project's ** bin ** directory:

    bin
    └── prometheusproxy



#### Running the service
To run the service, simply run the binary:

        ./bin/prometheusproxy --config.file [path to config file]

If you do not specify the config.file argument, the binary will attempt to read the default ** promproxy.yml ** in the current directory.


A config file example:

        port: 9191
        services:
            service-a:
                endpoint: http://localhost:9100/metrics
                labels:
       	            user: mina
                    source: proxy
            service-b:
                endpoint: http://localhost:9100/metrics
                labels:
                    user: chuck
                    source: proxy


### Usage
#### /


      Returns the build information of the service  


** Example  **

Request

    HTTP GET /     

Response

    {
      branch: "master",
      buildDate: "20151216-14:01:45",
      buildUser: "@outbrain-nix",
      goVersion: "1.5.1",
      revision: "b758f28",
      version: "0.1"
    }

#### /metrics
Accept Headers: application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7, text/plain  

      Endpoint responsible for scraping the configured services



###### Query Parameters  


| parameter | required |    description      |  e.g |
|-----------|----------|:--------------------|-----|
| service | yes |service name for the configured endpoint in the config file | service-a |
| labels |  no  |delimited k/value label-set for ad-hoc application   |   user&#124;mina,manager&#124;gil |

*labels BNF:*

      <label> ::= <key"|"value>
    <labels> ::= <label>{"," <label> }


** Example **   

Request  

    HTTP GET /metrics/service=service-a


Response

    # HELP go_gc_duration_seconds A summary of the GC invocation durations.
    # TYPE go_gc_duration_seconds summary
    go_gc_duration_seconds{quantile="0", user="mina", source="proxy"} 0.000223927
    go_gc_duration_seconds{quantile="0.25", user="mina", source="proxy"} 0.00033225800000000004
    go_gc_duration_seconds{quantile="0.5", user="mina", source="proxy"} 0.00043606200000000003
    go_gc_duration_seconds{quantile="0.75", user="mina", source="proxy"} 0.000620278
    go_gc_duration_seconds{quantile="1", user="mina", source="proxy"} 0.007438669
    go_gc_duration_seconds_sum 0.06762654900000001
    go_gc_duration_seconds_count 109


Request  

    HTTP GET /metrics/service=service-a&labels=user|mina


Response

    # HELP go_gc_duration_seconds A summary of the GC invocation durations.
    # TYPE go_gc_duration_seconds summary
    go_gc_duration_seconds{quantile="0",user="mina"} 0.000223927
    go_gc_duration_seconds{quantile="0.25",user="mina"} 0.00033225800000000004
    go_gc_duration_seconds{quantile="0.5",user="mina"} 0.00043606200000000003
    go_gc_duration_seconds{quantile="0.75",user="mina"} 0.000620278
    go_gc_duration_seconds{quantile="1",user="mina"} 0.007438669
    go_gc_duration_seconds_sum 0.06762654900000001
    go_gc_duration_seconds_count 109

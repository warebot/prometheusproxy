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

Using the example config above, the following request:

    GET /metrics?service=service-a"

will fetch the metrics from the endpoint defined by ** service-a ** (http://localhost:9100/metrics),   
and apply the labels (user="mina", source="proxy") to each metric in the respons

# telepoller
snmp bulk polling, using influxdata/telegraf as a baseline

* git clone 
* build telepoller_X binary by cd'ing to folder
* copy telepoller_X.conf.dist to telepoller_X.conf
* customize as necessary
* configure telegraf to listen to telepoller_X.response topics
* run daemon
* send it job requests over nats topic telepoller_X.request somehow, scripting example:
  ```bash
  nats-pub telepoller_snmp.request '{"hosts":{"awesome.router":"192.0.2.0"},"params":{"community":"public","table":"ifMIB"}}'
  ```
# telepoller
snmp bulk polling, using influxdata/telegraf as a baseline

## vague instructions
* git clone 
* build telepoller_X binary by cd'ing to folder and running 'go build'
* copy telepoller_X.conf.dist to telepoller_X.conf
* customize as necessary
* configure telegraf to listen to telepoller_X.result topics
* run daemon
* send it job requests over nats topic telepoller_X.request somehow, scripting example:
  ```bash
  nats-pub telepoller_snmp.request '{"hosts":{"awesome.router":"192.0.2.0"},"params":{"community":"public","table":"ifMIB"}}'
  ```

## requirements
* a nats server (probably gnatsd)
* some snmp devices
* something to send poll requests (a shell script will do, but it's a naivë implementation)
* InfluxDB with Uint64 support enabled
* Telegraf with Uint64 support (I should probably send this over to Influx, but, test cases).  In the meantime grab it [here](https://github.com/ragzilla/telegraf/tree/metrics-rewrite).

## legal
* This software is licensed under the MIT license
* Portions of this software are Copyright (c) 2015 InfluxDB
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
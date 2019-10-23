#MySQL audit

This parser wil parse logfiles from MySQL audit saved as JSON
## JSON Example
an audit record would look like this:
```json
{"audit_record": {
    "name": "Query",
    "record": "1073_1970-01-01T00:00:00",
    "timestamp": "2019-09-09T09:09:09 UTC",
    "command_class": "select",
    "connection_id": "37",
    "status": 0,
    "sqltext": "SELECT * from table where answer = 42",
    "user": user[user] @  [127.0.0.1]",
    "host": "",
    "os_user": "",
    "ip": "127.0.0.1",
    "db": "mydatabase"
  }
}
```
This can be imported to clickhouse with this command.
```bash
clicktail --dataset='clicktail.mysql_audit_log' --parser=mysqlaudit --file=//path/to/logfile.log --backfill --config
```
## Syslog example
If you output audit to syslog. (either a local file or collecting audit log on a syslog server) your log will further have a timestamp, a syslog hostname and a syslog identifier.
The same record will look like this.
```json
Sep 09 09:09:09 syslog-hostname syslog-identifier: {"audit_record":{"name":"Query","record":"1073_1970-01-01T00:00:00","timestamp":"2019-09-09T09:09:09 UTC","command_class":"select","connection_id":"37","status":0,"sqltext":"SELECT * from table where answer = 42","user":user[user]@[127.0.0.1]","host": "","os_user": "", "ip": "127.0.0.1","db": "mydatabase"} }
```

```bash
clicktail --dataset='clicktail.mysql_audit_log' --parser=mysqlaudit --file=/root/percona.log --backfill  --debug --mysqlaudit.filter_regex='audit' --config=/etc/clicktail/clicktail.conf --mysqlaudit.syslog_ident=syslog-identifier:
```

Now you're able to parse and import audits from different servers into clickhouse. This can be useful if you need to analyze patterns across muliple servers, or if you have audit logs from multiple replication slaves.


Please note the timestamp is stripped away by using the PrefixRegex parameter. So you'll have to have a RegEXP defined in ```/etc/clicktail/clicktail.conf``` matching the timestamp format used by your syslog service.
If it is not defined, syslog will not be handled correct
Examples:

```conf
PrefixRegex = ^[A-z]{1,3} [0-9]{1,2} [0-9]{1,2}:[0-9]{1,2}:[0-9]{1,2}\s #Sep 09 09:09:09
PrefixRegex = ^(-?(?:[1-9][0-9]*)?[0-9]{4})-(1[0-2]|0[1-9])-(3[01]|0[1-9]|[12][0-9])T(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])(\\.[0-9]+)?(Z)?$ #ISO8601 2019-09-09T09:09:09Z
```


## Extra Syslog Example

In the file ```/var/log/percona_audit.log```
```json
  Sep 08 12:08:52 pxc-1 percona_audit: {"audit_record":{"name":"Query","record":"1073_1970-01-01T00:00:00","timestamp":"2019-09-08T09:08:52 UTC","command_class":"select","connection_id":"37","status":0,"sqltext":"SELECT `id`, `field`, `value` FROM `blog`.`posts`","user":"web[web] @  [127.0.0.1]","host":"","os_user":"","ip":"127.0.0.1","db":"blog"}}
 Sep 08 12:09:52 pxc-2 percona_audit: {"audit_record":{"name":"Query","record":"1073_1970-01-01T00:00:00","timestamp":"2019-09-08T09:09:52 UTC","command_class":"select","connection_id":"37","status":0,"sqltext":"SELECT `id`, `field`, `value` FROM `blog`.`posts`","user":"web[web] @  [127.0.0.1]","host":"","os_user":"","ip":"127.0.0.1","db":"blog"}}
```
This file is imported into clickhouse with this command:
```bash
[root@localhost]# clicktail --dataset='clicktail.mysql_audit_log' --parser=mysqlaudit --file=/var/log/percona_audit.log --mysqlaudit.filter_regex='percona_audit' --config=/etc/clicktail/clicktail.conf --mysqlaudit.syslog_ident=percona_audit: --backfill
```
Breakdown of this command
* `--dataset='clicktail.mysql_audit_log'` which dataset to store data in clickhouse
* `--parser=mysqlaudit` which parser to use
* `--file=/var/log/percona_audit.log` which audit logfile to parse
* `--mysqlaudit.filter_regex='percona_audit'` looking for this pattern before parsing excluding log entries not containing this pattern
* `--config=/etc/clicktail/clicktail.conf` This file contain other parameters for clickhouse to work.
* `--mysqlaudit.syslog_ident=percona_audit:` The syslog identifier. This is defined on your MySQL server in the variable `audit_log_syslog_ident`
* `--backfill` Retroactive logs loading



Now the audit records are imported into ClickHouse and you're able to work with it
```SQL
[root@localhost]# clickhouse-client --multiline

SELECT *
FROM clicktail.mysql_audit_log

Row 1:
──────
_time:            2019-09-08 09:08:52
_date:            2019-09-08
_ms:              0
command_class:    select
connection_id:    37
db:               blog
host:             
ip:               127.0.0.1
name:             Query
os_user:          
os_login:         
os_version:       
mysql_version:    
priv_user:        
proxy_user:       
record:           1073_1970-01-01T00:00:00
sqltext:          SELECT `id`, `field`, `value` FROM `blog`.`posts`
status:           0
user:             web[web] @  [127.0.0.1]
startup_optionsi: 
dbserver:         pxc-1 

Row 2:
──────
_time:            2019-09-08 09:09:52
_date:            2019-09-08
_ms:              0
command_class:    select
connection_id:    37
db:               blog
host:             
ip:               127.0.0.1
name:             Query
os_user:          
os_login:         
os_version:       
mysql_version:    
priv_user:        
proxy_user:       
record:           1073_1970-01-01T00:00:00
sqltext:          SELECT `id`, `field`, `value` FROM `blog`.`posts`
status:           0
user:             web[web] @  [127.0.0.1]
startup_optionsi: 
dbserver:         pxc-2 

2 rows in set. Elapsed: 0.003 sec. 

localhost :)
```


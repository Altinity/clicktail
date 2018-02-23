CREATE TABLE IF NOT EXISTS clicktail.mysql_audit_log
(
    `_time` DateTime,
    `_date` Date default toDate(`_time`),
    `_ms` UInt32,

    command_class String,
    connection_id UInt32,
    db String,
    host String,
    ip String,
    name String,
    os_user String,
    os_login String,
    os_version String,
    mysql_version String,
    priv_user String,
    proxy_user String,
    record String,
    sqltext String,
    status UInt32,
    user String,
    startup_optionsi String

) ENGINE = MergeTree(`_date`, (`_time`, host, user), 8192);
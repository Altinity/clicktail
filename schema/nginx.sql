CREATE TABLE IF NOT EXISTS clicktail.nginx_log
(
    `_time` DateTime,
    `_date` Date default toDate(`_time`),
    `_ms` UInt32,

    body_bytes_sent UInt32,
    http_user_agent String,
    http_referer String,
    http_bost String,
    remote_addr String,
    request String,
    request_method String,
    request_path String,
    request_pathshape String,
    request_protocol_version String,
    request_shape String,
    request_uri String,
    request_query String,
    request_queryshape String,
    remote_user String,
    status UInt32,
    strExtra1 String,
    strExtra2 String,
    strExtra3 String,
    intExtra1 UInt32,
    intExtra2 UInt32,
    intExtra3 UInt32,
    decExtra1 Float32,
    decExtra2 Float32,
    decExtra3 Float32

) ENGINE = MergeTree(`_date`, (`_time`, request_method, status), 8192)
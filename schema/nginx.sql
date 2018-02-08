CREATE TABLE IF NOT EXISTS clicktail.nginx_log
(
    `_time` DateTime,
    `_date` Date default toDate(`_time`),

    body_bytes_sent UInt32,
    http_user_agent String,
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
    status UInt32

) ENGINE = MergeTree(`_date`, (`_time`, request_method, status), 8192)
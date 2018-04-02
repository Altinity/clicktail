CREATE TABLE IF NOT EXISTS clicktail.mysql_slow_log
(
    `_time` DateTime,
    `_date` Date default toDate(`_time`),
    `_ms` UInt32,

    client String,
    query String,
    normalized_query String,
    query_time Float32,
    user String,
    statement String,
    tables String,
    schema String,
    rows_examined UInt32,
    rows_sent UInt32,
    lock_time Float32,
    connection_id UInt32,

    error_num UInt32,
    killed UInt16,

    rows_affected UInt32,
    database String,
    comments String,

    bytes_sent UInt32,
    tmp_tables UInt8,
    tmp_disk_tables UInt8,
    tmp_table_sizes UInt32,
    transaction_id String,
    query_cache_hit UInt8,
    full_scan UInt8,
    full_join UInt8,
    tmp_table UInt8,
    tmp_table_on_disk UInt8,
    filesort UInt8,
    filesort_on_disk UInt8,
    merge_passes UInt32,
    IO_r_ops UInt32,
    IO_r_bytes UInt32,
    IO_r_wait_sec Float32,
    rec_lock_wait_sec Float32,
    queue_wait_sec Float32,
    pages_distinct UInt32,

    sl_rate_type String,
    sl_rate_limit UInt16,

    hosted_on String,
    read_only UInt8,
    replica_lag UInt64,
    role String
    
) ENGINE = MergeTree(`_date`, (`_time`, query), 8192);
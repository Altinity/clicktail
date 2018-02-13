CREATE TABLE IF NOT EXISTS clicktail.mysql_slow_log
(
    `_time` DateTime,
    `_date` Date default toDate(`_time`),

    client String,
    query String,
    normalized_query String,
    query_time Float32,
    user String,
    statement String,
    tables String,
    rows_examined UInt32,
    rows_sent UInt32,
    lock_time Float32,
    
    
    rows_affected UInt32,
    database String,
    comments String,

    bytes_sent UInt32,
    tmp_tables UInt8,
    tmp_disk_tables UInt8,
    tmp_table_sizes UInt32,
    transaction_id UInt32,
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

    hosted_on String,
    read_only String,
    replica_lag String,
    role String
    
) ENGINE = MergeTree(`_date`, (`_time`, query_time, lock_time), 8192)
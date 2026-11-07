SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

SELECT
    GETUTCDATE() AS sample_time,
    r.session_id,
    DB_NAME(r.database_id) AS database_name,
    s.login_name,
    s.host_name,
    CASE
        WHEN s.program_name LIKE '.Net SqlClient%' THEN '.NET Application'
        WHEN s.program_name LIKE 'EntityFramework%' THEN 'Entity Framework'
        WHEN s.program_name LIKE 'Microsoft SQL Server Management%' THEN 'SSMS'
        WHEN s.program_name LIKE 'SQLAgent%' THEN 'SQL Agent'
        WHEN s.program_name LIKE 'Report Server%' THEN 'SSRS'
        WHEN s.program_name LIKE 'Microsoft Office%' THEN 'Excel/Office'
        WHEN s.program_name IS NULL OR s.program_name = '' THEN 'Unknown'
        ELSE LEFT(s.program_name, 40)
    END AS program_name,
    r.command,
    r.status AS request_status,
    r.cpu_time AS cpu_time_ms,
    r.logical_reads,
    r.writes AS physical_writes,
    r.total_elapsed_time AS duration_ms,
    COALESCE(
        wt.wait_type,
        r.wait_type,
        CASE WHEN r.status IN ('running', 'runnable') THEN 'CPU_RUNNING' ELSE 'Other' END
    ) AS effective_wait_type,
    COALESCE(wt.wait_duration_ms, r.wait_time, 0) AS wait_duration_ms,
    CASE
        WHEN r.status IN ('running', 'runnable')
          OR COALESCE(wt.wait_type, r.wait_type) IN ('SOS_SCHEDULER_YIELD', 'THREADPOOL')
            THEN 'CPU'
        WHEN COALESCE(wt.wait_type, r.wait_type) LIKE 'LCK_M_%'
          OR COALESCE(wt.wait_type, r.wait_type) LIKE 'LATCH_%'
          OR COALESCE(wt.wait_type, r.wait_type) LIKE 'PAGELATCH_%'
            THEN 'Lock'
        WHEN COALESCE(wt.wait_type, r.wait_type) LIKE 'PAGEIOLATCH_%'
          OR COALESCE(wt.wait_type, r.wait_type) IN (
                'WRITELOG', 'ASYNC_IO_COMPLETION', 'IO_COMPLETION',
                'BACKUPIO', 'WRITE_COMPLETION', 'LOGBUFFER'
            )
            THEN 'IO'
        WHEN COALESCE(wt.wait_type, r.wait_type) IN (
                'RESOURCE_SEMAPHORE', 'RESOURCE_SEMAPHORE_QUERY_COMPILE',
                'CMEMTHREAD', 'SOS_VIRTUALMEMORY_LOW'
            )
            THEN 'Memory'
        WHEN COALESCE(wt.wait_type, r.wait_type) IN ('ASYNC_NETWORK_IO', 'NET_WAITFOR_PACKET')
            THEN 'Network'
        ELSE 'Other'
    END AS wait_category,
    COALESCE(NULLIF(wt.blocking_session_id, 0), NULLIF(r.blocking_session_id, 0), 0)
        AS blocking_session_id,
    r.wait_resource,
    -- Top SQL por texto: sql_handle como hex, para calcular statement_key
    -- en el sampler (FNV-1a 64-bit). Nunca usar el texto SQL como etiqueta
    -- Prometheus; el texto se guarda aparte en centinela_pi.sql_text.
    CONVERT(VARCHAR(128), r.sql_handle, 2) AS sql_handle_hex,
    r.statement_start_offset,
    r.statement_end_offset
FROM sys.dm_exec_requests AS r WITH (NOLOCK)
INNER JOIN sys.dm_exec_sessions AS s WITH (NOLOCK)
    ON r.session_id = s.session_id
LEFT JOIN sys.dm_os_waiting_tasks AS wt WITH (NOLOCK)
    ON r.session_id = wt.session_id
    AND wt.exec_context_id = 0
WHERE s.is_user_process = 1
    AND r.session_id <> @@SPID
    AND COALESCE(wt.wait_type, r.wait_type, '') NOT IN (
        'BROKER_EVENTHANDLER', 'BROKER_RECEIVE_WAITFOR', 'BROKER_TASK_STOP',
        'BROKER_TO_FLUSH', 'BROKER_TRANSMITTER', 'CHECKPOINT_QUEUE', 'CHKPT',
        'CLR_AUTO_EVENT', 'CLR_MANUAL_EVENT', 'CLR_SEMAPHORE',
        'DBMIRROR_DBM_EVENT', 'DBMIRROR_EVENTS_QUEUE', 'DBMIRROR_WORKER_QUEUE',
        'DBMIRRORING_CMD', 'DIRTY_PAGE_POLL', 'DISPATCHER_QUEUE_SEMAPHORE',
        'EXECSYNC', 'FSAGENT', 'FT_IFTS_SCHEDULER_IDLE_WAIT', 'FT_IFTSHC_MUTEX',
        'HADR_CLUSAPI_CALL', 'HADR_FILESTREAM_IOMGR_IOCOMPLETION',
        'HADR_LOGCAPTURE_WAIT', 'HADR_NOTIFICATION_DEQUEUE', 'HADR_TIMER_TASK',
        'HADR_WORK_QUEUE', 'KSOURCE_WAKEUP', 'LAZYWRITER_SLEEP', 'LOGMGR_QUEUE',
        'MEMORY_ALLOCATION_EXT', 'ONDEMAND_TASK_QUEUE',
        'PARALLEL_REDO_DRAIN_WORKER', 'PARALLEL_REDO_LOG_CACHE',
        'PARALLEL_REDO_TRAN_LIST', 'PARALLEL_REDO_WORKER_SYNC',
        'PARALLEL_REDO_WORKER_WAIT_WORK', 'PREEMPTIVE_XE_GETTARGETSTATE',
        'PWAIT_ALL_COMPONENTS_INITIALIZED', 'PWAIT_DIRECTLOGCONSUMER_GETNEXT',
        'QDS_CLEANUP_STALE_QUERIES_TASK_MAIN_LOOP_SLEEP',
        'QDS_PERSIST_TASK_MAIN_LOOP_SLEEP', 'QDS_ASYNC_QUEUE',
        'REQUEST_FOR_DEADLOCK_SEARCH', 'RESOURCE_QUEUE', 'SERVER_IDLE_CHECK',
        'SLEEP_BPOOL_FLUSH', 'SLEEP_DBSTARTUP', 'SLEEP_DCOMSTARTUP',
        'SLEEP_MASTERDBREADY', 'SLEEP_MASTERMDREADY', 'SLEEP_MASTERUPGRADED',
        'SLEEP_MSDBSTARTUP', 'SLEEP_SYSTEMTASK', 'SLEEP_TASK',
        'SLEEP_TEMPDBSTARTUP', 'SNI_HTTP_ACCEPT', 'SOS_WORK_DISPATCHER',
        'SP_SERVER_DIAGNOSTICS_SLEEP', 'SQLTRACE_BUFFER_FLUSH',
        'SQLTRACE_INCREMENTAL_FLUSH_SLEEP', 'SQLTRACE_WAIT_ENTRIES',
        'WAIT_FOR_RESULTS', 'WAITFOR', 'WAITFOR_TASKSHUTDOWN',
        'WAIT_XTP_CKPT_CLOSE', 'WAIT_XTP_HOST_WAIT',
        'WAIT_XTP_OFFLINE_CKPT_NEW_LOG', 'WAIT_XTP_RECOVERY',
        'XE_BUFFERMGR_ALLPROCESSED_EVENT', 'XE_DISPATCHER_JOIN',
        'XE_DISPATCHER_WAIT', 'XE_LIVE_TARGET_TVF', 'XE_TIMER_EVENT'
    );
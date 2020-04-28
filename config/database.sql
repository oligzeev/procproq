-- Process
DROP TABLE IF EXISTS pp_process;
CREATE TABLE IF NOT EXISTS pp_process
(
    process_id uuid NOT NULL,
    name varchar(255) NOT NULL,
    CONSTRAINT pp_process_pkey PRIMARY KEY (process_id)
);

-- Task
DROP TABLE IF EXISTS pp_task;
CREATE TABLE IF NOT EXISTS pp_task
(
    process_id uuid NOT NULL,
    task_id uuid NOT NULL,
    name varchar(255) NOT NULL,
    category integer NOT NULL,
    action varchar(255) NOT NULL,
    read_mapping_id uuid NOT NULL,
    CONSTRAINT pp_task_pkey PRIMARY KEY (task_id)
);
CREATE INDEX IF NOT EXISTS pp_task_1 ON pp_task(process_id);

-- Task relation
DROP TABLE IF EXISTS pp_task_rel;
CREATE TABLE IF NOT EXISTS pp_task_rel
(
    process_id uuid NOT NULL,
    parent_id uuid NOT NULL,
    child_id uuid NOT NULL,
    CONSTRAINT pp_task_rel_pkey PRIMARY KEY (parent_id, child_id)
);
CREATE INDEX IF NOT EXISTS pp_task_rel_1 ON pp_task_rel(process_id);

-- Read mapping
DROP TABLE IF EXISTS pp_read_mapping;
CREATE TABLE IF NOT EXISTS pp_read_mapping
(
    read_mapping_id uuid NOT NULL,
    body jsonb,
    CONSTRAINT pp_read_mapping_pkey PRIMARY KEY (read_mapping_id)
);

-- Order
DROP TABLE IF EXISTS pp_order;
CREATE TABLE IF NOT EXISTS pp_order
(
    order_id uuid NOT NULL,
    process_id uuid NOT NULL,
    body jsonb,
    CONSTRAINT pp_order_pkey PRIMARY KEY (order_id)
);

-- Job
DROP TABLE IF EXISTS pp_job;
CREATE TABLE IF NOT EXISTS pp_job
(
    process_id uuid NOT NULL,
    task_id uuid NOT NULL,
    category integer NOT NULL,
    action varchar(255) NOT NULL,
    order_id uuid NOT NULL,
    read_mapping_id uuid NOT NULL,
    started boolean NOT NULL,
    completed boolean NOT NULL,
    ready_num integer NOT NULL,
    ready_req integer NOT NULL,
    trace varchar(510) NOT NULL,
    CONSTRAINT pp_job_pkey PRIMARY KEY (task_id, order_id)
);
CREATE INDEX IF NOT EXISTS pp_job_1 ON pp_job(task_id, order_id, completed);
CREATE INDEX IF NOT EXISTS pp_job_2 ON pp_job(ready_num, ready_req, started, task_id, order_id); -- ???
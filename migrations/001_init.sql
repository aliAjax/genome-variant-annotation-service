CREATE TABLE IF NOT EXISTS reference_datasets(id text primary key, name text not null, build text not null, version bigint not null, status text not null);
CREATE TABLE IF NOT EXISTS annotation_jobs(id text primary key, dataset_id text not null, status text not null, created_at timestamptz not null);
CREATE TABLE IF NOT EXISTS annotation_results(id text primary key, job_id text not null, variant_key text not null, payload jsonb not null);
